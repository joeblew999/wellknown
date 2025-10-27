package server

import "net/http"

// RegisterRoutes registers all HTTP routes with the given mux.
// This is exported for testing and can be used with custom mux implementations.
func RegisterRoutes(mux *http.ServeMux) {
	// Google Calendar - UI Schema with validation
	mux.HandleFunc("/google/calendar", GoogleCalendar)
	mux.HandleFunc("/google/calendar/showcase", GoogleCalendarShowcase)
	registerService(ServiceConfig{
		Platform:    "google",
		AppType:     "calendar",
		Title:       "Google Calendar",
		HasCustom:   true,
		HasShowcase: true,
	})

	// Apple Calendar - UI Schema with validation
	mux.HandleFunc("/apple/calendar", AppleCalendar)
	mux.HandleFunc("/apple/calendar/showcase", AppleCalendarShowcase)
	mux.HandleFunc("/apple/calendar/download", AppleCalendarDownload)
	registerService(ServiceConfig{
		Platform:    "apple",
		AppType:     "calendar",
		Title:       "Apple Calendar",
		HasCustom:   true,
		HasShowcase: true,
	})

	// Stub services (coming soon)
	mux.HandleFunc("/google/maps", Stub("google", "maps"))
	mux.HandleFunc("/google/maps/showcase", Stub("google", "maps"))
	registerService(ServiceConfig{
		Platform:    "google",
		AppType:     "maps",
		Title:       "Google Maps",
		HasCustom:   true,
		HasShowcase: true,
	})

	mux.HandleFunc("/apple/maps", Stub("apple", "maps"))
	mux.HandleFunc("/apple/maps/showcase", Stub("apple", "maps"))
	registerService(ServiceConfig{
		Platform:    "apple",
		AppType:     "maps",
		Title:       "Apple Maps",
		HasCustom:   true,
		HasShowcase: true,
	})

	// Tools - GCP OAuth Setup
	mux.HandleFunc("/tools/gcp-setup", handleGCPSetup)
	mux.HandleFunc("/api/gcp-setup/status", handleGCPSetupStatus)
	mux.HandleFunc("/api/gcp-setup/save-project", handleGCPSaveProject)
	mux.HandleFunc("/api/gcp-setup/save-creds", handleGCPSaveCreds)
	mux.HandleFunc("/api/gcp-setup/reset", handleGCPReset)

	// Homepage - shows all available services
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			Home(w, r)
			return
		}
		http.NotFound(w, r)
	})
}
