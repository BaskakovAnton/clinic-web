package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type PageData struct {
	Title   string
	Content string
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(
			filepath.Join("templates", "base.html"),
			filepath.Join("templates", "blocks", "header.html"),
			filepath.Join("templates", "blocks", "content.html"),
			filepath.Join("templates", "blocks", "footer.html"),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := PageData{
			Title:   "Главная",
			Content: "Добро пожаловать в клинику!",
		}
		err = tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
