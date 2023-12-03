package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/honesea/go-chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	db             database.DB
	jwtSecret      string
	polkaApiKey    string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			return
		}
		cfg.fileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits)))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) listChirps(w http.ResponseWriter, r *http.Request) {
	authorIDStr := r.URL.Query().Get("author_id")
	authorID, err := strconv.Atoi(authorIDStr)
	if err != nil {
		authorID = 0
	}

	sortStr := r.URL.Query().Get("sort")
	sortDesc := false
	if sortStr == "desc" {
		sortDesc = true
	}

	chirpList, err := cfg.db.ListChirps(authorID, sortDesc)
	if err != nil {
		respondWithError(w, 500, "There was a problem retrieving chirps")
		return
	}

	respondWithJSON(w, 200, chirpList)
}

func (cfg *apiConfig) readChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDParam := chi.URLParam(r, "chirp_id")
	chirpID, err := strconv.Atoi(chirpIDParam)

	if err != nil {
		respondWithError(w, 400, "Chirp ID must be an integer")
		return
	}

	chirp, err := cfg.db.ReadChirp(chirpID)
	if err != nil {
		respondWithError(w, 500, "There was a problem retrieving chirps")
		return
	}

	// Chirp doesn't exist
	if chirp == (database.Chirp{}) {
		respondWithError(w, 404, "Chirp doesn't exist")
		return
	}

	respondWithJSON(w, 200, chirp)
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	userId, err := authenticate(cfg.jwtSecret, auth)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	cleanedBody := cleanProfanity(params.Body)

	chirp, err := cfg.db.CreateChirp(userId, cleanedBody)
	if err != nil {
		respondWithError(w, 500, "There was a problem creating the chirp")
		return
	}

	respondWithJSON(w, 201, chirp)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	userId, err := authenticate(cfg.jwtSecret, auth)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	chirpIDParam := chi.URLParam(r, "chirp_id")
	chirpID, err := strconv.Atoi(chirpIDParam)

	if err != nil {
		respondWithError(w, 400, "Chirp ID must be an integer")
		return
	}

	chirp, err := cfg.db.DeleteChirp(userId, chirpID)
	if err != nil {
		respondWithError(w, 403, "There was a problem deleteing the chirp")
		return
	}

	respondWithJSON(w, 200, chirp)
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.db.CreateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, 500, "There was a problem creating the user")
		return
	}

	respondWithJSON(w, 201, user)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	userId, err := authenticate(cfg.jwtSecret, auth)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.db.UpdateUser(userId, params.Email, params.Password)
	if err != nil {
		respondWithError(w, 500, "There was a problem creating the user")
		return
	}

	respondWithJSON(w, 200, user)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.db.Login(params.Email, params.Password)
	if err != nil || user == (database.User{}) {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	// User successfully authenticated so we can generate access tokens
	userIdStr := fmt.Sprintf("%v", user.ID)
	accessToken, accessErr := generateAccessToken(cfg.jwtSecret, userIdStr)
	refreshToken, refreshErr := generateRefreshToken(cfg.jwtSecret, userIdStr)
	if accessErr != nil || refreshErr != nil {
		respondWithError(w, 500, "Could not generate JWT")
		return
	}

	err = cfg.db.SaveRefreshToken(refreshToken)
	if err != nil {
		respondWithError(w, 500, "Could not save refresh token")
		return
	}

	access := struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}{
		ID:           user.ID,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	}

	respondWithJSON(w, 200, access)
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	userId, err := authenticateRefresh(cfg.jwtSecret, auth)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	splitAuth := strings.Split(auth, " ")
	token := splitAuth[1]

	valid := cfg.db.CheckRefreshToken(token)
	if !valid {
		respondWithError(w, 401, "The refresh token is invalid")
		return
	}

	userIdStr := fmt.Sprintf("%v", userId)
	accessToken, err := generateAccessToken(cfg.jwtSecret, userIdStr)
	if err != nil {
		respondWithError(w, 500, "Could not generate JWT")
		return
	}

	access := struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	}

	respondWithJSON(w, 200, access)
}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	_, err := authenticateRefresh(cfg.jwtSecret, auth)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	splitAuth := strings.Split(auth, " ")
	token := splitAuth[1]

	err = cfg.db.RevokeRefreshToken(token)
	if err != nil {
		respondWithError(w, 500, "There was an issue revoking the refresh token")
		return
	}

	access := struct {
		RevokedToken string `json:"token"`
	}{
		RevokedToken: token,
	}

	respondWithJSON(w, 200, access)
}

func (cfg *apiConfig) polkaWebhook(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	err := authenticatePolka(cfg.polkaApiKey, auth)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(200)
		return
	}

	err = cfg.db.ActivateChirpyRed(params.Data.UserID)
	if err != nil {
		respondWithError(w, 404, "There was a problem deleteing the chirp")
		return
	}

	w.WriteHeader(200)
}
