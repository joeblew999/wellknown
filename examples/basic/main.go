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
	"github.com/joeblew999/wellknown/pkg/web"
)

//go:embed templates/*
var templatesFS embed.FS

var (
	port = flag.String("port", "8080", "Port to run the web server on")
)

type PageData struct {
	NativeURL string
	WebURL    string
	Error     string
	Event     *types.CalendarEvent
}

func main() {
	flag.Parse()

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Show empty form
			tmpl.Execute(w, PageData{})
			return
		}

		if r.Method == "POST" {
			// Parse form
			r.ParseForm()

			// Parse times
			startTime, err := time.Parse("2006-01-02T15:04", r.FormValue("start_time"))
			if err != nil {
				tmpl.Execute(w, PageData{Error: "Invalid start time: " + err.Error()})
				return
			}

			endTime, err := time.Parse("2006-01-02T15:04", r.FormValue("end_time"))
			if err != nil {
				tmpl.Execute(w, PageData{Error: "Invalid end time: " + err.Error()})
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

			// Generate native deep link URL
			nativeURL, err := google.Calendar(event)
			if err != nil {
				tmpl.Execute(w, PageData{
					Error: "Failed to generate native URL: " + err.Error(),
					Event: &event,
				})
				return
			}

			// Generate web fallback URL (testable in browser!)
			webURL, err := web.GoogleCalendar(event)
			if err != nil {
				tmpl.Execute(w, PageData{
					Error: "Failed to generate web URL: " + err.Error(),
					Event: &event,
				})
				return
			}

			// Show both URLs
			tmpl.Execute(w, PageData{
				NativeURL: nativeURL,
				WebURL:    webURL,
				Event:     &event,
			})
			return
		}
	})

	addr := ":" + *port
	fmt.Fprintf(os.Stderr, "🚀 wellknown demo server starting...\n")
	fmt.Fprintf(os.Stderr, "📱 Open http://localhost%s in your browser\n", addr)
	fmt.Fprintf(os.Stderr, "💡 Press Ctrl+C to stop\n\n")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
