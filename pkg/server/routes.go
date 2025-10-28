package server

import (
	"fmt"
	"net/http"

	applecalendar "github.com/joeblew999/wellknown/pkg/apple/calendar"
	googlecalendar "github.com/joeblew999/wellknown/pkg/google/calendar"
	"github.com/joeblew999/wellknown/pkg/types"
)

// registerAllRoutes registers all HTTP routes with the server's mux and registry
// This is called during Server.New() initialization
func (s *Server) registerAllRoutes() {
	// Register calendar services
	s.registerCalendarServices()

	// Maps services (stubs for now)
	s.registerMapsRoutes()

	// Tools
	s.registerGCPSetupRoutes()

	// Homepage - shows all available services
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.handleHome(w, r)
			return
		}
		http.NotFound(w, r)
	})
}

// registerCalendarServices registers all calendar services
func (s *Server) registerCalendarServices() {
	// Define all calendar services in ONE place
	services := []struct {
		Platform     string
		AppType      string
		Title        string
		SuccessLabel string
		GenerateURL  CalendarURLGenerator
		ExtraRoutes  map[string]http.HandlerFunc
	}{
		{
			Platform:     "google",
			AppType:      "calendar",
			Title:        "Google Calendar",
			SuccessLabel: "URL",
			GenerateURL:  googlecalendar.GenerateURL,
			ExtraRoutes:  nil,
		},
		{
			Platform:     "apple",
			AppType:      "calendar",
			Title:        "Apple Calendar",
			SuccessLabel: "Download Link",
			GenerateURL:  applecalendar.GenerateDownloadURL,
			ExtraRoutes: map[string]http.HandlerFunc{
				"/apple/calendar/download": s.handleAppleCalendarDownload,
			},
		},
	}

	for _, svc := range services {
		// Create calendar handler
		handler := s.makeGenericCalendarHandler(CalendarConfig{
			Platform:     svc.Platform,
			AppType:      svc.AppType,
			SuccessLabel: svc.SuccessLabel,
			GenerateURL:  svc.GenerateURL,
		})

		// Load showcase examples from JSON
		examplesPath := fmt.Sprintf("pkg/%s/%s/data-examples.json", svc.Platform, svc.AppType)
		examples, _ := types.LoadExamples(examplesPath)
		examplesHandler := s.makeExamplesHandler(svc.Platform, svc.AppType, examples)

		// Register main handler
		mainPath := "/" + svc.Platform + "/" + svc.AppType
		s.mux.HandleFunc(mainPath, handler)

		// Register showcase handler
		showcasePath := mainPath + "/showcase"
		s.mux.HandleFunc(showcasePath, examplesHandler)

		// Register extra routes (if any)
		for path, extraHandler := range svc.ExtraRoutes {
			s.mux.HandleFunc(path, extraHandler)
		}

		// Register for navigation
		s.registry.Register(ServiceConfig{
			Platform:    svc.Platform,
			AppType:     svc.AppType,
			Title:       svc.Title,
			HasCustom:   true,
			HasExamples: true,
		})
	}
}
