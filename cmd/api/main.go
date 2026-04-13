package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Завантажуємо змінні з .env
	_ = godotenv.Load()

	// Ініціалізуємо підключення до SQLite та створюємо таблиці
	InitDB()
	log.Println("Базу даних SQLite ініціалізовано.")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := routes()

	// Мідлвар для логування запитів
	loggingMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Запит: %s %s", r.Method, r.URL.Path)
		mux.ServeHTTP(w, r)
	})

	log.Printf("Auth service запущено на порту %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, loggingMux))
}
