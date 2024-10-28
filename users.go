package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ac *apiConfig) handlerCreateUser(rw http.ResponseWriter, req *http.Request) {
	type reqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	defer req.Body.Close()

	body := reqData{}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&body); err != nil {
		respondWithError(rw, http.StatusBadRequest, "Failed to parse request body", err)
		return
	}
	hashed_pwd, err := auth.HashPassword(body.Password)
	if err != nil {
		log.Printf("Failed to hash password: %s\n", err)
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	user, err := ac.dbQueries.CreateUser(req.Context(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: hashed_pwd,
	})

	if err != nil {
		log.Printf("Failed to create user: %s", err)
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	apiUser := User{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
	respondWithJSON(rw, http.StatusCreated, apiUser)
}

func (ac *apiConfig) handlerLogin(rw http.ResponseWriter, req *http.Request) {
	type reqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	body := reqData{}
	if err := decoder.Decode(&body); err != nil {
		respondWithError(rw, http.StatusBadRequest, "Invalid body", err)
		return
	}

	user, err := ac.dbQueries.FindUserByEmail(req.Context(), body.Email)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	err = auth.CheckPasswordHash(body.Password, user.HashedPassword)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			respondWithError(rw, http.StatusUnauthorized, "Invalid credentials", err)
			return
		}
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(rw, http.StatusOK, User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}
