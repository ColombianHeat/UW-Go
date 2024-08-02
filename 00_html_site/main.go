package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var tmpl *template.Template
// var repTmpl *template.Template

func ParseTemplates() *template.Template {
    templ := template.New("")
	tmpDir := "./templates"
    err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
        if strings.Contains(path, ".html") {
            _, err = templ.ParseFiles(path)
            if err != nil {
                log.Println(err)
            }
        }

        return err
    })

    if err != nil {
        panic(err)
    }

    return templ
}

func init() {
	tmpl = ParseTemplates()
	// tmpl = template.Must(template.ParseGlob("templates/*/*.html"))
	// repTmpl = template.Must(template.ParseGlob("templates/energy_reports/*/*.html"))
}

func main() {
	portnum := "8080"

	// serve everything in the assets folder
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	initHandlers()

	// start the server
	listenloc := "localhost:" + portnum // TODO: revert to ":portnum" for production (no localhost)
	fmt.Printf("Listening on %s\n", listenloc)
	http.ListenAndServe(listenloc, nil)
}
