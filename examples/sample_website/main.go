package main

import (
	"fmt"
	"net/http"
	"text/template"
)

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}
func reportsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "reports.html", nil)
}
func readerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "ppcl_reader.html", nil)
}

func main() {
	portnum := "8080"

	// serve everything in the assets folder
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// // serve the home page
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/reports", reportsHandler)
	http.HandleFunc("/ppcl-reader", readerHandler)

	// start the server
	listenloc := "localhost:" + portnum // TODO: revert to ":portnum" for production (no localhost)
	fmt.Printf("Listening on %s\n", listenloc)
	http.ListenAndServe(listenloc, nil)
}
