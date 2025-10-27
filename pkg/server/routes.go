package server

import "net/http"

// RegisterRoutes registers all HTTP routes with the given mux.
// This is exported for testing and can be used with custom mux implementations.
// Each feature registers its own routes via self-registration functions.
func RegisterRoutes(mux *http.ServeMux) {
	// Calendar services
	RegisterGoogleCalendarRoutes(mux)
	RegisterAppleCalendarRoutes(mux)

	// Maps services (stubs for now)
	RegisterMapsRoutes(mux)

	// Tools
	RegisterGCPSetupRoutes(mux)

	// Homepage - shows all available services
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			Home(w, r)
			return
		}
		http.NotFound(w, r)
	})
}
