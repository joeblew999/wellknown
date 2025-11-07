package wellknown

import (
	"log"
	"net/http"

	"github.com/joeblew999/wellknown/pkg/server"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterDemoRoutes registers the standalone server routes under /demo/* prefix
// These routes provide testing and demonstration UI for various features
func RegisterDemoRoutes(wk *Wellknown, e *core.ServeEvent, registry *RouteRegistry) {
	log.Println("   📋 Registering demo/testing routes...")

	// Create a standalone server instance (we'll use its handlers)
	// Note: We pass "8090" as port since we're embedding in PocketBase
	standaloneSvr, err := server.New("8090")
	if err != nil {
		log.Printf("⚠️  Failed to initialize demo routes: %v", err)
		return
	}

	// Set URL prefix so links in templates include /demo prefix
	standaloneSvr.URLPrefix = "/demo"

	// Get the embedded http.ServeMux from standalone server
	// We'll extract its handlers and mount them under /demo/* in PocketBase
	mux := standaloneSvr.GetMux()

	// Register demo routes with /demo prefix
	// All standalone routes are public (no auth required)

	// Home page - /demo/
	e.Router.GET("/demo/", adaptHandler(mux, "/", "Demo home page"))
	registry.Register("Demo", "/demo/", "GET", "Demo & testing home page", false)

	// Google Calendar routes
	e.Router.GET("/demo/google/calendar", adaptHandler(mux, "/google/calendar", "Google Calendar form"))
	e.Router.POST("/demo/google/calendar", adaptHandler(mux, "/google/calendar", "Google Calendar form POST"))
	e.Router.GET("/demo/google/calendar/examples", adaptHandler(mux, "/google/calendar/examples", "Google Calendar examples"))

	registry.Register("Demo", "/demo/google/calendar", "GET, POST", "Google Calendar URL generator", false)
	registry.Register("Demo", "/demo/google/calendar/examples", "GET", "Google Calendar examples", false)

	// Apple Calendar routes
	e.Router.GET("/demo/apple/calendar", adaptHandler(mux, "/apple/calendar", "Apple Calendar form"))
	e.Router.POST("/demo/apple/calendar", adaptHandler(mux, "/apple/calendar", "Apple Calendar form POST"))
	e.Router.GET("/demo/apple/calendar/examples", adaptHandler(mux, "/apple/calendar/examples", "Apple Calendar examples"))
	e.Router.GET("/demo/apple/calendar/download", adaptHandler(mux, "/apple/calendar/download", "Apple Calendar ICS download"))

	registry.Register("Demo", "/demo/apple/calendar", "GET, POST", "Apple Calendar ICS generator", false)
	registry.Register("Demo", "/demo/apple/calendar/examples", "GET", "Apple Calendar examples", false)
	registry.Register("Demo", "/demo/apple/calendar/download", "GET", "Download Apple Calendar .ics file", false)

	// Google Maps routes (stubs)
	e.Router.GET("/demo/google/maps", adaptHandler(mux, "/google/maps", "Google Maps stub"))
	registry.Register("Demo", "/demo/google/maps", "GET", "Google Maps URL generator (stub)", false)

	// Apple Maps routes (stubs)
	e.Router.GET("/demo/apple/maps", adaptHandler(mux, "/apple/maps", "Apple Maps stub"))
	registry.Register("Demo", "/demo/apple/maps", "GET", "Apple Maps URL generator (stub)", false)

	// GCP Setup Wizard routes
	e.Router.GET("/demo/tools/gcp-setup", adaptHandler(mux, "/tools/gcp-setup", "GCP setup wizard"))
	registry.Register("Demo", "/demo/tools/gcp-setup", "GET", "GCP OAuth setup wizard", false)

	// GCP Setup API endpoints (all POST requests)
	gcpAPIRoutes := []string{
		"/demo/api/gcp-setup/create-project",
		"/demo/api/gcp-setup/enable-apis",
		"/demo/api/gcp-setup/create-credentials",
		"/demo/api/gcp-setup/configure-consent",
		"/demo/api/gcp-setup/save-env",
		"/demo/api/gcp-setup/status",
	}

	for _, route := range gcpAPIRoutes {
		// Map /demo/api/gcp-setup/* to /api/gcp-setup/*
		originalRoute := "/api" + route[10:] // Remove /demo prefix
		e.Router.POST(route, adaptHandler(mux, originalRoute, "GCP setup API"))
	}

	registry.Register("Demo", "/demo/api/gcp-setup/*", "POST", "GCP setup API endpoints", false)

	log.Println("   ✅ Demo routes registered successfully")
}

// adaptHandler adapts a stdlib http.Handler from the standalone server's mux
// to work with PocketBase's echo-based routing system
func adaptHandler(mux *http.ServeMux, originalPath string, description string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// We need to route based on the stripped path (without /demo prefix)
		// but keep the original request URL intact for the handler

		// Create a custom ResponseWriter that can intercept the response
		// and a modified request that the mux can route correctly
		routingReq, err := http.NewRequest(e.Request.Method, originalPath, e.Request.Body)
		if err != nil {
			return err
		}

		// Copy headers from original request
		routingReq.Header = e.Request.Header

		// Copy form values if present
		if e.Request.Form != nil {
			routingReq.Form = e.Request.Form
		}
		if e.Request.PostForm != nil {
			routingReq.PostForm = e.Request.PostForm
		}

		// Set the context to the original request's context
		routingReq = routingReq.WithContext(e.Request.Context())

		// Serve using the standalone mux with our routing request
		mux.ServeHTTP(e.Response, routingReq)

		return nil
	}
}
