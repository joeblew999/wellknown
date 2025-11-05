package wellknown

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// Wellknown wraps a Pocketbase instance with wellknown-specific functionality
type Wellknown struct {
	*pocketbase.PocketBase
}

// New creates a standalone Wellknown app with embedded Pocketbase
// Uses .data/pb as the default data directory for multi-service architecture
func New() *Wellknown {
	return NewWithConfig(pocketbase.Config{
		DefaultDataDir: ".data/pb",
	})
}

// NewWithConfig creates a Wellknown app with custom PocketBase configuration
func NewWithConfig(config pocketbase.Config) *Wellknown {
	app := pocketbase.NewWithConfig(config)
	wk := &Wellknown{app}
	bindAppHooks(wk)
	return wk
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

		// Create route registry for auto-documentation
		registry := NewRouteRegistry()

		// Register core system routes
		registry.Register("System", "/_/", "GET", "PocketBase Admin UI", false)
		registry.Register("System", "/api/health", "GET", "Health check endpoint", false)
		registry.Register("System", "/api/collections", "GET", "List all collections", false)

		// Register domain routes
		RegisterOAuthRoutes(wk, e, registry)
		RegisterCalendarRoutes(wk, e, registry)
		RegisterBankingRoutes(wk, e, registry)

		// Register root HTML route (shows all endpoints)
		e.Router.GET("/", func(e *core.RequestEvent) error {
			return e.HTML(200, registry.GenerateHTML())
		})

		// Register API index route (JSON) - Returns proper JSON
		e.Router.GET("/api/", func(e *core.RequestEvent) error {
			return e.JSON(200, registry.GenerateJSON())
		})

		log.Println("✅ All routes registered successfully")
		return e.Next()
	})
}

