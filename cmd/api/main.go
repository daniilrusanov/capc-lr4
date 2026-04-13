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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := routes()

	// Мідлвар для логування запитів [cite: 409]
	loggingMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Запит: %s %s", r.Method, r.URL.Path)
		mux.ServeHTTP(w, r)
	})

	log.Printf("Auth service запущено на порту %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, loggingMux))
}
