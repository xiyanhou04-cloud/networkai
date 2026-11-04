package main

import (
	"fmt"
	"log"
	"net/http"

	"networkai/backend"
	"networkai/handler"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Println("started-service")
	if err := backend.InitElasticsearchBackend(); err != nil {
		return fmt.Errorf("initialize Elasticsearch: %w", err)
	}
	defer backend.ESBackend.Close()
	return http.ListenAndServe(":8080", handler.InitRouter())
}
