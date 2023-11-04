package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/honesea/go-chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	db             database.DB
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
	chirpList, err := cfg.db.ListChirps()
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
	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	cleanedBody := cleanProfanity(params.Body)

	chirp, err := cfg.db.CreateChirp(cleanedBody)
	if err != nil {
		respondWithError(w, 500, "There was a problem creating the chirp")
		return
	}

	respondWithJSON(w, 201, chirp)
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.db.CreateUser(params.Email)
	if err != nil {
		respondWithError(w, 500, "There was a problem creating the user")
		return
	}

	respondWithJSON(w, 201, user)
}
