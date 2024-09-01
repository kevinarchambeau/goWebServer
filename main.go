package main

import (
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

	db, err := NewDB("database.json")
	if err != nil {
		log.Fatal("Can't connect to db")
	}
	//UsersDb, err := NewDB("users.json")
	//if err != nil {
	//	log.Fatal("Can't connect to db")
	//}

	mux.Handle("GET /app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.getCount)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /api/reset", apiCfg.resetCount)
	mux.HandleFunc("POST /api/chirps", db.createChirp)
	mux.HandleFunc("GET /api/chirps", db.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", db.getChirp)
	mux.HandleFunc("POST /api/users", db.createUser)
	http.ListenAndServe(serverConfig.Addr, mux)

}
