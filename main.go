package main

import (
	"database/sql"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/gaschneider/go/httpserver/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
	polkaKey       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return
	}

	dbQueries := database.New(db)

	serveMux := http.NewServeMux()
	config := apiConfig{fileserverHits: atomic.Int32{}, db: dbQueries, platform: platform, secret: secret, polkaKey: polkaKey}
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", config.middlewareMetricsInc(fileServerHandler))
	serveMux.HandleFunc("GET /admin/metrics", config.displayCountRequestsHandler)
	serveMux.HandleFunc("POST /admin/reset", config.resetCountRequestsHandler)

	serveMux.HandleFunc("POST /api/chirps", config.createChirpHandler)
	serveMux.HandleFunc("GET /api/chirps", config.getAllChirpHandler)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", config.getChirpHandler)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", config.deleteChirpHandler)

	serveMux.HandleFunc("POST /api/users", config.createUsersHandler)
	serveMux.HandleFunc("POST /api/login", config.loginUserHandler)
	serveMux.HandleFunc("POST /api/refresh", config.refreshTokenHandler)
	serveMux.HandleFunc("POST /api/revoke", config.revokeRefreshTokenHandler)
	serveMux.HandleFunc("PUT /api/users", config.updateUserHandler)

	serveMux.HandleFunc("POST /api/polka/webhooks", config.handlePolkaEvents)

	serveMux.HandleFunc("GET /api/healthz", healthHandler)

	server := http.Server{
		Addr:    ":8081",
		Handler: serveMux,
	}

	server.ListenAndServe()
}
