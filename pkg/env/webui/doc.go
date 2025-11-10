// Package webui provides a reusable web GUI for environment variable inspection and debugging.
//
// # Overview
//
// The webui package provides ready-to-use HTTP handlers that integrate with any env.Registry
// to provide a clean, professional web interface for viewing and debugging environment variables.
//
// # Features
//
//   - Clean HTML interface styled with Pico CSS (classless framework)
//   - Dark mode support (auto-detects system preference)
//   - Table-based layout with grouped variables and status icons
//   - Dual-format support (HTML and JSON)
//   - Secret value hiding (shows ••••••••)
//   - Environment detection (local, docker, fly.io, kubernetes)
//   - Health check endpoint with uptime and Go runtime info
//   - Works with ANY env.Registry (not hardcoded)
//   - Minimal CSS footprint (~20KB from CDN)
//   - Semantic HTML for accessibility
//
// # Quick Start
//
// Import the package and register routes with your HTTP mux:
//
//	import (
//	    "net/http"
//	    "github.com/joeblew999/wellknown/pkg/env"
//	    "github.com/joeblew999/wellknown/pkg/env/webui"
//	)
//
//	func main() {
//	    // Create your registry
//	    registry := env.NewRegistry([]env.EnvVar{
//	        {Name: "API_KEY", Secret: true, Required: true},
//	        {Name: "PORT", Default: "8080"},
//	    })
//
//	    // Create webui handler
//	    handler := webui.NewHandler(registry)
//
//	    // Register routes
//	    mux := http.NewServeMux()
//	    handler.RegisterRoutes(mux)
//
//	    // Start server
//	    http.ListenAndServe(":8080", mux)
//	}
//
// # Available Endpoints
//
//   - GET /env - Environment variables view (HTML by default, ?format=json for JSON)
//   - GET /health - Health check with environment detection and uptime
//
// # HTML View Features
//
// The /env endpoint provides a beautiful HTML interface with:
//   - Grouped variables (by registry Group field)
//   - Color-coded badges (SECRET, REQUIRED, CONFIGURED)
//   - Secret value hiding (shows ••••••••)
//   - Stats cards (total vars, groups, configured, secrets)
//   - Responsive design with gradient styling
//   - One-click JSON view
//
// # JSON View
//
// Access JSON format by:
//   - Query parameter: /env?format=json
//   - Accept header: Accept: application/json
//
// Example JSON response:
//
//	{
//	  "total_variables": 5,
//	  "environment": "local",
//	  "groups": {
//	    "Server": [...],
//	    "APIs": [...]
//	  },
//	  "variables": {
//	    "api_key_configured": true,
//	    "port_configured": true
//	  }
//	}
//
// # Environment Detection
//
// The webui automatically detects the runtime environment:
//   - fly.io - Detected via FLY_APP_NAME environment variable
//   - docker - Detected via /.dockerenv file
//   - kubernetes - Detected via KUBERNETES_SERVICE_HOST
//   - local - Default when none of the above match
//
// # Integration Example
//
// Integrate with existing server (like the example app):
//
//	func main() {
//	    mux := http.NewServeMux()
//
//	    // Register webui routes
//	    webuiHandler := webui.NewHandler(AppRegistry)
//	    webuiHandler.RegisterRoutes(mux)
//
//	    // Add your app-specific routes
//	    mux.HandleFunc("/", handleHome)
//	    mux.HandleFunc("/api", handleAPI)
//
//	    http.ListenAndServe(":8080", mux)
//	}
//
// # Security
//
// Secret values are automatically hidden in both HTML and JSON views.
// They are replaced with ••••••••, but their configuration status is shown.
package webui
