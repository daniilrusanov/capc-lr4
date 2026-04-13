package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Drug struct {
	ID           int
	Name         string
	Manufacturer string
	Price        float64
}

type TemplateData struct {
	Drugs           []Drug
	Drug            Drug
	SearchQuery     string
	CurrentDate     string
	FeedbackName    string
	FeedbackMessage string
	SuccessMessage  string // Виправляє помилку обірваної форми
	IsAuthenticated bool   // Виправляє помилку статичної шапки
}

var catalog = []Drug{
	{ID: 1, Name: "Парацетамол", Manufacturer: "Дарниця", Price: 35.50},
	{ID: 2, Name: "Аспірин", Manufacturer: "Bayer", Price: 120.00},
	{ID: 3, Name: "Ібупрофен", Manufacturer: "Фармак", Price: 65.00},
	{ID: 4, Name: "Но-шпа", Manufacturer: "Sanofi", Price: 150.00},
	{ID: 5, Name: "Цитрамон", Manufacturer: "Здоров'я", Price: 25.00},
}

const authServiceURL = "http://localhost:8081"

func main() {
	_ = godotenv.Load()
	_ = os.Getenv("JWT_SECRET")

	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Перевірка через API
			req, _ := http.NewRequest("GET", authServiceURL+"/verify", nil)
			req.Header.Set("Authorization", "Bearer "+cookie.Value)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)

			if err != nil || resp.StatusCode != http.StatusOK {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		}
	}

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			email := r.FormValue("email")
			password := r.FormValue("password")

			reqBody, _ := json.Marshal(map[string]string{
				"email":    email,
				"password": password,
			})

			resp, err := http.Post(authServiceURL+"/register", "application/json", bytes.NewBuffer(reqBody))
			if err != nil || resp.StatusCode != http.StatusCreated {
				render(w, r, "register.page.gohtml", TemplateData{FeedbackMessage: "Помилка реєстрації. Можливо, користувач вже існує."})
				return
			}

			render(w, r, "login.page.gohtml", TemplateData{SuccessMessage: "Реєстрація успішна! Тепер увійдіть."})
			return
		}
		render(w, r, "register.page.gohtml", TemplateData{})
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			email := r.FormValue("email")
			password := r.FormValue("password")

			reqBody, _ := json.Marshal(map[string]string{
				"email":    email,
				"password": password,
			})

			resp, err := http.Post(authServiceURL+"/login", "application/json", bytes.NewBuffer(reqBody))
			if err != nil || resp.StatusCode != http.StatusOK {
				render(w, r, "login.page.gohtml", TemplateData{FeedbackMessage: "Невірний email або пароль"})
				return
			}

			var apiResp map[string]string
			bodyBytes, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bodyBytes, &apiResp)

			http.SetCookie(w, &http.Cookie{
				Name:     "access_token",
				Value:    apiResp["access_token"],
				Path:     "/",
				HttpOnly: true,
			})

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		render(w, r, "login.page.gohtml", TemplateData{})
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "access_token", Value: "", MaxAge: -1, Path: "/"})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	http.HandleFunc("/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query().Get("query")
		var filteredDrugs []Drug
		for _, d := range catalog {
			if query == "" || strings.Contains(strings.ToLower(d.Name), strings.ToLower(query)) {
				filteredDrugs = append(filteredDrugs, d)
			}
		}
		render(w, r, "home.page.gohtml", TemplateData{Drugs: filteredDrugs, SearchQuery: query})
	}))

	http.HandleFunc("/drug/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/drug/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		var foundDrug *Drug
		for _, d := range catalog {
			if d.ID == id {
				foundDrug = &d
				break
			}
		}
		if foundDrug == nil {
			http.NotFound(w, r)
			return
		}
		render(w, r, "drug.page.gohtml", TemplateData{Drug: *foundDrug})
	}))

	http.HandleFunc("/about", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "about.page.gohtml", TemplateData{})
	}))

	http.HandleFunc("/feedback", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			name := r.FormValue("name")
			message := r.FormValue("message")

			data := TemplateData{
				FeedbackName:    name,
				FeedbackMessage: message,
			}
			render(w, r, "result.page.gohtml", data)
			return
		}
		render(w, r, "feedback.page.gohtml", TemplateData{})
	}))

	fmt.Println("Starting front end service on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func render(w http.ResponseWriter, r *http.Request, t string, data TemplateData) {
	data.CurrentDate = time.Now().Format("02.01.2006 15:04:05")

	// Автоматично перевіряємо, чи є кука, і ставимо прапорець IsAuthenticated
	_, err := r.Cookie("access_token")
	if err == nil {
		data.IsAuthenticated = true
	} else {
		data.IsAuthenticated = false
	}

	files := []string{
		"./cmd/web/templates/base.layout.gohtml",
		"./cmd/web/templates/header.partial.gohtml",
		"./cmd/web/templates/footer.partial.gohtml",
		fmt.Sprintf("./cmd/web/templates/%s", t),
	}
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "base.layout.gohtml", data)
}
