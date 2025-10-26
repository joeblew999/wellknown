package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joeblew999/wellknown/pkg/google"
	"github.com/joeblew999/wellknown/pkg/types"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	port      = flag.String("port", "8080", "Port to run the web server on")
	templates *template.Template
)

type PageData struct {
	Platform     string // "google", "apple"
	AppType      string // "calendar", "maps", "drive", etc
	CurrentPage  string // "custom", "showcase"
	IsStub       bool   // true if this is a stub page
	GeneratedURL string
	Error        string
	Event        *types.CalendarEvent
	TestCases    []google.CalendarTestCase
}

func main() {
	flag.Parse()

	// Parse all templates with custom functions
	var err error
	funcMap := template.FuncMap{
		"title": strings.Title,
	}
	templates, err = template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Google Calendar routes
	http.HandleFunc("/google/calendar", handleGoogleCalendar)
	http.HandleFunc("/google/calendar/showcase", handleGoogleCalendarShowcase)

	// Google Maps routes (stub)
	http.HandleFunc("/google/maps", handleStub("google", "maps"))
	http.HandleFunc("/google/maps/showcase", handleStub("google", "maps"))

	// Redirect root to /google/calendar
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/google/calendar", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	addr := ":" + *port
	fmt.Fprintf(os.Stderr, "🚀 wellknown demo server starting...\n")
	fmt.Fprintf(os.Stderr, "📱 Open http://localhost%s in your browser\n", addr)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Routes:\n")
	fmt.Fprintf(os.Stderr, "  /google/calendar          - Google Calendar Custom\n")
	fmt.Fprintf(os.Stderr, "  /google/calendar/showcase - Google Calendar Showcase\n")
	fmt.Fprintf(os.Stderr, "  /google/maps              - Google Maps (stub)\n")
	fmt.Fprintf(os.Stderr, "  /google/maps/showcase     - Google Maps Showcase (stub)\n")
	fmt.Fprintf(os.Stderr, "\n💡 Press Ctrl+C to stop\n\n")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleGoogleCalendar(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	if r.Method == "GET" {
		templates.ExecuteTemplate(w, "base", PageData{
			Platform:    "google",
			AppType:     "calendar",
			CurrentPage: "custom",
			TestCases:   google.CalendarEvents,
		})
		return
	}

	if r.Method == "POST" {
		r.ParseForm()

		startTime, err := time.Parse("2006-01-02T15:04", r.FormValue("start_time"))
		if err != nil {
			templates.ExecuteTemplate(w, "base", PageData{
				Platform:    "google",
				AppType:     "calendar",
				CurrentPage: "custom",
				Error:       "Invalid start time format: " + err.Error(),
				TestCases:   google.CalendarEvents,
			})
			return
		}

		endTime, err := time.Parse("2006-01-02T15:04", r.FormValue("end_time"))
		if err != nil {
			templates.ExecuteTemplate(w, "base", PageData{
				Platform:    "google",
				AppType:     "calendar",
				CurrentPage: "custom",
				Error:       "Invalid end time format: " + err.Error(),
				TestCases:   google.CalendarEvents,
			})
			return
		}

		event := types.CalendarEvent{
			Title:       r.FormValue("title"),
			StartTime:   startTime,
			EndTime:     endTime,
			Location:    r.FormValue("location"),
			Description: r.FormValue("description"),
		}

		url, err := google.Calendar(event)
		if err != nil {
			templates.ExecuteTemplate(w, "base", PageData{
				Platform:    "google",
				AppType:     "calendar",
				CurrentPage: "custom",
				Error:       err.Error(),
				Event:       &event,
				TestCases:   google.CalendarEvents,
			})
			return
		}

		log.Printf("SUCCESS! Generated URL: %s", url)

		templates.ExecuteTemplate(w, "base", PageData{
			Platform:     "google",
			AppType:      "calendar",
			CurrentPage:  "custom",
			GeneratedURL: url,
			Event:        &event,
			TestCases:    google.CalendarEvents,
		})
		return
	}
}

func handleGoogleCalendarShowcase(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	templates.ExecuteTemplate(w, "base", PageData{
		Platform:    "google",
		AppType:     "calendar",
		CurrentPage: "showcase",
		TestCases:   google.CalendarEvents,
	})
}

// handleStub returns a handler for stub pages
func handleStub(platform, appType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s (stub)", r.Method, r.URL.Path)

		currentPage := "custom"
		if strings.HasSuffix(r.URL.Path, "/showcase") {
			currentPage = "showcase"
		}

		templates.ExecuteTemplate(w, "base", PageData{
			Platform:    platform,
			AppType:     appType,
			CurrentPage: currentPage,
			IsStub:      true,
		})
	}
}
