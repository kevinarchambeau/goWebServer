package main

import (
	"fmt"
	"net/http"
)

type Server struct {
	Addr    string
	Handler http.ServeMux
}

func main() {
	mux := http.NewServeMux()
	serverConfig := Server{
		Addr: ":8080",
	}
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.Handle("assets/", http.FileServer(http.Dir("./assets/")))
	err := http.ListenAndServe(serverConfig.Addr, mux)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
