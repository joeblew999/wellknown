package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
)

//go:embed templates/*
var templatesFS embed.FS

// Server represents the wellknown demo/test server with all dependencies
type Server struct {
	Port      string
	LocalURL  string
	MobileURL string
	URLPrefix string // URL prefix when embedded (e.g., "/demo" in PocketBase)

	// Dependencies (no more globals!)
	templates *template.Template
	mux       *http.ServeMux
	registry  *ServiceRegistry

	// State (no more package-level globals!)
	gcpSetupStatus GCPSetupStatus
}

// New creates a new Server instance with all dependencies initialized
func New(port string) (*Server, error) {
	if port == "" {
		port = "8080"
	}

	// Initialize templates
	funcMap := template.FuncMap{
		"title": strings.Title,
		"toJSON": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	addr := ":" + port
	localIP := getLocalIP()

	server := &Server{
		Port:      port,
		LocalURL:  "http://localhost" + addr,
		MobileURL: "http://" + localIP + addr,
		templates: tmpl,
		mux:       http.NewServeMux(),
		registry:  NewServiceRegistry(),
	}

	// Register all routes with server's mux and registry
	server.registerAllRoutes()

	return server, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":" + s.Port

	log.Println("🚀 wellknown demo server starting...")
	log.Printf("💻 Local:  %s", s.LocalURL)
	log.Printf("📱 Mobile: %s", s.MobileURL)
	log.Println("")
	log.Println("💡 Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, s.mux); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}

// GetTemplates returns the server's template instance (for handlers)
func (s *Server) GetTemplates() *template.Template {
	return s.templates
}

// GetMux returns the server's HTTP mux (for testing)
func (s *Server) GetMux() *http.ServeMux {
	return s.mux
}

// GetRegistry returns the server's service registry (for handlers)
func (s *Server) GetRegistry() *ServiceRegistry {
	return s.registry
}

// getLocalIP returns the local IP address of the machine
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

// render is the SINGLE method to render any page
// All other render methods call this with different PageData
func (s *Server) render(w http.ResponseWriter, r *http.Request, data PageData) {
	// Auto-populate common fields
	data.LocalURL = s.LocalURL
	data.MobileURL = s.MobileURL
	data.URLPrefix = s.URLPrefix

	// Use prefix-aware navigation if URLPrefix is set
	if s.URLPrefix != "" {
		data.Navigation = s.registry.GetNavigationWithPrefix(r.URL.Path, s.URLPrefix)
	} else {
		data.Navigation = s.registry.GetNavigation(r.URL.Path)
	}

	// Render template
	err := s.templates.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}
