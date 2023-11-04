package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

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
