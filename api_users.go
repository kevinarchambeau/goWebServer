package main

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte
}

func (db *DB) createUser(w http.ResponseWriter, req *http.Request) {
	db.mux.Lock()
	defer db.mux.Unlock()

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "no email address provided")
		return
	}
	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "password can't be blank")
		return
	}

	users, err := db.loadDB()
	if err != nil {
		log.Printf("failed to get chirps: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}
	if _, ok := users.Emails[params.Email]; ok {
		respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	users.UserId++
	id := users.UserId
	responseBody := response{
		Id:    id,
		Email: params.Email,
	}
	password, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to generate password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}
	users.Users[id] = User{
		Id:       id,
		Email:    params.Email,
		Password: password,
	}
	users.Emails[responseBody.Email] = id

	err = db.writeDB(users)
	if err != nil {
		log.Printf("failed to write db: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondWithJSON(w, http.StatusCreated, responseBody)
}
