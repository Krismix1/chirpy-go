package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func getCleanedMsg(msg string, badWords map[string]struct{}) string {
	words := strings.Split(msg, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func filterProfanity(msg string) string {
	var profanityWords = map[string]struct{}{"kerfuffle": {}, "sharbert": {}, "fornax": {}}
	return getCleanedMsg(msg, profanityWords)
}

type chirpInfo struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (ac *apiConfig) handlerCreateChirp(rw http.ResponseWriter, req *http.Request) {
	type createChirpBody struct {
		Body string `json:"body"`
	}

	defer req.Body.Close()

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Invalid authorization", nil)
		return
	}
	userID, err := auth.ValidateJWT(token, ac.tokenSecret)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Invalid authorization", nil)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := createChirpBody{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(rw, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirp, err := ac.dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   filterProfanity(params.Body),
		UserID: userID,
	})
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Failed to create chirp", err)
		return
	}

	respondWithJSON(rw, http.StatusCreated, chirpInfo{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (ac *apiConfig) handlerListAllChirps(rw http.ResponseWriter, req *http.Request) {
	chirps, err := ac.dbQueries.ListAllChirps(req.Context())
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Failed to get chirps", err)
		return
	}
	response := make([]chirpInfo, 0, len(chirps))
	for _, chirp := range chirps {
		response = append(response, chirpInfo{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	respondWithJSON(rw, http.StatusOK, response)
}

func (ac *apiConfig) GetChirpById(rw http.ResponseWriter, req *http.Request) {
	chirpId := req.PathValue("id")
	if chirpId == "" {
		respondWithError(rw, http.StatusBadRequest, "Chirp ID not specified", nil)
		return
	}
	id, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(rw, http.StatusBadRequest, "Chirp ID is invalid", err)
		return
	}

	chirp, err := ac.dbQueries.GetChirpById(req.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(rw, http.StatusNotFound, fmt.Sprintf("Chirp %s not found", id), err)
			return
		}
		respondWithError(rw, http.StatusInternalServerError, "Failed to get chirp", err)
		return
	}

	respondWithJSON(rw, http.StatusOK, chirpInfo{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
