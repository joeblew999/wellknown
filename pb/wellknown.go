package wellknown

import (
	"fmt"
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Wellknown wraps a Pocketbase instance with wellknown-specific functionality
type Wellknown struct {
	*pocketbase.PocketBase
}

// New creates a standalone Wellknown app with embedded Pocketbase
func New() *Wellknown {
	return NewWithApp(pocketbase.New())
}

// NewWithApp attaches wellknown functionality to an existing PocketBase app
// This allows other projects to import our Google Calendar OAuth integration
func NewWithApp(app *pocketbase.PocketBase) *Wellknown {
	wk := &Wellknown{app}
	bindAppHooks(wk)
	return wk
}

// bindAppHooks registers all lifecycle hooks and routes
func bindAppHooks(wk *Wellknown) {
	// Setup collections and routes on server start
	wk.OnServe().BindFunc(func(e *core.ServeEvent) error {
		log.Println("🔗 Wellknown: Initializing routes...")

		// NOTE: Collections are now managed via migrations in cmd/pb_migrations/
		// No runtime collection creation needed

		// Define available endpoints (single source of truth)
		endpoints := map[string]string{
			"health":          "/api/health",
			"collections":     "/api/collections",
			"oauth_google":    "/auth/google",
			"oauth_callback":  "/auth/google/callback",
			"oauth_logout":    "/auth/logout",
			"oauth_status":    "/auth/status",
			"calendar_list":   "/api/calendar/events (GET, authenticated)",
			"calendar_create": "/api/calendar/events (POST, authenticated)",
			"admin_ui":        "/_/",
		}

		// Register root HTML route (dynamic)
		e.Router.GET("/", func(e *core.RequestEvent) error {
			return e.HTML(200, generateIndexHTML(endpoints))
		})

		// Register API index route (JSON)
		e.Router.GET("/api/", func(e *core.RequestEvent) error {
			return e.JSON(200, map[string]any{
				"message":   "Wellknown PocketBase API",
				"version":   "1.0.0",
				"endpoints": endpoints,
			})
		})

		// Register Google OAuth routes
		if err := registerOAuthRoutes(wk, e); err != nil {
			return err
		}

		// Register Calendar API routes
		if err := registerCalendarRoutes(wk, e); err != nil {
			return err
		}

		return e.Next()
	})
}

// generateIndexHTML creates a dynamic HTML page listing all endpoints
func generateIndexHTML(endpoints map[string]string) string {
	var links strings.Builder
	for name, path := range endpoints {
		links.WriteString(fmt.Sprintf(
			"\n        <li><strong>%s:</strong> <a href=\"%s\"><span class=\"endpoint\">%s</span></a></li>",
			name, path, path,
		))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Wellknown PocketBase</title>
    <style>
        body { font-family: system-ui; max-width: 800px; margin: 40px auto; padding: 0 20px; }
        h1 { color: #333; }
        ul { list-style: none; padding: 0; }
        li { margin: 10px 0; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .endpoint { font-family: monospace; background: #f5f5f5; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🔐 Wellknown PocketBase API</h1>
    <p>Available endpoints:</p>
    <ul>%s
    </ul>
    <p><small>💡 Tip: Visit <a href="/api/">/api/</a> for JSON version</small></p>
</body>
</html>`, links.String())
}
