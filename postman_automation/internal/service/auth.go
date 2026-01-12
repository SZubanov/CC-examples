package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/postman-automation/task-manager/internal/model"
	"github.com/postman-automation/task-manager/internal/storage"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

const jwtSecret = "your-secret-key-change-in-production"

type AuthService struct {
	storage *storage.Storage
}

func NewAuthService(storage *storage.Storage) *AuthService {
	return &AuthService{storage: storage}
}

func (s *AuthService) Register(email, password string) (*model.RegisterResponse, error) {
	hashedPassword := hashPassword(password)
	user := model.NewUser(email, hashedPassword)

	if err := s.storage.CreateUser(user); err != nil {
		return nil, err
	}

	return &model.RegisterResponse{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

func (s *AuthService) Login(email, password string) (*model.LoginResponse, error) {
	user, err := s.storage.GetUserByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	hashedPassword := hashPassword(password)
	if user.Password != hashedPassword {
		return nil, ErrInvalidCredentials
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:  token,
		UserID: user.ID,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", ErrInvalidToken
		}
		return userID, nil
	}

	return "", ErrInvalidToken
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
