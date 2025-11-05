package wellknown

import (
	"log"

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
func NewWithConfig(cfg *Config) (*Wellknown, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: cfg.Database.DataDir,
	})

	// Initialize OAuth service
	var oauthService *OAuthService
	if cfg.OAuth.Google.Enabled {
		oauthService = &OAuthService{
			GoogleConfig: cfg.OAuth.Google.ToOAuth2Config(),
			db:           app,
		}
	}

	wk := &Wellknown{
		PocketBase:   app,
		config:       cfg,
		registry:     nil, // Will be set in bindAppHooks
		oauthService: oauthService,
	}

	bindAppHooks(wk)
	return wk, nil
}

// NewWithApp attaches wellknown functionality to an existing PocketBase app
// Deprecated: Use NewWithConfig instead
func NewWithApp(app *pocketbase.PocketBase) *Wellknown {
	cfg, _ := LoadConfig()

	var oauthService *OAuthService
	if cfg.OAuth.Google.Enabled {
		oauthService = &OAuthService{
			GoogleConfig: cfg.OAuth.Google.ToOAuth2Config(),
			db:           app,
		}
	}

	wk := &Wellknown{
		PocketBase:   app,
		config:       cfg,
		registry:     nil,
		oauthService: oauthService,
	}
	bindAppHooks(wk)
	return wk
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
func bindAppHooks(wk *Wellknown) {
	// Setup collections and routes on server start
	wk.OnServe().BindFunc(func(e *core.ServeEvent) error {
		log.Println("🔗 Wellknown: Initializing routes...")

		// NOTE: Collections are now managed via migrations in cmd/pb_migrations/
		// No runtime collection creation needed

		// Create route registry for auto-documentation
		registry := NewRouteRegistry()
		wk.registry = registry // Store for later access

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

