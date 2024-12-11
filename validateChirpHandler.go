package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	type returnVals struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		CleanedBody string `json:"cleaned_body"`
	}
	respBody := returnVals{
		CleanedBody: replaceBadWordsInChirp(params.Body),
	}
	respondWithJSON(w, 200, respBody)
}
