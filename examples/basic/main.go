package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
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

	fmt.Fprintf(os.Stderr, "🚀 wellknown demo server starting...\n")
	fmt.Fprintf(os.Stderr, "💻 Local:  http://localhost%s\n", addr)
	fmt.Fprintf(os.Stderr, "📱 Mobile: http://%s%s\n", localIP, addr)
	fmt.Fprintf(os.Stderr, "\n💡 Press Ctrl+C to stop\n\n")

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
