package main

import (
	"chirpy/internal/auth"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (ac *apiConfig) handlerPolkaWebhook(rw http.ResponseWriter, req *http.Request) {
	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil || apiKey != ac.polkaKey {
		respondWithError(rw, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	type reqData struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(req.Body)
	var data = reqData{}
	if err := decoder.Decode(&data); err != nil {
		respondWithError(rw, http.StatusBadRequest, "Invalid data", err)
		return
	}

	if data.Event != "user.upgraded" {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	rowsAffected, err := ac.dbQueries.UpdateUserToChirpyRed(req.Context(), data.Data.UserID)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}
	if rowsAffected == 0 {
		respondWithError(rw, http.StatusNotFound, fmt.Sprintf("User %s not found", data.Data.UserID), err)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
