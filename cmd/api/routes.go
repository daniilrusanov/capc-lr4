package main

import "net/http"

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", RegisterHandler)
	mux.HandleFunc("/login", LoginHandler)
	mux.HandleFunc("/refresh", RefreshHandler)
	mux.HandleFunc("/verify", VerifyHandler)

	return mux
}
