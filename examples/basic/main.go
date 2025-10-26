package main

import (
	"embed"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/joeblew999/wellknown/examples/basic/handlers"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	port = flag.String("port", "8080", "Port to run the web server on")
)

func main() {
	flag.Parse()

	// Parse all templates with custom functions
	funcMap := template.FuncMap{
		"title": strings.Title,
	}
	var err error
	handlers.Templates, err = template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Routes
	http.HandleFunc("/", handlers.GoogleCalendar)
	http.HandleFunc("/google/calendar", handlers.GoogleCalendar)
	http.HandleFunc("/google/calendar/showcase", handlers.GoogleCalendarShowcase)
	http.HandleFunc("/google/maps", handlers.Stub("google", "maps"))
	http.HandleFunc("/google/maps/showcase", handlers.Stub("google", "maps"))

	addr := ":" + *port
	localIP := getLocalIP()

	// Set global URLs for templates
	handlers.LocalURL = "http://localhost" + addr
	handlers.MobileURL = "http://" + localIP + addr

	log.Println("🚀 wellknown demo server starting...")
	log.Printf("💻 Local:  %s", handlers.LocalURL)
	log.Printf("📱 Mobile: %s", handlers.MobileURL)
	log.Println("")
	log.Println("💡 Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

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
