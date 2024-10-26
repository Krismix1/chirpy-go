package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
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

func (ac *apiConfig) metrics(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/html; charset=utf-8")
	rw.Write([]byte(fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, ac.fileServerHits.Load())))
}

func (ac *apiConfig) reset(rw http.ResponseWriter, req *http.Request) {
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

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ac *apiConfig) createUser(rw http.ResponseWriter, req *http.Request) {
	type reqData struct {
		Email string `json:"email"`
	}

	defer req.Body.Close()

	body := reqData{}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&body); err != nil {
		respondWithError(rw, http.StatusBadRequest, "Failed to parse request body", err)
		return
	}

	user, err := ac.dbQueries.CreateUser(req.Context(), body.Email)

	if err != nil {
		log.Printf("Failed to create user: %s", err)
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	apiUser := User{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
	respondWithJSON(rw, http.StatusCreated, apiUser)
}
