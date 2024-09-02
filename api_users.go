package main

import (
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte
}

type RequestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func (db *DB) createUser(w http.ResponseWriter, req *http.Request) {
	db.mux.Lock()
	defer db.mux.Unlock()

	params, err := checkRequest(w, req)
	if err != nil {
		log.Printf("params check failed: %v", err)
		return
	}

	users, err := db.loadDB()
	if err != nil {
		log.Printf("failed to get db: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}
	if _, ok := users.Emails[params.Email]; ok {
		respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	users.UserId++
	id := users.UserId
	responseBody := Response{
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

func (db *DB) userLogin(w http.ResponseWriter, req *http.Request) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	params, err := checkRequest(w, req)
	if err != nil {
		log.Printf("params check failed: %v", err)
		return
	}

	users, err := db.loadDB()
	if err != nil {
		log.Printf("failed to get db: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	id, ok := users.Emails[params.Email]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}

	result := bcrypt.CompareHashAndPassword(users.Users[id].Password, []byte(params.Password))
	if result == nil {
		response := Response{
			Id:    id,
			Email: users.Users[id].Email,
		}
		respondWithJSON(w, http.StatusOK, response)
	} else {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	}
}

func checkRequest(w http.ResponseWriter, req *http.Request) (RequestParams, error) {
	decoder := json.NewDecoder(req.Body)
	params := RequestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return RequestParams{}, err
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "no email address provided")
		return RequestParams{}, errors.New("no email address provided")
	}
	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "password can't be blank")
		return RequestParams{}, errors.New("blank password")
	}

	return params, nil
}
