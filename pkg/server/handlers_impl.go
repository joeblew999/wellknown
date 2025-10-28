package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// handleHome handles the homepage showing all available services
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Render homepage with all services - use the SINGLE render method
	s.render(w, r, PageData{
		Platform:     "",
		AppType:      "",
		CurrentPage:  "home",
		TemplateName: "home",
	})
}

// makeStubHandler creates a stub handler for unimplemented services
func (s *Server) makeStubHandler(platform, appType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s (stub)", r.Method, r.URL.Path)

		currentPage := "custom"
		if strings.HasSuffix(r.URL.Path, "/showcase") {
			currentPage = "showcase"
		}

		// Use the SINGLE render method
		s.render(w, r, PageData{
			Platform:    platform,
			AppType:     appType,
			CurrentPage: currentPage,
			IsStub:      true,
		})
	}
}

// registerMapsRoutes registers all Maps routes (stubs for now)
func (s *Server) registerMapsRoutes() {
	// Google Maps
	s.mux.HandleFunc("/google/maps", s.makeStubHandler("google", "maps"))
	s.mux.HandleFunc("/google/maps/showcase", s.makeStubHandler("google", "maps"))
	s.registry.Register(ServiceConfig{
		Platform:    "google",
		AppType:     "maps",
		Title:       "Google Maps",
		HasCustom:   true,
		HasExamples: true,
	})

	// Apple Maps
	s.mux.HandleFunc("/apple/maps", s.makeStubHandler("apple", "maps"))
	s.mux.HandleFunc("/apple/maps/showcase", s.makeStubHandler("apple", "maps"))
	s.registry.Register(ServiceConfig{
		Platform:    "apple",
		AppType:     "maps",
		Title:       "Apple Maps",
		HasCustom:   true,
		HasExamples: true,
	})
}

// handleAppleCalendarDownload serves .ics file for download
// This is the CORRECT way to handle Apple Calendar on iOS/macOS
// Safari cannot handle data:text/calendar URIs - it requires actual file downloads
func (s *Server) handleAppleCalendarDownload(w http.ResponseWriter, r *http.Request) {
	eventParam := r.URL.Query().Get("event")
	if eventParam == "" {
		http.Error(w, "Missing event parameter", http.StatusBadRequest)
		return
	}

	icsContent, err := base64.URLEncoding.DecodeString(eventParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid event data: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline; filename=\"event.ics\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(icsContent)))
	w.Write(icsContent)
}
