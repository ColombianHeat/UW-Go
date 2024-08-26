package main

import (
	"net/http"
)

func reportsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "reports_home.html", nil)
}

func chws_eda_idx_Handler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "chws_eda_idx.html", nil)
}

func chws_eda_rep_Handler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "chws_eda_rep.html", nil)
}

func cmh_kitchen_exhaust_idx_Handler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "cmh_kitchen_exhaust_idx.html", nil)
}

func cmh_kitchen_exhaust_rep_Handler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "cmh_kitchen_exhaust_rep.html", nil)
}

func initReportHandlers() {

	http.HandleFunc("/htmltools/reports", reportsHandler)

	http.HandleFunc("/htmltools/reports/CHWS-EDA-index", chws_eda_idx_Handler)
	http.HandleFunc("/htmltools/reports/CHWS-EDA-report", chws_eda_rep_Handler)

	http.HandleFunc("/htmltools/reports/CMH-kitchen-exhaust-index", cmh_kitchen_exhaust_idx_Handler)
	http.HandleFunc("/htmltools/reports/CMH-kitchen-exhaust-report", cmh_kitchen_exhaust_rep_Handler)
}
