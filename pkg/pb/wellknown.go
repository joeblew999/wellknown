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
		log.Printf("🔍 DEBUG: Router object in wellknown.go: %p", e.Router)

		// NOTE: Collections are now managed via migrations in cmd/pb_migrations/
		// No runtime collection creation needed

		// Define available endpoints (single source of truth)
		endpoints := map[string]string{
			"health":              "/api/health",
			"collections":         "/api/collections",
			"oauth_google":        "/auth/google",
			"oauth_callback":      "/auth/google/callback",
			"oauth_logout":        "/auth/logout",
			"oauth_status":        "/auth/status",
			"calendar_list":       "/api/calendar/events (GET, authenticated)",
			"calendar_create":     "/api/calendar/events (POST, authenticated)",
			"banking_accounts":    "/api/banking/accounts (GET/POST)",
			"banking_account":     "/api/banking/accounts/:id (GET)",
			"banking_transactions": "/api/banking/accounts/:id/transactions (GET)",
			"banking_create_tx":   "/api/banking/transactions (POST)",
			"admin_ui":            "/_/",
		}

		// Register root HTML route (dynamic)
		e.Router.GET("/", func(e *core.RequestEvent) error {
			return e.HTML(200, generateIndexHTML(endpoints))
		})

		// Register Google OAuth routes
		if err := registerOAuthRoutes(wk, e); err != nil {
			return err
		}

		// Register Calendar API routes
		if err := registerCalendarRoutes(wk, e); err != nil {
			return err
		}

		// Register Banking API routes INLINE (external function doesn't work)
		registerBankingRoutesInline(wk, e)

		// Register API index route (JSON) - MUST be last to avoid shadowing specific routes
		e.Router.GET("/api/", func(e *core.RequestEvent) error {
			return e.JSON(200, map[string]any{
				"message":   "Wellknown PocketBase API",
				"version":   "1.0.0",
				"endpoints": endpoints,
			})
		})

		return e.Next()
	})
}

// registerBankingRoutesInline registers banking routes directly (inline approach)
func registerBankingRoutesInline(wk *Wellknown, e *core.ServeEvent) {
	log.Println("🔗 Registering banking routes inline...")

	// List accounts
	e.Router.GET("/api/banking/accounts", func(c *core.RequestEvent) error {
		userID := c.Request.URL.Query().Get("user_id")
		if userID == "" {
			return c.JSON(400, map[string]any{"error": "user_id required"})
		}
		records, err := wk.FindRecordsByFilter("accounts", "user_id = {:userID}", "-created", 100, 0, map[string]any{"userID": userID})
		if err != nil {
			return c.JSON(500, map[string]any{"error": err.Error()})
		}
		return c.JSON(200, map[string]any{"accounts": records})
	})

	// Create account
	e.Router.POST("/api/banking/accounts", func(c *core.RequestEvent) error {
		data := &struct {
			UserID        string  `json:"user_id"`
			AccountNumber string  `json:"account_number"`
			AccountName   string  `json:"account_name"`
			AccountType   string  `json:"account_type"`
			Balance       float64 `json:"balance"`
			Currency      string  `json:"currency"`
			IsActive      bool    `json:"is_active"`
		}{}

		if err := c.BindBody(data); err != nil {
			return c.JSON(400, map[string]any{"error": "Invalid request body"})
		}

		collection, err := wk.FindCollectionByNameOrId("accounts")
		if err != nil {
			return c.JSON(500, map[string]any{"error": "Collection not found"})
		}

		record := core.NewRecord(collection)
		record.Set("user_id", data.UserID)
		record.Set("account_number", data.AccountNumber)
		record.Set("account_name", data.AccountName)
		record.Set("account_type", data.AccountType)
		record.Set("balance", data.Balance)
		record.Set("currency", data.Currency)
		record.Set("is_active", data.IsActive)

		if err := wk.Save(record); err != nil {
			return c.JSON(500, map[string]any{"error": err.Error()})
		}

		return c.JSON(201, record)
	})

	log.Println("✅ Banking routes registered inline")
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
