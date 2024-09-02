package main

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (db *DB) createChirp(apiCfg apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		db.mux.Lock()
		defer db.mux.Unlock()

		type parameters struct {
			Body string `json:"body"`
		}

		claims := jwt.RegisteredClaims{}
		token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
		parsedToken, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(apiCfg.jwtSecret), nil
		})
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err = decoder.Decode(&params)
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
		chirpId := chirps.ChirpId

		temp, _ := parsedToken.Claims.GetSubject()
		userId, _ := strconv.Atoi(temp)
		responseBody := Chirp{
			Id:       chirpId,
			Body:     cleanString(params.Body),
			AuthorId: userId,
		}
		chirps.Chirps[chirpId] = responseBody

		err = db.writeDB(chirps)
		if err != nil {
			log.Printf("failed to write db: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}

		respondWithJSON(w, http.StatusCreated, responseBody)
	}
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
