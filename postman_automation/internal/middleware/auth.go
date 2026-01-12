package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/postman-automation/task-manager/internal/model"
	"github.com/postman-automation/task-manager/internal/service"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "missing_authorization", "Authorization header is required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "invalid_authorization", "Authorization header must be Bearer token")
			return
		}

		token := parts[1]
		userID, err := m.authService.ValidateToken(token)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func respondWithError(w http.ResponseWriter, code int, err, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := model.NewErrorResponse(code, err, message)
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := encodeJSON(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func encodeJSON(w http.ResponseWriter, v interface{}) error {
	enc := json.NewEncoder(w)
	return enc.Encode(v)
}
