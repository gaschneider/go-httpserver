package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gaschneider/go/httpserver/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlePolkaEvents(w http.ResponseWriter, r *http.Request) {
	type bodyData struct {
		UserId uuid.UUID `json:"user_id"`
	}
	type body struct {
		Event string   `json:"event"`
		Data  bodyData `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := body{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	_, err = cfg.db.GetUser(r.Context(), params.Data.UserId)
	if err != nil {
		respondWithError(w, 404, "User not found")
		return
	}

	_, err = cfg.db.UpdateUserChirpyRed(r.Context(), database.UpdateUserChirpyRedParams{
		IsChirpyRed: true,
		ID:          params.Data.UserId,
	})
	if err != nil {
		respondWithError(w, 400, "Error updating user")
		return
	}

	w.WriteHeader(204)
}
