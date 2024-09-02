package main

import (
	"flag"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

type Server struct {
	Addr    string
	Handler http.ServeMux
}

type apiConfig struct {
	fileserverHits int
	jwtSecret      string
}

func main() {
	dbFile := "database.json"
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		err := os.Remove(dbFile)
		if err != nil {
			log.Fatal("Couldn't delete db file")
		}
	}
	godotenv.Load()

	mux := http.NewServeMux()
	serverConfig := Server{
		Addr: ":8080",
	}
	apiCfg := apiConfig{
		fileserverHits: 0,
		jwtSecret:      os.Getenv("JWT_SECRET"),
	}

	db, err := NewDB(dbFile)
	if err != nil {
		log.Fatal("Can't connect to db")
	}

	mux.Handle("GET /app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.getCount)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /api/reset", apiCfg.resetCount)
	mux.HandleFunc("POST /api/chirps", db.createChirp)
	mux.HandleFunc("GET /api/chirps", db.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", db.getChirp)
	mux.HandleFunc("POST /api/users", db.createUser)
	mux.HandleFunc("PUT /api/users", db.updateUser(apiCfg))
	mux.HandleFunc("POST /api/login", db.userLogin(apiCfg))
	mux.HandleFunc("POST /api/revoke", db.revokeRefresh)
	http.ListenAndServe(serverConfig.Addr, mux)

}
