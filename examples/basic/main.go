package main

import (
	_ "embed"
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

//go:embed index.html
var indexHTML string

var (
	port = flag.String("port", "8080", "Port to run the web server on")
)

type PageData struct {
	GeneratedURL string
	Error        string
	Event        *types.CalendarEvent
}

func main() {
	flag.Parse()

	tmpl := template.Must(template.New("index").Parse(indexHTML))

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

			// Generate URL
			url, err := google.Calendar(event)
			if err != nil {
				tmpl.Execute(w, PageData{
					Error: "Failed to generate URL: " + err.Error(),
					Event: &event,
				})
				return
			}

			// Show result
			tmpl.Execute(w, PageData{
				GeneratedURL: url,
				Event:        &event,
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
