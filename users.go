package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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
	hashedPWD, err := auth.HashPassword(body.Password)
	if err != nil {
		log.Printf("Failed to hash password: %s\n", err)
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	user, err := ac.dbQueries.CreateUser(req.Context(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: hashedPWD,
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

	token, err := auth.MakeJWT(user.ID, ac.tokenSecret, 1*time.Hour)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}
	_, err = ac.dbQueries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(rw, http.StatusOK, User{
		ID:           user.ID,
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (ac *apiConfig) handlerRefreshToken(rw http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Must provide refresh token", err)
		return
	}
	refreshTokenInfo, err := ac.dbQueries.FindRefreshToken(req.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(rw, http.StatusUnauthorized, "Invalid credentials", err)
			return
		}
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}
	if refreshTokenInfo.ExpiresAt.UTC().Before(time.Now().UTC()) {
		respondWithError(rw, http.StatusUnauthorized, "Refresh token expired", nil)
		return
	}

	accessToken, err := auth.MakeJWT(refreshTokenInfo.UserID, ac.tokenSecret, 1*time.Hour)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Internal Server Error", err)
		return
	}

	type responseData struct {
		Token string `json:"token"`
	}
	respondWithJSON(rw, http.StatusOK, responseData{
		Token: accessToken,
	})
}

func (ac *apiConfig) handlerRevokeRefreshToken(rw http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Must provide refresh token", err)
		return
	}
	_, err = ac.dbQueries.FindRefreshToken(req.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(rw, http.StatusUnauthorized, "Invalid credentials", err)
			return
		}
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}
	err = ac.dbQueries.RevokeRefreshToken(req.Context(), refreshToken)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Internal Server Error", err)
		return
	}

	type responseData struct {
		Token string `json:"token"`
	}
	respondWithJSON(rw, http.StatusNoContent, nil)
}
