package main

import (
	"log"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
        log.Println("Template execution error:", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func bachataHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "bachata_index.html", nil)
}

func salsaHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "salsa_index.html", nil)
}

func initMiscHandlers() {
	http.HandleFunc("/htmltools", HomeHandler)

	http.HandleFunc("/htmltools/bachata", bachataHandler)
	http.HandleFunc("/htmltools/salsa", salsaHandler)
}
