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
	db.mux.Lock()
	defer db.mux.Unlock()

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
		respondWithError(w, http.StatusBadRequest, "message is too long")
		return
	}

	chirps, err := db.loadDB()
	if err != nil {
		log.Printf("failed to get chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}
	db.chirps = db.currentChirpCount(chirps)
	db.chirps++
	id := db.chirps
	responseBody := Chirp{
		Id:   id,
		Body: cleanString(params.Body),
	}
	chirps.Chirps[id] = responseBody

	err = db.writeDB(chirps)
	if err != nil {
		log.Printf("failed to write chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondWithJSON(w, http.StatusCreated, responseBody)
}

func (db *DB) getAllChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := db.GetChirps()
	if err != nil {
		log.Printf("failed to get chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	respondWithJSON(w, 200, chirps)
}

func (db *DB) currentChirpCount(chirps DBStructure) int {
	return len(chirps.Chirps)
}
