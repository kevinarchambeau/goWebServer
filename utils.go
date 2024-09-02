package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorReturnVals struct {
		Error string `json:"error"`
	}

	responseBody := errorReturnVals{
		Error: msg,
	}
	respondWithJSON(w, code, responseBody)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(data)
}

func cleanString(text string) string {
	naughty := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(text, " ")
	for i, word := range words {
		if slices.Contains(naughty, strings.ToLower(word)) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func generateRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)

	return hex.EncodeToString(b)

}

func (apiCfg *apiConfig) generateJWT(currentTime time.Time, expires int64, id int) string {

	// default
	expireTime := currentTime.Unix() + 3600
	if expires < 86400 && expires > 0 {
		expireTime = currentTime.Unix() + expires
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
		return ""
	}

	return signedToken
}
