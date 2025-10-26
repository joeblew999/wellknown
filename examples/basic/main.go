package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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
	CurrentPage  string
	GeneratedURL string
	Error        string
	Event        *types.CalendarEvent
	TestCases    []google.CalendarTestCase
}

func main() {
	flag.Parse()

	// Parse all templates with composition
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Routes
	http.HandleFunc("/", handleCustom)
	http.HandleFunc("/showcase", handleShowcase)

	addr := ":" + *port
	fmt.Fprintf(os.Stderr, "🚀 wellknown demo server starting...\n")
	fmt.Fprintf(os.Stderr, "📱 Open http://localhost%s in your browser\n", addr)
	fmt.Fprintf(os.Stderr, "   / - Custom (enter your own event)\n")
	fmt.Fprintf(os.Stderr, "   /showcase - Showcase (testdata examples)\n")
	fmt.Fprintf(os.Stderr, "💡 Press Ctrl+C to stop\n\n")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleCustom(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	if r.Method == "GET" {
		// Show custom form
		templates.ExecuteTemplate(w, "base", PageData{
			CurrentPage: "custom",
			TestCases:   google.CalendarEvents,
		})
		return
	}

	if r.Method == "POST" {
		// Handle form submission
		r.ParseForm()

		// Parse times from form
		startTime, err := time.Parse("2006-01-02T15:04", r.FormValue("start_time"))
		if err != nil {
			log.Printf("ERROR parsing start time: %v", err)
			templates.ExecuteTemplate(w, "base", PageData{
				CurrentPage: "custom",
				Error:       "Invalid start time format: " + err.Error(),
				TestCases:   google.CalendarEvents,
			})
			return
		}

		endTime, err := time.Parse("2006-01-02T15:04", r.FormValue("end_time"))
		if err != nil {
			log.Printf("ERROR parsing end time: %v", err)
			templates.ExecuteTemplate(w, "base", PageData{
				CurrentPage: "custom",
				Error:       "Invalid end time format: " + err.Error(),
				TestCases:   google.CalendarEvents,
			})
			return
		}

		// Create event
		event := types.CalendarEvent{
			Title:       r.FormValue("title"),
			StartTime:   startTime,
			EndTime:     endTime,
			Location:    r.FormValue("location"),
			Description: r.FormValue("description"),
		}

		// Generate URL - pkg validates!
		url, err := google.Calendar(event)
		if err != nil {
			log.Printf("ERROR: Validation failed: %v", err)
			templates.ExecuteTemplate(w, "base", PageData{
				CurrentPage: "custom",
				Error:       err.Error(),
				Event:       &event,
				TestCases:   google.CalendarEvents,
			})
			return
		}

		log.Printf("SUCCESS! Generated URL: %s", url)

		// Show result
		templates.ExecuteTemplate(w, "base", PageData{
			CurrentPage:  "custom",
			GeneratedURL: url,
			Event:        &event,
			TestCases:    google.CalendarEvents,
		})
		return
	}
}

func handleShowcase(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Show showcase page with testdata
	templates.ExecuteTemplate(w, "base", PageData{
		CurrentPage: "showcase",
		TestCases:   google.CalendarEvents,
	})
}
