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
	"github.com/joeblew999/wellknown/pkg/testdata"
	"github.com/joeblew999/wellknown/pkg/types"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	port = flag.String("port", "8080", "Port to run the web server on")
	tmpl *template.Template
)

type PageData struct {
	CurrentPage  string // "custom" or "showcase"
	GeneratedURL string
	Error        string
	Event        *types.CalendarEvent
	TestCases    []testdata.CalendarTestCase
}

func main() {
	flag.Parse()
	
	tmpl = template.Must(template.ParseFS(templatesFS, "templates/index.html"))

	// Route: / (Custom - user input)
	http.HandleFunc("/", handleCustom)
	
	// Route: /showcase (Showcase - testdata)
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
		tmpl.Execute(w, PageData{
			CurrentPage: "custom",
			TestCases:   testdata.CalendarEvents,
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
			tmpl.Execute(w, PageData{
				CurrentPage: "custom",
				Error:       "Invalid start time format: " + err.Error(),
				TestCases:   testdata.CalendarEvents,
			})
			return
		}

		endTime, err := time.Parse("2006-01-02T15:04", r.FormValue("end_time"))
		if err != nil {
			log.Printf("ERROR parsing end time: %v", err)
			tmpl.Execute(w, PageData{
				CurrentPage: "custom",
				Error:       "Invalid end time format: " + err.Error(),
				TestCases:   testdata.CalendarEvents,
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
			tmpl.Execute(w, PageData{
				CurrentPage: "custom",
				Error:       err.Error(),
				Event:       &event,
				TestCases:   testdata.CalendarEvents,
			})
			return
		}

		log.Printf("SUCCESS! Generated URL: %s", url)

		// Show result
		tmpl.Execute(w, PageData{
			CurrentPage:  "custom",
			GeneratedURL: url,
			Event:        &event,
			TestCases:    testdata.CalendarEvents,
		})
		return
	}
}

func handleShowcase(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	
	// Show showcase page with testdata
	tmpl.Execute(w, PageData{
		CurrentPage: "showcase",
		TestCases:   testdata.CalendarEvents,
	})
}
