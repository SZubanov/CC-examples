package handler

import (
	"encoding/json"
	"net/http"

	"github.com/postman-automation/task-manager/internal/model"
	"github.com/postman-automation/task-manager/internal/service"
	"github.com/postman-automation/task-manager/internal/storage"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "validation_error", "Email is required")
		return
	}

	if req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "validation_error", "Password is required")
		return
	}

	resp, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		if err == storage.ErrUserExists {
			respondWithError(w, http.StatusConflict, "user_exists", "User with this email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "validation_error", "Email is required")
		return
	}

	if req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "validation_error", "Password is required")
		return
	}

	resp, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			respondWithError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}
