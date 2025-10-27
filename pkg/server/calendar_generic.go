package server

import (
	"log"
	"net/http"
	"time"

	"github.com/joeblew999/wellknown/pkg/schema"
)

// CalendarEventBuilder is a function that creates a platform-specific event from form data
type CalendarEventBuilder func(r *http.Request) (interface{}, error)

// CalendarURLGenerator is a function that generates a URL/data URI from a platform-specific event
type CalendarURLGenerator func(event interface{}) (string, error)

// CalendarConfig configures the generic calendar handler for a specific platform
type CalendarConfig struct {
	Platform     string               // "google" or "apple"
	AppType      string               // "calendar"
	BuildEvent   CalendarEventBuilder // Function to build platform-specific event from form
	GenerateURL  CalendarURLGenerator // Function to generate URL/data URI from event
	SuccessLabel string               // "URL" or "data URI"
}

// GenericCalendarHandler creates a handler for calendar event creation
// This eliminates code duplication between Google and Apple Calendar handlers
func GenericCalendarHandler(cfg CalendarConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Load schemas from external JSON files
			schemaJSON, err := loadSchemaFromFile(cfg.Platform, cfg.AppType, "schema")
			if err != nil {
				http.Error(w, "Failed to load schema: "+err.Error(), http.StatusInternalServerError)
				return
			}
			uiSchemaJSON, err := loadSchemaFromFile(cfg.Platform, cfg.AppType, "uischema")
			if err != nil {
				http.Error(w, "Failed to load UI schema: "+err.Error(), http.StatusInternalServerError)
				return
			}
			renderUISchemaBasedForm(w, r, cfg.Platform, cfg.AppType, schemaJSON, uiSchemaJSON)
			return
		}

		if r.Method == "POST" {
			handleCalendarPost(w, r, cfg)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCalendarPost handles POST requests with validation for calendar events
func handleCalendarPost(w http.ResponseWriter, r *http.Request, cfg CalendarConfig) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Load schemas from external JSON files
	schemaJSON, err := loadSchemaFromFile(cfg.Platform, cfg.AppType, "schema")
	if err != nil {
		http.Error(w, "Failed to load schema: "+err.Error(), http.StatusInternalServerError)
		return
	}
	uiSchemaJSON, err := loadSchemaFromFile(cfg.Platform, cfg.AppType, "uischema")
	if err != nil {
		http.Error(w, "Failed to load UI schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse JSON Schema for validation
	jsonSchema, err := schema.ParseSchema(schemaJSON)
	if err != nil {
		log.Printf("Schema parse error: %v", err)
		http.Error(w, "Failed to parse schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert form data to map
	formData := schema.FormDataToMap(r.Form)

	// Validate against schema
	validationErrors := schema.ValidateAgainstSchema(formData, jsonSchema)

	// If there are validation errors, re-render form with errors
	if len(validationErrors) > 0 {
		log.Printf("Validation errors: %v", validationErrors)
		renderUISchemaBasedFormWithErrors(w, r, cfg.Platform, cfg.AppType, schemaJSON, uiSchemaJSON, formData, validationErrors)
		return
	}

	// Validation passed - build platform-specific event
	event, err := cfg.BuildEvent(r)
	if err != nil {
		http.Error(w, "Failed to build event: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Generate URL or data URI
	url, err := cfg.GenerateURL(event)
	if err != nil {
		http.Error(w, "Failed to generate "+cfg.SuccessLabel+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS! Generated %s %s (length: %d bytes)", cfg.Platform, cfg.SuccessLabel, len(url))
	renderSuccess(w, r, cfg.Platform, cfg.AppType, url)
}

// parseFormTime is a helper to parse datetime-local form values
func parseFormTime(value string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04", value)
}
