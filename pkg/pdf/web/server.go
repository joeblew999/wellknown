package web

import (
	"fmt"
	"net/http"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
	"github.com/joeblew999/wellknown/pkg/pdf/web/api"
	"github.com/joeblew999/wellknown/pkg/pdf/web/gui"
)

// Server represents the PDF form web server
type Server struct {
	port       int
	config     *pdfform.Config
	apiHandler *api.Handler
	guiHandler *gui.Handler
	https      bool
}

// NewServer creates a new web server instance
func NewServer(port int, config *pdfform.Config, https bool) *Server {
	return &Server{
		port:       port,
		config:     config,
		apiHandler: api.NewHandler(config),
		guiHandler: gui.NewHandler(config),
		https:      https,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	// Initialize GUI templates
	if err := gui.InitTemplates(); err != nil {
		return fmt.Errorf("failed to initialize GUI templates: %w", err)
	}

	mux := http.NewServeMux()

	// Register API routes
	s.apiHandler.RegisterRoutes(mux)

	// Register GUI routes
	s.guiHandler.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", s.port)

	if s.https {
		certPath := s.config.CertFilePath()
		keyPath := s.config.KeyFilePath()
		return http.ListenAndServeTLS(addr, certPath, keyPath, mux)
	}

	return http.ListenAndServe(addr, mux)
}

// StartServer is a convenience function to start the server
// Deprecated: Use Start() or StartWithConfig() for easier mounting
func StartServer(port int, config *pdfform.Config, https bool) error {
	server := NewServer(port, config, https)
	return server.Start()
}

// Start starts the web server with auto-discovery and HTTPS by default
// This is the easiest way to mount the web server from other projects.
// Example:
//   import "github.com/joeblew999/wellknown/pkg/pdf/web"
//   web.Start(8080)
func Start(port int) error {
	return StartWithOptions(port, "", true)
}

// StartHTTP starts the web server with auto-discovery using HTTP
// Example:
//   import "github.com/joeblew999/wellknown/pkg/pdf/web"
//   web.StartHTTP(8080)
func StartHTTP(port int) error {
	return StartWithOptions(port, "", false)
}

// StartWithConfig starts the web server with a custom data directory and HTTPS
// Example:
//   import "github.com/joeblew999/wellknown/pkg/pdf/web"
//   web.StartWithConfig(8080, "/custom/data/path")
func StartWithConfig(port int, dataDir string) error {
	return StartWithOptions(port, dataDir, true)
}

// StartWithOptions provides full control over server configuration
// If dataDir is empty, auto-discovery is used
// If https is true, certificates are auto-generated if needed
func StartWithOptions(port int, dataDir string, https bool) error {
	// Auto-discover data directory if not provided
	if dataDir == "" {
		dataDir = pdfform.FindDataDir()
	}

	// Create config
	config := pdfform.NewConfig(dataDir)

	// Ensure directories exist
	if err := config.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Setup HTTPS if enabled
	if https {
		cm := pdfform.NewCertManager(config)
		if err := cm.EnsureCerts(); err != nil {
			fmt.Printf("⚠️  Failed to setup HTTPS: %v\n", err)
			fmt.Println("   Falling back to HTTP...")
			https = false
		}
	}

	// Show server info
	pdfform.PrintServerInfo(fmt.Sprintf("%d", port), https)

	// Start server
	return StartServer(port, config, https)
}
