package handler

import (
	"encoding/json"
	"net/http"

	"github.com/postman-automation/task-manager/internal/model"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, err, message string) {
	resp := model.NewErrorResponse(code, err, message)
	respondWithJSON(w, code, resp)
}
