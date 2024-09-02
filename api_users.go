package main

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UserRequestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Expires  int64  `json:"expires_in_seconds"`
}

type Response struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
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
		Id:          id,
		Email:       params.Email,
		IsChirpyRed: false,
	}
	password, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to generate password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}
	users.Users[id] = User{
		Id:          id,
		Email:       params.Email,
		Password:    password,
		IsChirpyRed: false,
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

func (db *DB) userLogin(apiCfg apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		type Response struct {
			Id           int    `json:"id"`
			Email        string `json:"email"`
			IsChirpyRed  bool   `json:"is_chirpy_red"`
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
		}

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

		id, ok := users.Emails[params.Email]
		if !ok {
			respondWithError(w, http.StatusBadRequest, "invalid request")
			return
		}

		result := bcrypt.CompareHashAndPassword(users.Users[id].Password, []byte(params.Password))
		if result != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		currentTime := time.Now()

		token := apiCfg.generateJWT(currentTime, params.Expires, id)
		if token == "" {
			respondWithError(w, http.StatusInternalServerError, "server error")
		}

		refreshToken := generateRefreshToken()
		users.RefreshTokens[refreshToken] = RefreshToken{
			UserId:     id,
			Expiration: currentTime.Unix() + 5184000,
		}

		err = db.writeDB(users)
		if err != nil {
			log.Printf("failed to write db: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}

		response := Response{
			Id:           id,
			Email:        users.Users[id].Email,
			IsChirpyRed:  users.Users[id].IsChirpyRed,
			Token:        token,
			RefreshToken: refreshToken,
		}
		respondWithJSON(w, http.StatusOK, response)
	}

}

func (db *DB) updateUser(apiCfg apiConfig) func(http.ResponseWriter, *http.Request) {
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

		params, err := checkRequest(w, req)
		if err != nil {
			log.Printf("params check failed: %v", err)
			return
		}

		stringId, _ := parsedToken.Claims.GetSubject()
		id, _ := strconv.Atoi(stringId)

		users, err := db.loadDB()
		if err != nil {
			log.Printf("failed to get db: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}

		responseBody := Response{
			Id:          id,
			Email:       params.Email,
			IsChirpyRed: users.Users[id].IsChirpyRed,
		}
		password, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("failed to generate password: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}
		oldEmail := users.Users[id].Email

		users.Users[id] = User{
			Id:          id,
			Email:       params.Email,
			Password:    password,
			IsChirpyRed: users.Users[id].IsChirpyRed,
		}

		// keep the email map clean, so it doesn't cause issues
		if oldEmail != params.Email {
			delete(users.Emails, oldEmail)
			users.Emails[params.Email] = id
		}

		err = db.writeDB(users)
		if err != nil {
			log.Printf("failed to write db: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}

		respondWithJSON(w, http.StatusOK, responseBody)

	}
}

func (db *DB) polkaWebhook(w http.ResponseWriter, req *http.Request) {
	type requestParams struct {
		Event string         `json:"event"`
		Data  map[string]int `json:"data"`
	}

	db.mux.Lock()
	defer db.mux.Unlock()

	decoder := json.NewDecoder(req.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	users, err := db.loadDB()
	if err != nil {
		log.Printf("failed to get db: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	requestUserId, ok := params.Data["user_id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "user_id doesn't exist in data object")
		return
	}

	if _, ok := users.Users[requestUserId]; !ok {
		respondWithError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	tempUser := users.Users[requestUserId]
	tempUser.IsChirpyRed = true
	users.Users[requestUserId] = tempUser

	err = db.writeDB(users)
	if err != nil {
		log.Printf("failed to write db: %s", err)
		respondWithError(w, http.StatusInternalServerError, "server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func checkRequest(w http.ResponseWriter, req *http.Request) (UserRequestParams, error) {
	decoder := json.NewDecoder(req.Body)
	params := UserRequestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return UserRequestParams{}, err
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "no email address provided")
		return UserRequestParams{}, errors.New("no email address provided")
	}
	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "password can't be blank")
		return UserRequestParams{}, errors.New("blank password")
	}

	return params, nil
}
