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

// makeGenericCalendarHandler creates a handler for calendar event creation
func (s *Server) makeGenericCalendarHandler(cfg CalendarConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			s.handleCalendarGET(w, r, cfg)
			return
		}

		if r.Method == "POST" {
			s.handleCalendarPOST(w, r, cfg)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCalendarGET renders the calendar form
func (s *Server) handleCalendarGET(w http.ResponseWriter, r *http.Request, cfg CalendarConfig) {
	// Load schemas
	uiSchemaJSON, compiledSchema, _, err := schema.LoadSchemasForRendering(cfg.Platform, cfg.AppType)
	if err != nil {
		http.Error(w, "Failed to load schemas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse UI Schema and generate form HTML
	uiSchema, err := schema.ParseUISchema(uiSchemaJSON)
	if err != nil {
		http.Error(w, "Failed to parse UI schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	formHTML := uiSchema.GenerateFormHTMLWithData(compiledSchema, nil, nil)

	// Render
	s.render(w, r, PageData{
		Platform:       cfg.Platform,
		AppType:        cfg.AppType,
		CurrentPage:    "custom",
		TemplateName:   "schema_form",
		SchemaFormHTML: formHTML,
	})
}

// handleCalendarPOST handles form submission with validation
func (s *Server) handleCalendarPOST(w http.ResponseWriter, r *http.Request, cfg CalendarConfig) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Load schemas
	uiSchemaJSON, compiledSchema, validator, err := schema.LoadSchemasForRendering(cfg.Platform, cfg.AppType)
	if err != nil {
		http.Error(w, "Failed to load schemas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert and validate
	formData := schema.FormDataToMap(r.Form)
	validationErrors := validator.Validate(formData, compiledSchema)

	// If validation failed, re-render form with errors
	if len(validationErrors) > 0 {
		log.Printf("Validation errors: %v", validationErrors)

		uiSchema, err := schema.ParseUISchema(uiSchemaJSON)
		if err != nil {
			http.Error(w, "Failed to parse UI schema: "+err.Error(), http.StatusInternalServerError)
			return
		}

		formHTML := uiSchema.GenerateFormHTMLWithData(compiledSchema, formData, validationErrors)

		s.render(w, r, PageData{
			Platform:         cfg.Platform,
			AppType:          cfg.AppType,
			CurrentPage:      "custom",
			TemplateName:     "schema_form",
			SchemaFormHTML:   formHTML,
			FormData:         formData,
			ValidationErrors: validationErrors,
		})
		return
	}

	// Generate URL
	url, err := cfg.GenerateURL(formData)
	if err != nil {
		http.Error(w, "Failed to generate "+cfg.SuccessLabel+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS! Generated %s %s (length: %d bytes)", cfg.Platform, cfg.SuccessLabel, len(url))

	// Render success
	s.render(w, r, PageData{
		Platform:     cfg.Platform,
		AppType:      cfg.AppType,
		CurrentPage:  "custom",
		TemplateName: "success",
		GeneratedURL: url,
	})
}

// makeExamplesHandler creates a showcase handler
func (s *Server) makeExamplesHandler(platform, appType string, examples interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		s.render(w, r, PageData{
			Platform:     platform,
			AppType:      appType,
			CurrentPage:  "showcase",
			TemplateName: "showcase",
			TestCases:    examples,
		})
	}
}
