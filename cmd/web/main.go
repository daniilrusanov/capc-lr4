package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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
}

// catalog - наш імпровізований масив даних (база даних аптеки)
var catalog = []Drug{
	{ID: 1, Name: "Парацетамол", Manufacturer: "Дарниця", Price: 35.50},
	{ID: 2, Name: "Аспірин", Manufacturer: "Bayer", Price: 120.00},
	{ID: 3, Name: "Ібупрофен", Manufacturer: "Фармак", Price: 65.00},
	{ID: 4, Name: "Но-шпа", Manufacturer: "Sanofi", Price: 150.00},
	{ID: 5, Name: "Цитрамон", Manufacturer: "Здоров'я", Price: 25.00},
}

func main() {
	// Маршрут 1: Головна сторінка
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		render(w, "home.page.gohtml", TemplateData{Drugs: filteredDrugs, SearchQuery: query})
	})

	// Маршрут 2: Деталі препарату
	http.HandleFunc("/drug/", func(w http.ResponseWriter, r *http.Request) {
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
		render(w, "drug.page.gohtml", TemplateData{Drug: *foundDrug})
	})

	// Маршрут 3: Про нас
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		render(w, "about.page.gohtml", TemplateData{})
	})

	// МАРШРУТ ДЛЯ ФОРМИ (POST)
	http.HandleFunc("/feedback", func(w http.ResponseWriter, r *http.Request) {
		// Якщо метод POST - обробляємо дані форми
		if r.Method == http.MethodPost {
			// Парсимо дані форми
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Помилка обробки форми", http.StatusBadRequest)
				return
			}

			// Зчитуємо значення полів
			name := r.FormValue("name")
			message := r.FormValue("message")

			// Формуємо дані для сторінки результату
			data := TemplateData{
				FeedbackName:    name,
				FeedbackMessage: message,
			}

			// Рендеримо сторінку підтвердження
			render(w, "result.page.gohtml", data)
			return
		}

		// Якщо метод GET - просто показуємо порожню форму
		render(w, "feedback.page.gohtml", TemplateData{})
	})

	fmt.Println("Starting front end service on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic(err)
	}
}

func render(w http.ResponseWriter, t string, data TemplateData) {
	data.CurrentDate = time.Now().Format("02.01.2006 15:04:05")
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
	err = tmpl.ExecuteTemplate(w, "base.layout.gohtml", data)
	if err != nil {
		log.Println("Template error:", err)
	}
}
