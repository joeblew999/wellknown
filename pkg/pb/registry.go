package wellknown

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// RouteMetadata contains information about a registered route
type RouteMetadata struct {
	Path        string   `json:"path"`
	Methods     []string `json:"methods"`
	Description string   `json:"description"`
	AuthRequired bool     `json:"auth_required"`
	Domain      string   `json:"domain"`
}

// RouteRegistry maintains a registry of all API routes with their metadata
type RouteRegistry struct {
	routes []RouteMetadata
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make([]RouteMetadata, 0),
	}
}

// Register adds a route to the registry
func (r *RouteRegistry) Register(domain, path, method, description string, authRequired bool) {
	// Check if route already exists
	for i, route := range r.routes {
		if route.Path == path && route.Domain == domain {
			// Add method if not already present
			found := false
			for _, m := range route.Methods {
				if m == method {
					found = true
					break
				}
			}
			if !found {
				r.routes[i].Methods = append(r.routes[i].Methods, method)
				sort.Strings(r.routes[i].Methods)
			}
			return
		}
	}

	// Add new route
	r.routes = append(r.routes, RouteMetadata{
		Path:        path,
		Methods:     []string{method},
		Description: description,
		AuthRequired: authRequired,
		Domain:      domain,
	})
}

// GetRoutes returns all registered routes, grouped by domain
func (r *RouteRegistry) GetRoutes() map[string][]RouteMetadata {
	grouped := make(map[string][]RouteMetadata)

	for _, route := range r.routes {
		grouped[route.Domain] = append(grouped[route.Domain], route)
	}

	// Sort routes within each domain
	for domain := range grouped {
		sort.Slice(grouped[domain], func(i, j int) bool {
			return grouped[domain][i].Path < grouped[domain][j].Path
		})
	}

	return grouped
}

// GetAllRoutes returns all routes in a flat list
func (r *RouteRegistry) GetAllRoutes() []RouteMetadata {
	return r.routes
}

// GenerateHTML generates an HTML page listing all routes
func (r *RouteRegistry) GenerateHTML() string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <title>Wellknown PocketBase API</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; max-width: 1200px; margin: 40px auto; padding: 0 20px; }
        h1 { color: #333; }
        h2 { color: #666; margin-top: 30px; border-bottom: 2px solid #e0e0e0; padding-bottom: 10px; }
        .route { margin: 15px 0; padding: 15px; background: #f8f9fa; border-left: 4px solid #007bff; border-radius: 4px; }
        .methods { display: inline-block; }
        .method { display: inline-block; padding: 4px 8px; margin-right: 5px; border-radius: 3px; font-weight: bold; font-size: 12px; }
        .method.GET { background: #28a745; color: white; }
        .method.POST { background: #ffc107; color: black; }
        .method.PUT { background: #17a2b8; color: white; }
        .method.DELETE { background: #dc3545; color: white; }
        .path { font-family: monospace; font-size: 16px; color: #0366d6; margin: 5px 0; }
        .description { color: #666; margin: 5px 0; }
        .auth-badge { display: inline-block; padding: 4px 8px; background: #e74c3c; color: white; border-radius: 3px; font-size: 11px; font-weight: bold; margin-left: 10px; }
        .tip { background: #fff3cd; padding: 15px; border-radius: 4px; margin-top: 30px; border-left: 4px solid #ffc107; }
    </style>
</head>
<body>
    <h1>🔐 Wellknown PocketBase API</h1>
`)

	// Get routes grouped by domain
	grouped := r.GetRoutes()

	// Sort domains
	domains := make([]string, 0, len(grouped))
	for domain := range grouped {
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	// Generate sections for each domain
	for _, domain := range domains {
		routes := grouped[domain]
		html.WriteString(fmt.Sprintf("    <h2>%s</h2>\n", domain))

		for _, route := range routes {
			html.WriteString("    <div class=\"route\">\n")
			html.WriteString("        <div class=\"methods\">")
			for _, method := range route.Methods {
				html.WriteString(fmt.Sprintf("<span class=\"method %s\">%s</span>", method, method))
			}
			if route.AuthRequired {
				html.WriteString("<span class=\"auth-badge\">🔒 AUTH REQUIRED</span>")
			}
			html.WriteString("</div>\n")
			html.WriteString(fmt.Sprintf("        <div class=\"path\">%s</div>\n", route.Path))
			if route.Description != "" {
				html.WriteString(fmt.Sprintf("        <div class=\"description\">%s</div>\n", route.Description))
			}
			html.WriteString("    </div>\n")
		}
	}

	html.WriteString(`    <div class="tip">
        <strong>💡 Tip:</strong> Visit <a href="/api/">/api/</a> for JSON version
    </div>
</body>
</html>`)

	return html.String()
}

// GenerateJSON returns route data as a structured JSON-compatible map
func (r *RouteRegistry) GenerateJSON() map[string]interface{} {
	return map[string]interface{}{
		"message":   "Wellknown PocketBase API",
		"version":   "1.0.0",
		"endpoints": r.GetRoutes(),
	}
}

// Helper type for registering routes with middleware
type RouteOption func(*RouteMetadata)

// WithAuth marks a route as requiring authentication
func WithAuth() RouteOption {
	return func(m *RouteMetadata) {
		m.AuthRequired = true
	}
}

// WithDescription sets the route description
func WithDescription(desc string) RouteOption {
	return func(m *RouteMetadata) {
		m.Description = desc
	}
}

// RegisterRoute is a helper to register a route with options
func (r *RouteRegistry) RegisterRoute(domain, path, method string, opts ...RouteOption) {
	meta := &RouteMetadata{
		Path:    path,
		Methods: []string{method},
		Domain:  domain,
	}

	for _, opt := range opts {
		opt(meta)
	}

	r.Register(domain, path, method, meta.Description, meta.AuthRequired)
}

// RouteHandler wraps a PocketBase router with route registration
type RouteHandler struct {
	registry *RouteRegistry
	domain   string
	event    *core.ServeEvent
}

// NewRouteHandler creates a new route handler for a domain
func NewRouteHandler(registry *RouteRegistry, domain string, event *core.ServeEvent) *RouteHandler {
	return &RouteHandler{
		registry: registry,
		domain:   domain,
		event:    event,
	}
}

// GET registers a GET route
func (h *RouteHandler) GET(path string, handler func(*core.RequestEvent) error, opts ...RouteOption) {
	meta := &RouteMetadata{}
	for _, opt := range opts {
		opt(meta)
	}

	h.registry.RegisterRoute(h.domain, path, "GET", opts...)
	route := h.event.Router.GET(path, handler)

	// Apply auth middleware if required
	if meta.AuthRequired {
		route.BindFunc(RequireAuth())
	}
}

// POST registers a POST route
func (h *RouteHandler) POST(path string, handler func(*core.RequestEvent) error, opts ...RouteOption) {
	meta := &RouteMetadata{}
	for _, opt := range opts {
		opt(meta)
	}

	h.registry.RegisterRoute(h.domain, path, "POST", opts...)
	route := h.event.Router.POST(path, handler)

	// Apply auth middleware if required
	if meta.AuthRequired {
		route.BindFunc(RequireAuth())
	}
}

// PUT registers a PUT route
func (h *RouteHandler) PUT(path string, handler func(*core.RequestEvent) error, opts ...RouteOption) {
	meta := &RouteMetadata{}
	for _, opt := range opts {
		opt(meta)
	}

	h.registry.RegisterRoute(h.domain, path, "PUT", opts...)
	route := h.event.Router.PUT(path, handler)

	// Apply auth middleware if required
	if meta.AuthRequired {
		route.BindFunc(RequireAuth())
	}
}

// DELETE registers a DELETE route
func (h *RouteHandler) DELETE(path string, handler func(*core.RequestEvent) error, opts ...RouteOption) {
	meta := &RouteMetadata{}
	for _, opt := range opts {
		opt(meta)
	}

	h.registry.RegisterRoute(h.domain, path, "DELETE", opts...)
	route := h.event.Router.DELETE(path, handler)

	// Apply auth middleware if required
	if meta.AuthRequired {
		route.BindFunc(RequireAuth())
	}
}
