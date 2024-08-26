package main

import (
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "main_home.html", nil)
}

func readerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "ppcl_reader.html", nil)
}

func n4ColourSchemeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "n4_sample_graphic.html", nil)
}

func initMiscHandlers() {
	http.HandleFunc("/htmltools", HomeHandler)

	http.HandleFunc("/htmltools/ppcl-reader", readerHandler) // TODO: Not done yet
	http.HandleFunc("/htmltools/N4-colour-schemes", n4ColourSchemeHandler)
}
