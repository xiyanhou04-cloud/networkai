package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"networkai/model"
	"networkai/service"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse from body of request to get a json object.
	fmt.Println("Received one upload request")
	decoder := json.NewDecoder(r.Body)
	var p model.Post
	if err := decoder.Decode(&p); err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "Post received: %s\n", p.Message)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	keywords := r.URL.Query().Get("keywords")
	var posts []model.Post
	var err error
	if user != "" {
		posts, err = service.SearchPostsByUser(r.Context(), user)
	} else {
		posts, err = service.SearchPostsByKeywords(r.Context(), keywords)
	}
	if err != nil {
		http.Error(w, "Failed to read post from backend", http.StatusInternalServerError)
		fmt.Printf("Failed to read post from backend: %v\n", err)
		return
	}
	js, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, "Failed to parse posts into JSON format", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
