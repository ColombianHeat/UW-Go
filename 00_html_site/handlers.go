package main

import (
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "main_home.html", nil)
}

func reportsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "reports_home.html", nil)
}

func chws_eda_idx_Handler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "chws_eda_idx.html", nil)
}

func chws_eda_rep_Handler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "chws_eda_rep.html", nil)
}


func readerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "ppcl_reader.html", nil)
}

func n4ColourSchemeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "n4_sample_graphic.html", nil)
}

func initHandlers() {
	http.HandleFunc("/", homeHandler)

	http.HandleFunc("/reports", reportsHandler)
	http.HandleFunc("/reports/CHWS-EDA-index", chws_eda_idx_Handler)
	http.HandleFunc("/reports/CHWS-EDA-report", chws_eda_rep_Handler) // TODO: Need to figure out how to show images

	http.HandleFunc("/ppcl-reader", readerHandler) // TODO: Not done yet
	http.HandleFunc("/N4-colour-schemes", n4ColourSchemeHandler)
}
