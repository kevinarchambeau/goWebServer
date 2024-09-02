package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func (db *DB) revokeRefresh(w http.ResponseWriter, req *http.Request) {
	db.mux.Lock()
	defer db.mux.Unlock()

	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")

	users, err := db.loadDB()
	if err != nil {
		log.Printf("failed to get db: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	if _, ok := users.RefreshTokens[token]; !ok {
		respondWithError(w, http.StatusNotFound, "invalid token")
		return
	}
	delete(users.RefreshTokens, token)
	err = db.writeDB(users)
	if err != nil {
		log.Printf("failed to write db: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (db *DB) refresh(apiCfg apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		type Response struct {
			Token string `json:"token"`
		}

		db.mux.Lock()
		defer db.mux.Unlock()

		requestToken := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")

		users, err := db.loadDB()
		if err != nil {
			log.Printf("failed to get db: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}

		refreshToken, ok := users.RefreshTokens[requestToken]
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		if refreshToken.Expiration < time.Now().Unix() {
			// keep the db tidy
			delete(users.RefreshTokens, requestToken)
			err = db.writeDB(users)
			if err != nil {
				log.Printf("failed to write db: %s", err)
				respondWithError(w, http.StatusInternalServerError, "server error")
				return
			}
			respondWithError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		currentTime := time.Now()
		token := apiCfg.generateJWT(currentTime, 0, refreshToken.UserId)
		if token == "" {
			respondWithError(w, http.StatusInternalServerError, "server error")
		}

		response := Response{
			Token: token,
		}
		respondWithJSON(w, http.StatusOK, response)
	}

}
