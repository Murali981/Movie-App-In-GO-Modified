package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Auth struct {
	Issuer      string
	Audience    string
	Secret      string
	TokenExpiry time.Duration
	RefreshExpiry time.Duration
	CookieDomain string
	CookiePath string
	CookieName string
}


type jwtUser struct {
	ID int 	`json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
}



type TokenPairs struct {
	Token string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}


type Claims struct {
	jwt.RegisteredClaims
}



 func (j *Auth)  GenerateTokenPair(user *jwtUser) (TokenPairs , error) {
	//// Create a token
   token := jwt.New(jwt.SigningMethodHS256)

	//// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s" , user.FirstName , user.LastName)
	claims["sub"] =  fmt.Sprint(user.ID)
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"

	/// In the above step i  have setup all the claims for the token . It means that all the above claims will be
	/// stored in the token when it is generated.....


	//// Set the expiry for JWT
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()


	//// Create a signed token
	signedAccessToken , err := token.SignedString([]byte(j.Secret))

	if err != nil {
		return TokenPairs{} , err
	}


	///// Create a refresh token and set claims...
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()


	//// Set the expiry for the refresh token....
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()


	///// Create a signed refresh token.....
	signedRefreshToken , err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{} , err
	}


	///// Create tokenpairs and populate with signed tokens
	var tokenPairs = TokenPairs {
		Token : signedAccessToken,
		RefreshToken: signedRefreshToken,
	}


	//// Return the token pairs.....
	return tokenPairs , nil
 }



  func (j *Auth)  GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name : j.CookieName,
		Path : j.CookiePath,
		Value : refreshToken,
		Expires : time.Now().Add(j.RefreshExpiry),
		MaxAge : int(j.RefreshExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode, // Make this cookie only limited to this site....
		Domain : j.CookieDomain,
		HttpOnly: true,
		Secure : true,
	}
  }



  func (j *Auth)  GetExpiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name : j.CookieName,
		Path : j.CookiePath,
		Value : "",
		Expires : time.Unix(0,0),
		MaxAge : -1,
		SameSite: http.SameSiteStrictMode, // Make this cookie only limited to this site....
		Domain : j.CookieDomain,
		HttpOnly: true,
		Secure : true,
	}
  }


  func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter , r *http.Request) (string , *Claims , error) {
	
	w.Header().Add("Vary" , "Authorization") /// We are adding a header on this line....


	/// Get the auth header.....
	authHeader := r.Header.Get("Authorization") /// In this line we are checking the existence of the authorization header
	// and read it's value into the variable authHeader...

	/// Sanity check
	if authHeader == "" { // In this we are checking if there is a authHeader and if it is empty we are returning a error saying that there is no auth header...
		return "" , nil , errors.New("no auth header")
	}


	//// Split the header on spaces....
	headerParts := strings.Split(authHeader , " ")
	if len(headerParts) != 2 {
		return "" , nil , errors.New("invalid auth header") // If we didn't get exactly two values then we are returning a error saying that it is a invalid auth header....
	}


	////// Check to see if we have the word Bearer....
	 if headerParts[0] != "Bearer" {
		return "" , nil , errors.New("invalid auth header") // If there is no word Bearer in the authHeader then also we throw an error saying that it is an invalid auth header....
	 }

	 token := headerParts[1] // If there is a word bearer then store the token in the variable token...


	 //// Declare an empty claims.....
	 claims := &Claims{} /// Declaring an empty claims variable....

	 //// Parse the token.....
	 _ , err := jwt.ParseWithClaims(token , claims , func(token *jwt.Token) (interface{} , error) {
		if _ , ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // verify that is there the correct signing method
			return nil , fmt.Errorf("unexpected signing method: %v" , token.Header["alg"])
		}
		return []byte(j.Secret) , nil
	 })


     if err != nil {
		if strings.HasPrefix(err.Error() , "token is expired by") {
			return "" , nil , errors.New("expired token")
		}
		return "" , nil , err
	 }


	 if claims.Issuer != j.Issuer { // Finally we are checking whether we have issued the token (or) not ...
		return "" , nil , errors.New("invalid issuer")
	 }

	 return token , claims , nil

  }


