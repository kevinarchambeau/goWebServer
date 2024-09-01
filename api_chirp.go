package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func (db *DB) chirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if len(params.Body) > 140 {
		respondWithError(w, 400, "message is too long")
		return
	}

	responseBody := Chirp{
		Id:   1,
		Body: cleanString(params.Body),
	}

	respondWithJSON(w, 201, responseBody)

}

func (db *DB) getChirps(w http.ResponseWriter, req *http.Request) {

}
