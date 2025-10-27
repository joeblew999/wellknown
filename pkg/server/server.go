package server

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
)

//go:embed templates/*
var templatesFS embed.FS

// Server represents the wellknown demo/test server
type Server struct {
	Port      string
	LocalURL  string
	MobileURL string
}

// New creates a new Server instance with the specified port
func New(port string) (*Server, error) {
	if port == "" {
		port = "8080"
	}

	// Parse all templates with custom functions
	funcMap := template.FuncMap{
		"title": strings.Title,
	}
	var err error
	Templates, err = template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	addr := ":" + port
	localIP := getLocalIP()

	// Set package-level URLs for use in handlers
	LocalURL = "http://localhost" + addr
	MobileURL = "http://" + localIP + addr

	server := &Server{
		Port:      port,
		LocalURL:  LocalURL,
		MobileURL: MobileURL,
	}

	// Register all HTTP routes
	server.registerRoutes()

	return server, nil
}

// registerRoutes sets up all HTTP routes
func (s *Server) registerRoutes() {
	// Google Calendar
	http.HandleFunc("/google/calendar", GoogleCalendar)
	http.HandleFunc("/google/calendar/showcase", GoogleCalendarShowcase)
	http.HandleFunc("/google/calendar-schema", GoogleCalendarSchema) // JSON Schema version (WIP)

	// Apple Calendar
	http.HandleFunc("/apple/calendar", AppleCalendar)
	http.HandleFunc("/apple/calendar/showcase", AppleCalendarShowcase)

	// Stub routes (placeholders for future services)
	http.HandleFunc("/", GoogleCalendar) // Default to Google Calendar
	http.HandleFunc("/google/maps", Stub("google", "maps"))
	http.HandleFunc("/google/maps/showcase", Stub("google", "maps"))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":" + s.Port

	log.Println("🚀 wellknown demo server starting...")
	log.Printf("💻 Local:  %s", s.LocalURL)
	log.Printf("📱 Mobile: %s", s.MobileURL)
	log.Println("")
	log.Println("💡 Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
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
