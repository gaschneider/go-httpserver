package main

import (
	"fmt"
	"io"
	"net/http"
)

func (cfg *apiConfig) displayCountRequestsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetCountRequestsHandler(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Swap(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}
