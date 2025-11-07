package wellknown

import (
	"fmt"
	"log"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/oauth2"
)

// Wellknown wraps a Pocketbase instance with wellknown-specific functionality
type Wellknown struct {
	*pocketbase.PocketBase
	config       *Config
	registry     *RouteRegistry
	oauthService *OAuthService
}

// ServerInfo contains information about the running server
type ServerInfo struct {
	Address string
	Port    int
	BaseURL string
	Routes  map[string][]RouteMetadata
}

// OAuthService holds OAuth-related configuration and state
type OAuthService struct {
	GoogleConfig *oauth2.Config
	db           *pocketbase.PocketBase
}

// New creates a standalone Wellknown app with default configuration
func New() (*Wellknown, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return NewWithConfig(cfg)
}

// NewWithConfig creates a Wellknown app with custom configuration
// This function performs fail-fast validation to catch errors early
func NewWithConfig(cfg *Config) (*Wellknown, error) {
	// Fail-fast: Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Fail-fast: Check data directory is writable
	if err := validateDataDir(cfg.Database.DataDir); err != nil {
		return nil, fmt.Errorf("data directory validation failed: %w", err)
	}

	// Create PocketBase app with configuration
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  cfg.Database.DataDir,
		HideStartBanner: true, // We show custom banner in main.go
	})

	// Initialize OAuth service (if Google OAuth is enabled)
	var oauthService *OAuthService
	if cfg.OAuth.Google.Enabled {
		// Fail-fast: Validate OAuth configuration
		if cfg.OAuth.Google.ClientID == "" || cfg.OAuth.Google.ClientSecret == "" {
			return nil, fmt.Errorf("OAuth Google enabled but ClientID or ClientSecret is empty")
		}
		oauthService = &OAuthService{
			GoogleConfig: cfg.OAuth.Google.ToOAuth2Config(),
			db:           app,
		}
		log.Println("✅ Google OAuth service initialized")
	}

	wk := &Wellknown{
		PocketBase:   app,
		config:       cfg,
		registry:     nil, // Set immediately in bindAppHooks
		oauthService: oauthService,
	}

	// Register all lifecycle hooks and initialize route registry
	bindAppHooks(wk)

	log.Printf("✅ Wellknown initialized (data dir: %s)", cfg.Database.DataDir)
	return wk, nil
}


// GetRegistry returns the route registry
func (wk *Wellknown) GetRegistry() *RouteRegistry {
	return wk.registry
}

// GetOAuthService returns the OAuth service
func (wk *Wellknown) GetOAuthService() *OAuthService {
	return wk.oauthService
}

// GetConfig returns the configuration
func (wk *Wellknown) GetConfig() *Config {
	return wk.config
}

// bindAppHooks registers all lifecycle hooks and routes
// NOTE: Route registration MUST happen inside OnServe() because PocketBase
// creates the router during the serve event. This is a PocketBase constraint.
func bindAppHooks(wk *Wellknown) {
	// Initialize route registry immediately (not waiting for server start)
	// This allows early inspection of available routes
	wk.registry = NewRouteRegistry()

	// Register system routes in the registry (metadata only, actual routing happens in OnServe)
	wk.registry.Register("System", "/_/", "GET", "PocketBase Admin UI", false)
	wk.registry.Register("System", "/api/health", "GET", "Health check endpoint", false)
	wk.registry.Register("System", "/api/collections", "GET", "List all collections", false)

	// Initialize templates
	if err := initTemplates(); err != nil {
		log.Printf("⚠️  Template loading failed: %v", err)
		log.Println("   Template-based routes may not work")
	}

	// Setup routes on server start (when router is available)
	wk.OnServe().BindFunc(func(e *core.ServeEvent) error {
		log.Println("🔗 Wellknown: Registering HTTP routes...")

		// NOTE: Collections are now managed via migrations in cmd/pb_migrations/
		// No runtime collection creation needed

		// Register domain routes (both registry metadata + actual HTTP handlers)
		RegisterOAuthRoutes(wk, e, wk.registry)
		RegisterCalendarRoutes(wk, e, wk.registry)
		RegisterBankingRoutes(wk, e, wk.registry)
		RegisterDemoRoutes(wk, e, wk.registry)

		// Register root HTML route (shows all endpoints)
		e.Router.GET("/", func(e *core.RequestEvent) error {
			return e.HTML(200, wk.registry.GenerateHTML())
		})

		// Register API index route (JSON) - Returns proper JSON
		e.Router.GET("/api/", func(e *core.RequestEvent) error {
			return e.JSON(200, wk.registry.GenerateJSON())
		})

		log.Println("✅ All HTTP routes registered successfully")
		return e.Next()
	})
}

// GetServerInfo returns information about the running server
func (wk *Wellknown) GetServerInfo() *ServerInfo {
	if wk.registry == nil {
		return nil
	}

	return &ServerInfo{
		Address: wk.config.Server.ServerAddress(),
		Port:    wk.config.Server.Port,
		BaseURL: wk.config.Server.ServerURL(),
		Routes:  wk.registry.GetRoutes(),
	}
}

// validateDataDir checks if the data directory exists and is writable
func validateDataDir(dir string) error {
	// Check if directory exists, create if not
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create data directory: %w", err)
	}

	// Check if directory is writable
	testFile := dir + "/.write_test"
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("data directory is not writable: %w", err)
	}
	os.Remove(testFile)

	return nil
}

