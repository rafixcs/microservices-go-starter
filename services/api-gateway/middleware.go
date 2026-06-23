package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqBody, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(reqBody))

		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		handler(rec, r)

		log.Printf("%s %s %d %s reqBody=%s respBody=%s",
			r.Method, r.URL.Path, rec.status, time.Since(start), reqBody, rec.body.Bytes())
	}
}

func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// allow preflight requests from the browser API
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}
