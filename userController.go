package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gaschneider/go/httpserver/internal/auth"
	"github.com/gaschneider/go/httpserver/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type LoggedUser struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) createUsersHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 400, "Error creating user")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashed_password,
	})
	if err != nil {
		respondWithError(w, 400, "Error creating user")
		return
	}

	userToJson := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, 201, userToJson)
}

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(1*time.Hour))
	if err != nil {
		respondWithError(w, 401, "Something went wrong generating token")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 401, "Something went wrong generating token")
		return
	}

	now := time.Now().UTC()
	// 60 dias
	expiration := now.Add(60 * 24 * time.Hour)

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: expiration,
	})

	if err != nil {
		respondWithError(w, 401, "Something went wrong generating token")
		return
	}

	userToJson := LoggedUser{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, 200, userToJson)
}

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	dbToken, err := cfg.db.GetByToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	now := time.Now().UTC()

	if dbToken.RevokedAt.Valid || dbToken.ExpiresAt.Before(now) {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	token, err := auth.MakeJWT(dbToken.UserID, cfg.secret, time.Duration(1*time.Hour))
	if err != nil {
		respondWithError(w, 401, "Something went wrong generating token")
		return
	}

	type ResponseBody struct {
		Token string `json:"token"`
	}

	body := ResponseBody{
		Token: token,
	}

	respondWithJSON(w, 200, body)
}

func (cfg *apiConfig) revokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 400, "Error revoking token")
		return
	}

	err = cfg.db.RevokeToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 400, "Error revoking token")
		return
	}

	w.WriteHeader(204)
}
