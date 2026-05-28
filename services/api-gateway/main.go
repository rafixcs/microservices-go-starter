package main

import (
	"log"
	"net/http"
	"ride-sharing/shared/env"
)

var (
	httpAddr        = env.GetString("HTTP_ADDR", ":8081")
	tripServiceAddr = env.GetString("HTTP_TRIP_ADDR", "http://trip-service:8083/")
)

func main() {
	log.Println("Starting API Gateway")

	mux := http.NewServeMux()

	mux.HandleFunc("POST /trip/preview", handleTripPreview)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Println("HTTP server error: %v", err)
	}
}
