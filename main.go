package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	serveMux := http.NewServeMux()
	config := apiConfig{fileserverHits: atomic.Int32{}}
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", config.middlewareMetricsInc(fileServerHandler))
	serveMux.HandleFunc("GET /admin/metrics", config.displayCountRequestsHandler)
	serveMux.HandleFunc("POST /admin/reset", config.resetCountRequestsHandler)
	serveMux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	serveMux.HandleFunc("GET /api/healthz", healthHandler)

	server := http.Server{
		Addr:    ":8081",
		Handler: serveMux,
	}

	server.ListenAndServe()
}
