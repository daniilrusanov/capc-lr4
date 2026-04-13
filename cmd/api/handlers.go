package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		errorJSON(w, errors.New("некоректний JSON"))
		return
	}

	user, ok := users[creds.Email]
	if !ok {
		errorJSON(w, errors.New("користувача не знайдено"), http.StatusUnauthorized)
		return
	}

	// Перевірка хешу пароля
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(creds.Password))
	if err != nil {
		errorJSON(w, errors.New("неправильний пароль"), http.StatusUnauthorized)
		return
	}

	// Генерація JWT [cite: 434]
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	tokenString, _ := token.SignedString([]byte(secret))

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "authenticated",
		"token":   tokenString,
	})
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		errorJSON(w, errors.New("відсутній токен"), http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		errorJSON(w, errors.New("недійсний токен"), http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "valid"})
}
