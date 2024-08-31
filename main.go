package main

import (
	"fmt"
	"io"
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
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("./"))))
	mux.Handle("app/assets/", http.StripPrefix("/app", http.FileServer(http.Dir("./assets/"))))
	mux.HandleFunc("/healthz", healthz)
	err := http.ListenAndServe(serverConfig.Addr, mux)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func healthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "OK")
}
