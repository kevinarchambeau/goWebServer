package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	template := "<html>\n\n<body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n</body>\n\n</html>"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, template, cfg.fileserverHits)
}

func validateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type errorReturnVals struct {
		Error string `json:"error"`
	}

	type returnVals struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if len(params.Body) > 140 {
		responseBody := errorReturnVals{
			Error: "message is too long",
		}
		data, err := json.Marshal(responseBody)
		if err != nil {
			log.Printf("Error marshalling: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}
	responseBody := returnVals{
		Valid: true,
	}
	data, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("Error marshalling: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
