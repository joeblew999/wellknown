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
	// Setup collections on app start
	wk.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		log.Println("🔗 Wellknown: Initializing collections and routes...")
		return setupCollections(wk)
	})

	// Register Google OAuth routes
	wk.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		return registerOAuthRoutes(wk, e)
	})

	// Register Calendar API routes
	wk.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		return registerCalendarRoutes(wk, e)
	})
}
