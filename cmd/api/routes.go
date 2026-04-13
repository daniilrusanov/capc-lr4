package main

import "net/http"

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", RegisterHandler) // Новий маршрут реєстрації
	mux.HandleFunc("/login", LoginHandler)
	mux.HandleFunc("/refresh", RefreshHandler) // Новий маршрут оновлення токенів
	mux.HandleFunc("/verify", VerifyHandler)

	return mux
}
