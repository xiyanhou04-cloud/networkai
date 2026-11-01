package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func InitRouter() *mux.Router {
	router := mux.NewRouter()
	router.Handle("/posts", http.HandlerFunc(uploadHandler)).Methods("POST")
	return router
}
