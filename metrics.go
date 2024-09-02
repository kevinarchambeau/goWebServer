package main

import (
	"fmt"
	"net/http"
)

func (apiCfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// need to do it like this, so the middleware is executed then it's served
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		apiCfg.fileserverHits++
		next.ServeHTTP(w, req)
	})
}

func (apiCfg *apiConfig) resetCount(w http.ResponseWriter, req *http.Request) {
	apiCfg.fileserverHits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Count reset"))
}

func (apiCfg *apiConfig) getCount(w http.ResponseWriter, req *http.Request) {
	template := "<html>\n\n<body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n</body>\n\n</html>"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, template, apiCfg.fileserverHits)
}
