package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateAccessToken(jwtSecret string, subject string) (string, error) {
	issuedAt := time.Now()
	expiresAt := time.Now().Add(time.Minute * 60)

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   subject,
	}

	tokenSet := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenSet.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", errors.New("unauthorized")
	}

	return token, nil
}

func generateRefreshToken(jwtSecret string, subject string) (string, error) {
	issuedAt := time.Now()
	expiresAt := time.Now().Add(time.Minute * 86400)

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   subject,
	}

	tokenSet := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenSet.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", errors.New("unauthorized")
	}

	return token, nil
}

func authenticate(jwtSecret string, auth string) (int, error) {
	// Split 'Bearer ' from token
	splitAuth := strings.Split(auth, " ")
	if len(splitAuth) != 2 {
		return 0, errors.New("unauthorized")
	}

	token := splitAuth[1]
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return 0, errors.New("unauthorized")
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, errors.New("unauthorized")
	}

	if claims.Issuer != "chirpy-access" {
		return 0, errors.New("unauthorized")
	}

	return userId, nil
}

func authenticateRefresh(jwtSecret string, auth string) (int, error) {
	// Split 'Bearer ' from token
	splitAuth := strings.Split(auth, " ")
	if len(splitAuth) != 2 {
		return 0, errors.New("unauthorized")
	}

	token := splitAuth[1]
	claims := jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return 0, errors.New("unauthorized")
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, errors.New("unauthorized")
	}

	if claims.Issuer != "chirpy-refresh" {
		return 0, errors.New("unauthorized")
	}

	return userId, nil
}

func authenticatePolka(apiKey string, auth string) error {
	// Split 'Bearer ' from token
	splitAuth := strings.Split(auth, " ")
	if len(splitAuth) != 2 {
		return errors.New("unauthorized")
	}

	key := splitAuth[1]
	if key != apiKey {
		return errors.New("unauthorized")
	}

	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	errResponse := errorResponse{
		Error: msg,
	}

	respondWithJSON(w, code, errResponse)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func cleanProfanity(s string) string {
	profanity := [3]string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	stringReturn := s
	stringLower := strings.ToLower(s)

	for _, word := range profanity {
		for {
			index := strings.Index(stringLower, word)

			if index == -1 {
				break
			}

			stringLower = strings.Join([]string{stringLower[:index], stringLower[index+len(word):]}, "****")
			stringReturn = strings.Join([]string{stringReturn[:index], stringReturn[index+len(word):]}, "****")
		}
	}

	return stringReturn
}
