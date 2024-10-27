package main

import (
	"chirpy/internal/database"
	"encoding/json"
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

func (ac *apiConfig) handlerCreateChirp(rw http.ResponseWriter, req *http.Request) {
	type createChirpBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type createChirpResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	params := createChirpBody{}
	err := decoder.Decode(&params)
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
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Failed to create chirp", err)
		return
	}

	respondWithJSON(rw, http.StatusCreated, createChirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
