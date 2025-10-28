package server

import (
	"log"
	"net/http"

	"github.com/joeblew999/wellknown/pkg/schema"
)

// CalendarURLGenerator is a function that generates a URL/data URI from validated form data
type CalendarURLGenerator func(data map[string]interface{}) (string, error)

// CalendarConfig configures the generic calendar handler for a specific platform
type CalendarConfig struct {
	Platform     string               // "google" or "apple"
	AppType      string               // "calendar"
	GenerateURL  CalendarURLGenerator // Function to generate URL/data URI from validated data
	SuccessLabel string               // "URL" or "data URI"
}

// makeGenericCalendarHandler creates a handler for calendar event creation bound to server context
// This eliminates code duplication between Google and Apple Calendar handlers
// The handler receives dependencies via HandlerContext instead of using globals
func (s *Server) makeGenericCalendarHandler(hc *HandlerContext, cfg CalendarConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Load schemas once - returns compiled schema + UI schema
			uiSchemaJSON, compiledSchema, _, err := loadAndCompileSchemas(cfg.Platform, cfg.AppType)
			if err != nil {
				http.Error(w, "Failed to load schemas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			hc.renderUISchemaBasedForm(w, r, cfg.Platform, cfg.AppType, compiledSchema, uiSchemaJSON)
			return
		}

		if r.Method == "POST" {
			s.handleCalendarPost(hc, w, r, cfg)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCalendarPost handles POST requests with validation for calendar events
func (s *Server) handleCalendarPost(hc *HandlerContext, w http.ResponseWriter, r *http.Request, cfg CalendarConfig) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Load schemas once - returns compiled schema + UI schema + validator
	uiSchemaJSON, compiledSchema, validator, err := loadAndCompileSchemas(cfg.Platform, cfg.AppType)
	if err != nil {
		log.Printf("Schema load error: %v", err)
		http.Error(w, "Failed to load schemas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert form data to map
	// Properly parses arrays (field[0]), nested objects (field.sub), and array of objects (field[0].sub)
	formData := schema.FormDataToMap(r.Form)

	// Validate against schema using V6 validator
	validationErrors := validator.Validate(formData, compiledSchema)

	// If there are validation errors, re-render form with errors
	if len(validationErrors) > 0 {
		log.Printf("Validation errors: %v", validationErrors)
		hc.renderUISchemaBasedFormWithErrors(w, r, cfg.Platform, cfg.AppType, compiledSchema, uiSchemaJSON, formData, validationErrors)
		return
	}

	// Validation passed - generate URL/data URI directly from validated map data
	// NO MORE BuildEvent callback! No more Event structs!
	url, err := cfg.GenerateURL(formData)
	if err != nil {
		http.Error(w, "Failed to generate "+cfg.SuccessLabel+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS! Generated %s %s (length: %d bytes)", cfg.Platform, cfg.SuccessLabel, len(url))
	hc.renderSuccess(w, r, cfg.Platform, cfg.AppType, url)
}

// makeShowcaseHandler creates a showcase handler bound to server context
func (s *Server) makeShowcaseHandler(hc *HandlerContext, platform, appType string, examples interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hc.renderShowcase(w, r, platform, appType, examples)
	}
}
