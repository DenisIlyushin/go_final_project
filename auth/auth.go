package auth

import (
	"encoding/json"
	"errors"
	"github.com/DenisIlyushin/go_final_project/config"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrEmptyPassword   = errors.New("переменная TODO_PASSWORD не задана в конфиге")
	ErrInvalidPassword = errors.New("неверный пароль")
	ErrInvalidToken    = errors.New("некорректный токен")
)

type Service struct {
	settings *config.Settings
	jwtKey   []byte
}

func NewService(config *config.Settings) *Service {
	pass := config.AuthPassword
	var key []byte
	if pass != "" {
		key = []byte(pass)
	}
	return &Service{settings: config, jwtKey: key}
}

type Credentials struct {
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// HandleSignin обрабатывает вход по POST /api/signin.
func (s *Service) Signin(w http.ResponseWriter, r *http.Request) {
	if s.settings.AuthPassword == "" {
		// Аутентификация не настроена
		http.Error(w, ErrEmptyPassword.Error(), http.StatusInternalServerError)
		return
	}
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TokenResponse{Error: err.Error()})
		return
	}
	if creds.Password != s.settings.AuthPassword {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TokenResponse{Error: ErrInvalidPassword.Error()})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(config.TokenLifetimeInHours * time.Hour).Unix(),
		"pwd": creds.Password,
	})
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TokenResponse{Error: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(TokenResponse{Token: tokenString})
}

// ValidateToken проверяет токен на валидность и соответствие текущему паролю.
func (s *Service) ValidateToken(tokenStr string) error {
	if s.settings.AuthPassword == "" {
		// Аутентификация отключена
		return nil
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtKey, nil
	})
	if err != nil || !token.Valid {
		return ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["pwd"] != s.settings.AuthPassword {
		return ErrInvalidToken
	}
	return nil
}

// Middleware проверяет JWT из куки "token".
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// если пароль в конфиге пуст — аутентификация отключена
		if s.settings.AuthPassword != "" {
			cookie, err := r.Cookie("token")
			if err != nil || s.ValidateToken(cookie.Value) != nil {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
