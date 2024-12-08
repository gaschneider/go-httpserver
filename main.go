package main

import (
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

func (cfg *apiConfig) getDisplayCountRequestsHandler() func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits.Load())
	}
}

func (cfg *apiConfig) getResetCountRequestsHandler() func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		cfg.fileserverHits.Swap(0)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	}
}

func main() {
	serveMux := http.NewServeMux()
	config := apiConfig{fileserverHits: atomic.Int32{}}
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", config.middlewareMetricsInc(fileServerHandler))
	serveMux.HandleFunc("/healthz", healthHandler)
	serveMux.HandleFunc("/metrics", config.getDisplayCountRequestsHandler())
	serveMux.HandleFunc("/reset", config.getResetCountRequestsHandler())
	server := http.Server{
		Addr:    ":8081",
		Handler: serveMux,
	}

	server.ListenAndServe()
}
