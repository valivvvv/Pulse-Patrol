package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewInMemoryDocumentStore()
	service := NewDocumentService(store)
	handler := NewHttpHandler(service)

	server := &http.Server{
		Addr:    ":18081",
		Handler: handler,
	}

	log.Println("Patient Document Service running on http://localhost:18081")
	log.Fatal(server.ListenAndServe())
}
