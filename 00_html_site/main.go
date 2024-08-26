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
        if strings.Contains(path, ".html") && !strings.Contains(path, "html_") {
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

	// statically serve everything in the assets folder
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/htmltools/assets/", http.StripPrefix("/htmltools/assets/", fs))

	// dynamically serve .png images in ./templates/energy_reports/*/img/ folders
	http.HandleFunc("/htmltools/reports/", func(w http.ResponseWriter, r *http.Request) {
		imageName := strings.TrimPrefix(r.URL.Path, "/htmltools/reports/img/")
		
		matches, err := filepath.Glob("./templates/energy_reports/*/img/" + imageName)
		if err != nil || len(matches) == 0 {
			http.NotFound(w, r)
			return
		}
		
		http.ServeFile(w, r, matches[0])
		})

	initMiscHandlers()
	initReportHandlers()

	// start the server
	listenloc := ":" + portnum // NOTE: for production
	// listenloc := "localhost:" + portnum // NOTE: for local testing
	fmt.Printf("Listening on %s\n", listenloc)
	http.ListenAndServe(listenloc, nil)
}
