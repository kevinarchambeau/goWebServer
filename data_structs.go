package main

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte
}

type RefreshToken struct {
	UserId     int   `json:"userId"`
	Expiration int64 `json:"expiration"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}
