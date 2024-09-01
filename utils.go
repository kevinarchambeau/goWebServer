package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorReturnVals struct {
		Error string `json:"error"`
	}

	responseBody := errorReturnVals{
		Error: msg,
	}
	respondWithJSON(w, code, responseBody)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func cleanString(text string) string {
	naughty := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(text, " ")
	for i, word := range words {
		if slices.Contains(naughty, strings.ToLower(word)) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
