package main

import (
	"backend/internal/graph"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
)

//// Every handler in the go world takes two arguments

func (app *application) Home(w http.ResponseWriter , r *http.Request) { /// It is the default route to our API
	var payload = struct {   /// Here we have created a variable payload
		Status string `json:"status"`
		Message string  `json:"message"`
		Version string `json:"version"`
	} {   //// In this step we have populated the variable payload....
		Status : "active",
		Message : "Go movies up and running",
		Version : "1.0.0",
	}

   _ = app.writeJSON(w , http.StatusOK , payload)

}


func (app *application) AllMovies(w http.ResponseWriter , r *http.Request) {
	// var movies []models.Movie

	// rd , _ := time.Parse("2006-01-02" , "1986-03-07")


	// highlander := models.Movie {
	// 	ID : 1,
	// 	TITLE: "Highlander",
	// 	ReleaseDate: rd,
	// 	MPAARating: "R",
	// 	RunTime : 116,
	// 	Description: "A very nice movie",
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// }


	


    //  movies = append(movies , highlander)


	//  rd , _ = time.Parse("2006-01-02" , "1981-06-12")

	//  rotla := models.Movie {
	// 	ID : 2,
	// 	TITLE: "Raiders of the last arc",
	// 	ReleaseDate: rd,
	// 	MPAARating: "PG-13",
	// 	RunTime : 115,
	// 	Description: "A very good movie",
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// }

	// movies = append(movies , rotla)


	 movies , err := app.DB.AllMovies()

	 if err != nil {
		app.errorJSON(w,err)
		return
	 }



	 _ = app.writeJSON(w , http.StatusOK , movies)

}


 func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	/// Read the json payload

	var requestPayload struct {  // We have read the JSON payload here...
		Email string `json:"email"`
		Password string `json:"password"`
	}


	 err := app.readJSON(w,r, &requestPayload)

	 if err != nil {
		app.errorJSON(w,err,http.StatusBadRequest)
		return
	 }


	/// Validate the user against the database 
	user , err := app.DB.GetUserByEmail(requestPayload.Email)

	if err != nil {
		app.errorJSON(w , errors.New("invalid credentials") , http.StatusBadRequest)
		return
	}

	/// Check the password  which should match with the password that is stored in the database....
	valid , err := user.PasswordMatches(requestPayload.Password)

	if err != nil || !valid {
		app.errorJSON(w,errors.New("invalid credentials") , http.StatusBadRequest)
		return
	}


	/// Create a JWT User....
	u := jwtUser {
		ID : user.ID,
		FirstName : user.FirstName ,
		LastName : user.LastName,
	}


	  //// Generate tokens...
	  tokens , err := app.auth.GenerateTokenPair(&u)
	  if err != nil {
		app.errorJSON(w,err)
		return
	  }



	  refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)

	  http.SetCookie(w,refreshCookie)


	//   w.Write([]byte(tokens.Token)) /// This is writing the JWT to the browser window

	app.writeJSON(w,http.StatusAccepted , tokens)
 }


  func (app *application)  refreshToken(w http.ResponseWriter , r *http.Request) {
	for _,cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value


			// Parse the token to get the claims
			_ , err := jwt.ParseWithClaims(refreshToken , claims , func(token *jwt.Token) (interface{} , error) {
				return []byte(app.JWTSecret) , nil
			})

			if err != nil {
				app.errorJSON(w , errors.New("unauthorized") , http.StatusUnauthorized)
				return
			}


			//// Get the userId from the token claims....
			userID , err := strconv.Atoi(claims.Subject)

			if err != nil {
				app.errorJSON(w , errors.New("unknown user") , http.StatusUnauthorized)
				return
			}


			user , err := app.DB.GetUserByID(userID)

			
			if err != nil {
				app.errorJSON(w , errors.New("unknown user") , http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID : user.ID,
				FirstName : user.FirstName,
				LastName : user.LastName,
			}

			tokenPairs , err := app.auth.GenerateTokenPair(&u)
			
			if err != nil {
				app.errorJSON(w , errors.New("error generating tokens") , http.StatusUnauthorized)
				return
			}


			http.SetCookie(w , app.auth.GetRefreshCookie(tokenPairs.RefreshToken))

			app.writeJSON(w , http.StatusOK , tokenPairs) 
		}
	}
  }


  func (app *application)  logout(w http.ResponseWriter , r *http.Request)  {
	  http.SetCookie(w , app.auth.GetExpiredRefreshCookie())
	  w.WriteHeader(http.StatusAccepted)
  }


  func (app *application) MovieCatalog(w http.ResponseWriter , r *http.Request) {
	movies , err := app.DB.AllMovies()

	if err != nil {
	   app.errorJSON(w,err)
	   return
	}



	_ = app.writeJSON(w , http.StatusOK , movies)
  }



  func (app *application) GetMovie(w http.ResponseWriter , r *http.Request)  {

       id := chi.URLParam(r , "id")

	   movieID , err := strconv.Atoi(id) // We are converting an alphabet to an integer....

	   if err != nil {
		app.errorJSON(w,err)
		return 
	   }


	   movie , err := app.DB.OneMovie(movieID)
	   if err != nil {
		app.errorJSON(w,err)
		return 
	   }


	   _ = app.writeJSON(w , http.StatusOK , movie)


  }

  func (app *application) MovieForEdit(w http.ResponseWriter , r *http.Request)  {
	id := chi.URLParam(r , "id")

	movieID , err := strconv.Atoi(id) // We are converting an alphabet to an integer....

	if err != nil {
	 app.errorJSON(w,err)
	 return 
	}


	 movie , genres , err := app.DB.OneMovieForEdit(movieID)
	 if err != nil {
		app.errorJSON(w,err)
		return 
	   }

	   var payload =  struct {
		Movie *models.Movie `json:"movie"`
		Genres []*models.Genre `json:"genres"`
	   } {
		movie,
		genres,
	   }

	   _ = app.writeJSON(w , http.StatusOK , payload)
  }



   func (app *application) AllGenres(w http.ResponseWriter , r *http.Request) {
	 genres , err := app.DB.AllGenres()
	 if err != nil {
		app.errorJSON(w,err)
		return
	 }

	  _ = app.writeJSON(w , http.StatusOK , genres)
   }


   func (app *application) InsertMovie(w http.ResponseWriter , r *http.Request) {
	  var movie models.Movie


	   err := app.readJSON(w , r, &movie)
 
	    if (err != nil) {
			app.errorJSON(w , err)
			return
		}

		//// Try to get an image....

		movie = app.getPoster(movie)

		movie.CreatedAt = time.Now()
		movie.UpdatedAt = time.Now()


		newID , err := app.DB.InsertMovie(movie)
		if (err != nil) {
			app.errorJSON(w , err)
			return
		}


		//// Now handle the genres...
		 err = app.DB.UpdateMovieGenres(newID , movie.GenresArray)
		 if (err != nil) {
			app.errorJSON(w , err)
			return
		}

		resp := JSONResponse {
			Error : false,
			Message : "movie updated",
		}

		app.writeJSON(w , http.StatusAccepted , resp)

   }


   func (app *application) getPoster(movie models.Movie) models.Movie {
	  type TheMovieDB struct {
		Page int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`
		TotalPages int `json:"total_pages"`
	  }

	  client := &http.Client{} /// This is a standard library in GO which allows to make remote requests to sites
	  theUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s" , app.APIKey)

	  req , err := http.NewRequest("GET" , theUrl+"&query="+url.QueryEscape(movie.TITLE) , nil)

	  if err != nil {
		log.Println(err)
		return movie
	  }

       req.Header.Add("Accept" , "application/json")
	   req.Header.Add("Content-Type" , "application/json")

	   resp , err := client.Do(req)

	   if err != nil {
		log.Println(err)
		return movie
	  }

	  defer resp.Body.Close()

	  bodyBytes , err := io.ReadAll(resp.Body)

	  
	  if err != nil {
		log.Println(err)
		return movie
	  }

	  var responseObject TheMovieDB

	  json.Unmarshal(bodyBytes , &responseObject)

	  if len(responseObject.Results) > 0 {
		movie.Image = responseObject.Results[0].PosterPath
	  }

	  return movie
   }


   func (app *application) UpdateMovie(w http.ResponseWriter , r *http.Request) {

      var payload models.Movie


	   err := app.readJSON(w , r , &payload)


	   if err != nil {
		  app.errorJSON(w,err)
		  return
	   }


	   movie , err := app.DB.OneMovie(payload.ID)
	   if err != nil {
		app.errorJSON(w,err)
		return
	 }


	   movie.TITLE = payload.TITLE
	   movie.ReleaseDate = payload.ReleaseDate
	   movie.Description = payload.Description
	   movie.MPAARating = payload.MPAARating

	   movie.RunTime = payload.RunTime

	   movie.UpdatedAt = time.Now()


	   err = app.DB.UpdateMovie(*movie)
	   if err != nil {
		app.errorJSON(w,err)
		return
	 }


	  err = app.DB.UpdateMovieGenres(movie.ID , payload.GenresArray)
	  if err != nil {
		app.errorJSON(w,err)
		return
	 }

	  resp := JSONResponse {
		Error : false,
		Message : "movie updated",
	  }


	  app.writeJSON(w , http.StatusAccepted , resp)

   }


    func (app *application) DeleteMovie(w http.ResponseWriter , r *http.Request) {
 
		 id , err := strconv.Atoi(chi.URLParam(r , "id"))

		 if err != nil {
			app.errorJSON(w , err)
			return
		 }


		 err = app.DB.DeleteMovie(id)

		 if err != nil {
			app.errorJSON(w , err)
			return
		 }

		  
		  resp := JSONResponse {
			Error : false,
			Message : "movie deleted",
		  }

		  app.writeJSON(w , http.StatusAccepted , resp)


	}


	func (app *application) AllMoviesByGenre(w http.ResponseWriter , r *http.Request) {
 
		id , err := strconv.Atoi(chi.URLParam(r , "id"))

		if err != nil {
		   app.errorJSON(w , err)
		   return
		}


		movies , err := app.DB.AllMovies(id)
		if err != nil {
			app.errorJSON(w , err)
			return
		 }



		 app.writeJSON(w , http.StatusOK , movies)



	}

	 func (app *application) moviesGraphQL(w http.ResponseWriter , r *http.Request) {
		/// We need to populate the graph type with the movies....
		movies , _ := app.DB.AllMovies()


		//// Get the query from the request.....
		 q , _ := io.ReadAll(r.Body)
		  query := string(q) // We has the query as a string....


		//// Create a new variable of type *graph.Graph....
		  g := graph.New(movies)


		/// Set the query string on the variable....
		g.QueryString = query


		/// Perform the query.....
		  resp , err := g.Query()
		   if err != nil {
			app.errorJSON(w , err)
			return
		   }


		/// Send the response.....
		j , _ := json.MarshalIndent(resp , "" , "\t")
		w.Header().Set("Content-Type" , "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	 }