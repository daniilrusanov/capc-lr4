package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Хелпер для генерації пари токенів
func generateTokens(email string) (string, string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	// 1. Access Token (живе коротко, наприклад 15 хвилин)
	accToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	})
	accessTokenString, err := accToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	// 2. Refresh Token (живе довго - 7 днів)
	refToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshTokenString, err := refToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// РЕЄСТРАЦІЯ
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		errorJSON(w, errors.New("некоректний JSON"))
		return
	}

	// Хешуємо пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 10)
	if err != nil {
		errorJSON(w, errors.New("помилка хешування пароля"), http.StatusInternalServerError)
		return
	}

	// Зберігаємо в БД
	_, err = DB.Exec("INSERT INTO users (email, password) VALUES (?, ?)", creds.Email, string(hashedPassword))
	if err != nil {
		errorJSON(w, errors.New("користувач з таким email вже існує"), http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "користувача успішно зареєстровано"})
}

// ЛОГІН
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		errorJSON(w, errors.New("некоректний JSON"))
		return
	}

	// Шукаємо користувача в БД
	var hashedPassword string
	err := DB.QueryRow("SELECT password FROM users WHERE email = ?", creds.Email).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			errorJSON(w, errors.New("невірний email або пароль"), http.StatusUnauthorized)
		} else {
			errorJSON(w, err, http.StatusInternalServerError)
		}
		return
	}

	// Перевіряємо пароль
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password))
	if err != nil {
		errorJSON(w, errors.New("невірний email або пароль"), http.StatusUnauthorized)
		return
	}

	// Генеруємо токени
	accessToken, refreshToken, err := generateTokens(creds.Email)
	if err != nil {
		errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	// Зберігаємо refresh_token в БД
	_, _ = DB.Exec("UPDATE users SET refresh_token = ? WHERE email = ?", refreshToken, creds.Email)

	// Повертаємо обидва токени
	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// ОНОВЛЕННЯ ТОКЕНА (REFRESH)
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, errors.New("некоректний JSON"))
		return
	}

	secret := []byte(os.Getenv("JWT_SECRET"))

	// Валідуємо refresh токен
	token, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		errorJSON(w, errors.New("недійсний refresh токен"), http.StatusUnauthorized)
		return
	}

	// Дістаємо email з токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		errorJSON(w, errors.New("помилка читання токена"), http.StatusUnauthorized)
		return
	}
	email := claims["sub"].(string)

	// Перевіряємо, чи цей токен відповідає тому, що в базі (захист від викрадення)
	var dbToken string
	err = DB.QueryRow("SELECT refresh_token FROM users WHERE email = ?", email).Scan(&dbToken)
	if err != nil || dbToken != req.RefreshToken {
		errorJSON(w, errors.New("refresh токен відкликано або не існує"), http.StatusUnauthorized)
		return
	}

	// Генеруємо нові токени
	newAccess, newRefresh, err := generateTokens(email)
	if err != nil {
		errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	// Оновлюємо в базі
	_, _ = DB.Exec("UPDATE users SET refresh_token = ? WHERE email = ?", newRefresh, email)

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  newAccess,
		"refresh_token": newRefresh,
	})
}

// ПЕРЕВІРКА (ВЕРИФІКАЦІЯ)
func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		errorJSON(w, errors.New("відсутній токен"), http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		errorJSON(w, errors.New("недійсний або протермінований access токен"), http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "valid"})
}
