package handler

import (
	"net/http"

	jwtMiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/form3tech-oss/jwt-go"

	"github.com/gorilla/mux"
)

func InitRouter() *mux.Router {

	jwtMiddleware := jwtMiddleware.New(jwtMiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(mySigningKey), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	router := mux.NewRouter()
	router.Handle("/posts", jwtMiddleware.Handler(http.HandlerFunc(uploadHandler))).Methods("POST")
	router.Handle("/posts/search", jwtMiddleware.Handler(http.HandlerFunc(searchHandler))).Methods("GET")

	router.Handle("/users", http.HandlerFunc(signupHandler)).Methods("POST")
	router.Handle("/auth/login", http.HandlerFunc(signinHandler)).Methods("POST")

	return router
}
