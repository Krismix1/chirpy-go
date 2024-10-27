package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ac *apiConfig) handlerCreateUser(rw http.ResponseWriter, req *http.Request) {
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
