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

type RequestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Expires  int64  `json:"expires_in_seconds"`
}

func (db *DB) createUser(w http.ResponseWriter, req *http.Request) {
	db.mux.Lock()
	defer db.mux.Unlock()

	type Response struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}

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

func (db *DB) userLogin(apiCfg apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		db.mux.Lock()
		defer db.mux.Unlock()

		type Response struct {
			Id           int    `json:"id"`
			Email        string `json:"email"`
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
		}

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
		// default
		expireTime := currentTime.Unix() + 86400
		if params.Expires < 86400 && params.Expires > 0 {
			expireTime = currentTime.Unix() + params.Expires
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(time.Unix(expireTime, 0)),
			Subject:   strconv.Itoa(id),
		})

		signedToken, err := token.SignedString([]byte(apiCfg.jwtSecret))
		if err != nil {
			log.Printf("failed to sign token: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
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
			Token:        signedToken,
			RefreshToken: refreshToken,
		}
		respondWithJSON(w, http.StatusOK, response)
	}

}

func (db *DB) updateUser(apiCfg apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		db.mux.Lock()
		defer db.mux.Unlock()

		type Response struct {
			Id    int    `json:"id"`
			Email string `json:"email"`
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
			Id:    id,
			Email: params.Email,
		}
		password, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("failed to generate password: %s", err)
			respondWithError(w, http.StatusInternalServerError, "server error")
			return
		}
		oldEmail := users.Users[id].Email

		users.Users[id] = User{
			Id:       id,
			Email:    params.Email,
			Password: password,
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

func (db *DB) revokeRefresh(w http.ResponseWriter, req *http.Request) {

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
