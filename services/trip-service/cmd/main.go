package main

import (
	"log"
	"net/http"
	"time"

	tripHandler "ride-sharing/services/trip-service/internal/infrastructure/http"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8083")
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}

func main() {
	log.Println("Starting trip service on port", httpAddr)

	mux := http.NewServeMux()

	inmemrepo := repository.NewInmemRepository()
	service := service.NewTripService(inmemrepo)
	handler := tripHandler.HttpHandler{Service: service}

	mux.HandleFunc("POST /preview", handler.HandleTripPreview)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: loggingMiddleware(mux),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Println("HTTP server error: %v", err)
		return
	}
}
