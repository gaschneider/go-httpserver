package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gaschneider/go/httpserver/internal/auth"
	"github.com/gaschneider/go/httpserver/internal/database"
	"github.com/google/uuid"
)

var badWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

func replaceBadWordsInChirp(chirp string) string {
	words := strings.Split(chirp, " ")

	for i, w := range words {
		if badWords[strings.ToLower(w)] {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

type ChirpDTO struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserId    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   replaceBadWordsInChirp(params.Body),
		UserID: userID,
	})

	if err != nil {
		respondWithError(w, 400, "Error creating chirp")
		return
	}

	respBody := ChirpDTO{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserId:    chirp.UserID,
		Body:      chirp.Body,
	}

	respondWithJSON(w, 201, respBody)
}

func (cfg *apiConfig) getAllChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())

	if err != nil {
		respondWithError(w, 400, "Error retrieving chirps")
		return
	}

	respBody := make([]ChirpDTO, 0)

	for _, chirp := range chirps {
		jsonChirp := ChirpDTO{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserId:    chirp.UserID,
			Body:      chirp.Body,
		}

		respBody = append(respBody, jsonChirp)
	}

	respondWithJSON(w, 200, respBody)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	requestedChirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp ID")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), requestedChirpID)
	if err != nil {
		respondWithError(w, 404, "Error retrieving chirps")
		return
	}

	respBody := ChirpDTO{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserId:    chirp.UserID,
		Body:      chirp.Body,
	}

	respondWithJSON(w, 200, respBody)
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	requestedChirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp ID")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), requestedChirpID)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, 403, "Access denied")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), requestedChirpID)
	if err != nil {
		respondWithError(w, 400, "Error deleting chirp")
		return
	}

	w.WriteHeader(204)
}
