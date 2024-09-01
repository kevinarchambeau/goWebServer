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
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("/api/reset", apiCfg.resetCount)
	mux.HandleFunc("GET /api/metrics", apiCfg.getCount)
	err := http.ListenAndServe(serverConfig.Addr, mux)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func healthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// need to do it like this, so the middleware is executed then it's served
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) resetCount(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Count reset"))
}

func (cfg *apiConfig) getCount(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits)
}
