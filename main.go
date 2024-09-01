package main

import (
	"fmt"
	"net/http"
)

type Server struct {
	Addr    string
	Handler http.ServeMux
}

type apiConfig struct {
	fileserverHits int
}

func main() {
	mux := http.NewServeMux()
	serverConfig := Server{
		Addr: ":8080",
	}
	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	mux.Handle("GET /app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.getCount)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /api/reset", apiCfg.resetCount)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	err := http.ListenAndServe(serverConfig.Addr, mux)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
