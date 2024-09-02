package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func (db *DB) createChirp(w http.ResponseWriter, req *http.Request) {
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
	chirps.ChirpId++
	id := chirps.ChirpId
	responseBody := Chirp{
		Id:   id,
		Body: cleanString(params.Body),
	}
	chirps.Chirps[id] = responseBody

	err = db.writeDB(chirps)
	if err != nil {
		log.Printf("failed to write db: %s", err)
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

	respondWithJSON(w, 200, chirps)
}

func (db *DB) getChirp(w http.ResponseWriter, req *http.Request) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	chirps, err := db.loadDB()
	if err != nil {
		log.Printf("failed to get chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	id, err := strconv.Atoi(req.PathValue("chirpID"))
	if err != nil {
		log.Printf("failed to convert id to int: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	if data, ok := chirps.Chirps[id]; ok {
		respondWithJSON(w, http.StatusOK, data)
		return
	}

	respondWithError(w, http.StatusNotFound, "Id does not exist")
}
