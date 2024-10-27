package main

import (
	"chirpy/internal/database"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (ac *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ac.fileServerHits.Add(1)
		next.ServeHTTP(rw, req)
	})
}

func (ac *apiConfig) handlerMetrics(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/html; charset=utf-8")
	rw.Write([]byte(fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, ac.fileServerHits.Load())))
}

func (ac *apiConfig) handlerReset(rw http.ResponseWriter, req *http.Request) {
	if ac.platform != "dev" {
		respondWithError(rw, http.StatusForbidden, "Not allowed to call the endpoint", nil)
		return
	}

	ac.fileServerHits.Store(0)
	err := ac.dbQueries.DeleteAllUsers(req.Context())
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Failed to clear users", err)
		return
	}
}
