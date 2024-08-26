package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
)

// Mock the template execution
var testTemplate = template.Must(template.New("main_home.html").Parse(`<html><body>Home Page</body></html>`))

func TestHomeHandler(t *testing.T) {
	// Replace the global tmpl with the test version
	tmpl = testTemplate

	// Create a test request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a new recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HomeHandler)

	// Serve the test HTTP request
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "<html><body>Home Page</body></html>"
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}