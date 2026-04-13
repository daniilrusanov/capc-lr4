package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// DB - глобальна змінна для підключення до бази даних
var DB *sql.DB

func InitDB() {
	var err error
	// Відкриваємо (або створюємо) файл бази даних SQLite
	DB, err = sql.Open("sqlite", "./auth.db")
	if err != nil {
		log.Fatal("Помилка підключення до БД:", err)
	}

	// Створюємо таблицю користувачів, якщо її ще немає
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		refresh_token TEXT
	);`

	_, err = DB.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Помилка створення таблиці:", err)
	}
}

// Credentials описує дані, що приходять при логіні/реєстрації
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest описує запит на оновлення токена
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
