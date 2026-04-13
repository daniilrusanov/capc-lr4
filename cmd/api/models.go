package main

// Credentials описує дані, що приходять при логіні [cite: 508]
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User описує користувача в системі [cite: 509]
type User struct {
	ID             int
	Email          string
	HashedPassword string
}

// users — імпровізована база даних [cite: 509]
// Пароль для студента: "secret123" (захешовано)
var users = map[string]User{
	"student@example.com": {
		ID:             1,
		Email:          "student@example.com",
		HashedPassword: "$2a$10$8K9O2Mh8N5B1pM.vL2R9ueG.yXvM6Z8Vp8h.N3H6I3P2N1K5M6y1K",
	},
}
