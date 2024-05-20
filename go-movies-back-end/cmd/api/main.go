package main

import (
	"backend/internal/repository"
	"backend/internal/repository/dbrepo"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

const port = 8080

type application struct{
	DSN string
	Domain string
	DB repository.DatabaseRepo  // It is a pool of database connections...
	auth Auth
	JWTSecret string
	JWTIssuer string
	JWTAudience string
	CookieDomain string
	APIKey string
}

func main() { /// This is the entry point of  our go application...

	/// Set the application config ---> Application config stores the bits of information that my application needs

	password := "Secret"
	hash, _ := HashPassword(password) // ignore error for the sake of simplicity

	fmt.Println("Password:", password)
	fmt.Println("Hash:    ", hash)

	match := CheckPasswordHash(password, hash)
	fmt.Println("Match:   ", match)

	var app application

	/// Read from command line.....
	flag.StringVar(&app.DSN,"dsn","host=localhost port=54320 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5" , "postgres connection string")
	flag.StringVar(&app.JWTSecret , "jwt-secret" , "verysecret" , "signing secret")
	flag.StringVar(&app.JWTAudience , "jwt-audience" , "example.com" , "signing audience")
	flag.StringVar(&app.CookieDomain , "cookie-domain" , "localhost" , "cookie domain")
	flag.StringVar(&app.Domain , "domain" , "example.com" , "domain")
	flag.StringVar(&app.APIKey , "api-key" , "f59357caf14790ebad6ef99ba0e06077" , "api key")
    flag.Parse()
	/// Connect to the database....
  
	 conn , err := app.connectToDB()
	 if err != nil {
		log.Fatal(err)
	 }

	 app.DB = &dbrepo.PostgresDBRepo{DB : conn}

	 defer app.DB.Connection().Close()


	 app.auth = Auth {
		Issuer : app.JWTIssuer,
		Audience : app.JWTAudience,
		Secret : app.JWTSecret,
		TokenExpiry: time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath : "/",
		CookieName: "__Host-refresh_token",
		CookieDomain: app.CookieDomain,
	 }

 




	log.Println("Starting the application on port no" , port)

    




	//// Start a web server.....
	err = http.ListenAndServe(fmt.Sprintf(":%d" , port) , app.routes())

	if err != nil {
		log.Fatal(err);
	}

}