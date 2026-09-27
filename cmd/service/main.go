package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Starting StudHub Microservice: club-service...")

	// Health check for the API Gateway
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("club-service is active and routing correctly"))
	})

	// TODO: Initialize gRPC interface for ML recommendation pipeline connection

	port := ":8080"
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
