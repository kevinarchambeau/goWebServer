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
		type requestParams struct {
			Body string `json:"body"`
		}

		db.mux.Lock()
		defer db.mux.Unlock()

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
		params := requestParams{}
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
	authorIdString := req.URL.Query().Get("author_id")
	sort := req.URL.Query().Get("sort")

	allChirps, err := db.GetChirps()
	if err != nil {
		log.Printf("failed to get chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	if authorIdString == "" {
		if sort == "desc" {
			respondWithJSON(w, http.StatusOK, reverseChirps(allChirps))
		} else {
			respondWithJSON(w, http.StatusOK, allChirps)
		}
		return
	}

	authorId, err := strconv.Atoi(authorIdString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid author id")
		return
	}
	chirps := []Chirp{}
	for _, chirp := range allChirps {
		if authorId == chirp.AuthorId {
			chirps = append(chirps, chirp)
		}
	}
	if sort == "desc" {
		respondWithJSON(w, http.StatusOK, reverseChirps(chirps))
	} else {
		respondWithJSON(w, http.StatusOK, chirps)
	}
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

func (db *DB) deleteChirp(apiCfg apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		db.mux.Lock()
		defer db.mux.Unlock()

		claims := jwt.RegisteredClaims{}
		token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
		parsedToken, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(apiCfg.jwtSecret), nil
		})
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		temp, _ := parsedToken.Claims.GetSubject()
		tokenUserId, _ := strconv.Atoi(temp)

		chirps, err := db.loadDB()
		if err != nil {
			log.Printf("failed to get chirps: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}

		chirpId, err := strconv.Atoi(req.PathValue("chirpID"))
		if err != nil {
			log.Printf("failed to convert id to int: %s", err)
			respondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		if _, ok := chirps.Chirps[chirpId]; !ok {
			respondWithError(w, http.StatusNotFound, "Id does not exist")
			return
		}

		if tokenUserId != chirps.Chirps[chirpId].AuthorId {
			respondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		delete(chirps.Chirps, chirpId)
		err = db.writeDB(chirps)
		if err != nil {
			log.Printf("failed to write db: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func reverseChirps(allChirps []Chirp) []Chirp {
	chirps := []Chirp{}
	for i := len(allChirps) - 1; i >= 0; i-- {
		chirps = append(chirps, allChirps[i])
	}

	return chirps
}
