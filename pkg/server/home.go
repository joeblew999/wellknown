package server

import (
	"log"
	"net/http"
)

// Home handles the homepage showing all available services
func Home(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Render homepage with all services
	err := Templates.ExecuteTemplate(w, "base", PageData{
		Platform:     "",
		AppType:      "",
		CurrentPage:  "home",
		TemplateName: "home",
		LocalURL:     LocalURL,
		MobileURL:    MobileURL,
		Navigation:   GetNavigation("/"),
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}
