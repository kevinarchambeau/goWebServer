package main

import "net/http"

type Server struct {
	Addr    string
	Handler http.ServeMux
}

type apiConfig struct {
	fileserverHits int
	jwtSecret      string
	polkaKey       string
}

type User struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
	Password    []byte
	IsChirpyRed bool `json:"is_chirpy_red"`
}

type RefreshToken struct {
	UserId     int   `json:"userId"`
	Expiration int64 `json:"expiration"`
}

type Chirp struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}
