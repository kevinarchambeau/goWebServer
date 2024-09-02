package main

import (
	"flag"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Couldn't load env file")
	}

	mux := http.NewServeMux()
	serverConfig := Server{
		Addr: ":8080",
	}
	apiCfg := apiConfig{
		fileserverHits: 0,
		jwtSecret:      os.Getenv("JWT_SECRET"),
		polkaKey:       os.Getenv("POLKA_KEY"),
	}

	db, err := NewDB(dbFile)
	if err != nil {
		log.Fatal("Can't connect to db")
	}

	mux.Handle("GET /app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.getCount)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /api/reset", apiCfg.resetCount)
	mux.HandleFunc("POST /api/chirps", db.createChirp(apiCfg))
	mux.HandleFunc("GET /api/chirps", db.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", db.getChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", db.deleteChirp(apiCfg))
	mux.HandleFunc("POST /api/users", db.createUser)
	mux.HandleFunc("PUT /api/users", db.updateUser(apiCfg))
	mux.HandleFunc("POST /api/login", db.userLogin(apiCfg))
	mux.HandleFunc("POST /api/revoke", db.revokeRefresh)
	mux.HandleFunc("POST /api/refresh", db.refresh(apiCfg))
	mux.HandleFunc("POST /api/polka/webhooks", db.polkaWebhook)
	http.ListenAndServe(serverConfig.Addr, mux)

}
