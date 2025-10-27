package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/joeblew999/wellknown/pkg/google"
	"github.com/joeblew999/wellknown/pkg/types"
)

func GoogleCalendar(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	if r.Method == "GET" {
		err := Templates.ExecuteTemplate(w, "base", PageData{
			Platform:     "google",
			AppType:      "calendar",
			CurrentPage:  "custom",
			TemplateName: "google_calendar_custom",
			TestCases:    google.CalendarExamples,
			LocalURL:     LocalURL,
			MobileURL:    MobileURL,
		})
		if err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		r.ParseForm()

		startTime, err := time.Parse("2006-01-02T15:04", r.FormValue("start_time"))
		if err != nil {
			Templates.ExecuteTemplate(w, "base", PageData{
				Platform:     "google",
				AppType:      "calendar",
				CurrentPage:  "custom",
				TemplateName: "google_calendar_custom",
				Error:        "Invalid start time format: " + err.Error(),
				TestCases:    google.CalendarExamples,
				LocalURL:     LocalURL,
				MobileURL:    MobileURL,
			})
			return
		}

		endTime, err := time.Parse("2006-01-02T15:04", r.FormValue("end_time"))
		if err != nil {
			Templates.ExecuteTemplate(w, "base", PageData{
				Platform:     "google",
				AppType:      "calendar",
				CurrentPage:  "custom",
				TemplateName: "google_calendar_custom",
				Error:        "Invalid end time format: " + err.Error(),
				TestCases:    google.CalendarExamples,
				LocalURL:     LocalURL,
				MobileURL:    MobileURL,
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
			Templates.ExecuteTemplate(w, "base", PageData{
				Platform:     "google",
				AppType:      "calendar",
				CurrentPage:  "custom",
				TemplateName: "google_calendar_custom",
				Error:        err.Error(),
				Event:        &event,
				TestCases:    google.CalendarExamples,
				LocalURL:     LocalURL,
				MobileURL:    MobileURL,
			})
			return
		}

		log.Printf("SUCCESS! Generated URL: %s", url)

		// Use the same web URL - mobile browsers will offer to open in the Google Calendar app
		// The web URL works on both desktop and mobile, and mobile can deep link to the app
		appURL := url

		Templates.ExecuteTemplate(w, "base", PageData{
			Platform:     "google",
			AppType:      "calendar",
			CurrentPage:  "custom",
			TemplateName: "google_calendar_custom",
			GeneratedURL: url,
			AppURL:       appURL,
			Event:        &event,
			TestCases:    google.CalendarExamples,
			LocalURL:     LocalURL,
			MobileURL:    MobileURL,
		})
		return
	}
}

func GoogleCalendarShowcase(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Note: AppURL is set per test case in the template (each has its own URL)
	Templates.ExecuteTemplate(w, "base", PageData{
		Platform:     "google",
		AppType:      "calendar",
		CurrentPage:  "showcase",
		TemplateName: "google_calendar_showcase",
		TestCases:    google.CalendarEvents,
		LocalURL:     LocalURL,
		MobileURL:    MobileURL,
	})
}
