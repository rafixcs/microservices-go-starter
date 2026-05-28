package main

import (
	"log"
	"net/http"
	"ride-sharing/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8083")
)

func main() {
	log.Println("Starting trip service on port", httpAddr)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /preview", handleCreateTrip)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Println("HTTP server error: %v", err)
		return
	}
}
