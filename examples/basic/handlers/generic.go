package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joeblew999/wellknown/pkg/types"
)

// CalendarGenerator is a function that generates a calendar URL from an event
type CalendarGenerator func(types.CalendarEvent) (string, error)

// ServiceConfig holds configuration for a service handler
type ServiceConfig struct {
	Platform  string
	AppType   string
	Examples  interface{} // TestCases for showcase
	Generator CalendarGenerator
}

// ServiceRegistration holds a service config and its handlers
type ServiceRegistration struct {
	Config          ServiceConfig
	CustomHandler   http.HandlerFunc
	ShowcaseHandler http.HandlerFunc
}

// serviceRegistry holds all registered services
var serviceRegistry = make(map[string]*ServiceRegistration)

// RegisterService registers a service and returns its handlers
func RegisterService(config ServiceConfig) *ServiceRegistration {
	key := fmt.Sprintf("%s/%s", config.Platform, config.AppType)

	registration := &ServiceRegistration{
		Config:          config,
		CustomHandler:   CalendarHandler(config),
		ShowcaseHandler: ShowcaseHandler(config),
	}

	serviceRegistry[key] = registration
	return registration
}

// GetAllServices returns all registered services
func GetAllServices() map[string]*ServiceRegistration {
	return serviceRegistry
}

// RegisterRoutes automatically registers all service routes with http.DefaultServeMux
func RegisterRoutes() {
	for key, service := range serviceRegistry {
		// Register custom handler
		customPath := fmt.Sprintf("/%s", key)
		http.HandleFunc(customPath, service.CustomHandler)
		log.Printf("Registered route: %s", customPath)

		// Register showcase handler
		showcasePath := fmt.Sprintf("/%s/showcase", key)
		http.HandleFunc(showcasePath, service.ShowcaseHandler)
		log.Printf("Registered route: %s", showcasePath)
	}
}

// CalendarHandler creates a generic handler for calendar services
func CalendarHandler(config ServiceConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)

		if r.Method == "GET" {
			err := Templates.ExecuteTemplate(w, "base", PageData{
				Platform:     config.Platform,
				AppType:      config.AppType,
				CurrentPage:  "custom",
				TemplateName: "custom",
				TestCases:    config.Examples,
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

			// Parse form data into CalendarEvent
			event, err := parseCalendarEventForm(r)
			if err != nil {
				Templates.ExecuteTemplate(w, "base", PageData{
					Platform:     config.Platform,
					AppType:      config.AppType,
					CurrentPage:  "custom",
					TemplateName: "custom",
					Error:        err.Error(),
					TestCases:    config.Examples,
					LocalURL:     LocalURL,
					MobileURL:    MobileURL,
				})
				return
			}

			// Generate URL using the service-specific generator
			url, err := config.Generator(event)
			if err != nil {
				Templates.ExecuteTemplate(w, "base", PageData{
					Platform:     config.Platform,
					AppType:      config.AppType,
					CurrentPage:  "custom",
					TemplateName: "custom",
					Error:        err.Error(),
					Event:        &event,
					TestCases:    config.Examples,
					LocalURL:     LocalURL,
					MobileURL:    MobileURL,
				})
				return
			}

			log.Printf("SUCCESS! Generated URL: %s", url)

			Templates.ExecuteTemplate(w, "base", PageData{
				Platform:     config.Platform,
				AppType:      config.AppType,
				CurrentPage:  "custom",
				TemplateName: "custom",
				GeneratedURL: url,
				AppURL:       url, // For most services, web URL works universally
				Event:        &event,
				TestCases:    config.Examples,
				LocalURL:     LocalURL,
				MobileURL:    MobileURL,
			})
			return
		}
	}
}

// ShowcaseHandler creates a generic showcase handler for any service
func ShowcaseHandler(config ServiceConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)

		Templates.ExecuteTemplate(w, "base", PageData{
			Platform:     config.Platform,
			AppType:      config.AppType,
			CurrentPage:  "showcase",
			TemplateName: "showcase",
			TestCases:    config.Examples,
			LocalURL:     LocalURL,
			MobileURL:    MobileURL,
		})
	}
}

// parseCalendarEventForm parses form data into a CalendarEvent
func parseCalendarEventForm(r *http.Request) (types.CalendarEvent, error) {
	startTime, err := time.Parse("2006-01-02T15:04", r.FormValue("start_time"))
	if err != nil {
		return types.CalendarEvent{}, err
	}

	endTime, err := time.Parse("2006-01-02T15:04", r.FormValue("end_time"))
	if err != nil {
		return types.CalendarEvent{}, err
	}

	return types.CalendarEvent{
		Title:       r.FormValue("title"),
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    r.FormValue("location"),
		Description: r.FormValue("description"),
	}, nil
}
