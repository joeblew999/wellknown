package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joeblew999/wellknown/pkg/schema"
)

// loadSchemaFromFile loads a JSON Schema from an external file
// Handles both running from project root and from cmd/server/
func loadSchemaFromFile(platform, appType, schemaType string) (string, error) {
	// Try relative path from project root first
	path := fmt.Sprintf("pkg/%s/%s/%s.json", platform, appType, schemaType)
	content, err := os.ReadFile(path)
	if err != nil {
		// If that fails, try from cmd/server/ directory (Air case)
		path = fmt.Sprintf("../../pkg/%s/%s/%s.json", platform, appType, schemaType)
		content, err = os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read schema file: %w", err)
		}
	}
	return string(content), nil
}

// renderPage is a helper to render a page with error handling
func renderPage(w http.ResponseWriter, platform, appType, currentPage, templateName string, data interface{}) {
	err := Templates.ExecuteTemplate(w, "base", PageData{
		Platform:     platform,
		AppType:      appType,
		CurrentPage:  currentPage,
		TemplateName: templateName,
		TestCases:    data,
		LocalURL:     LocalURL,
		MobileURL:    MobileURL,
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// renderShowcase is a helper specifically for showcase pages
func renderShowcase(w http.ResponseWriter, r *http.Request, platform, appType string, examples interface{}) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	renderPage(w, platform, appType, "showcase", "showcase", examples)
}

// renderCustomPage is a helper specifically for custom form pages
func renderCustomPage(w http.ResponseWriter, r *http.Request, platform, appType string, examples interface{}) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	renderPage(w, platform, appType, "custom", "custom", examples)
}

// renderSchemaBasedForm renders a form generated from JSON Schema
func renderSchemaBasedForm(w http.ResponseWriter, r *http.Request, platform, appType, schemaJSON string) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse the schema
	jsonSchema, err := schema.ParseSchema(schemaJSON)
	if err != nil {
		log.Printf("Schema parse error: %v", err)
		http.Error(w, "Failed to parse schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate form HTML from schema
	formHTML := jsonSchema.GenerateFormHTML()

	// Render with schema-generated form
	err = Templates.ExecuteTemplate(w, "base", PageData{
		Platform:       platform,
		AppType:        appType,
		CurrentPage:    "schema", // Set to "schema" for menu highlighting
		TemplateName:   "schema_form",
		SchemaFormHTML: formHTML,
		LocalURL:       LocalURL,
		MobileURL:      MobileURL,
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// renderUISchemaBasedForm renders a form generated from JSON Schema + UI Schema
func renderUISchemaBasedForm(w http.ResponseWriter, r *http.Request, platform, appType, schemaJSON, uiSchemaJSON string) {
	renderUISchemaBasedFormWithErrors(w, r, platform, appType, schemaJSON, uiSchemaJSON, nil, nil)
}

// renderUISchemaBasedFormWithErrors renders a form with validation errors and pre-filled data
func renderUISchemaBasedFormWithErrors(w http.ResponseWriter, r *http.Request, platform, appType, schemaJSON, uiSchemaJSON string, formData map[string]interface{}, validationErrors schema.ValidationErrors) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse JSON Schema
	jsonSchema, err := schema.ParseSchema(schemaJSON)
	if err != nil {
		log.Printf("JSON Schema parse error: %v", err)
		http.Error(w, "Failed to parse JSON schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse UI Schema
	uiSchema, err := schema.ParseUISchema(uiSchemaJSON)
	if err != nil {
		log.Printf("UI Schema parse error: %v", err)
		http.Error(w, "Failed to parse UI schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate form HTML from UI Schema + JSON Schema with validation errors
	formHTML := uiSchema.GenerateFormHTMLWithData(jsonSchema, formData, validationErrors)

	// Render with UI schema-generated form
	err = Templates.ExecuteTemplate(w, "base", PageData{
		Platform:         platform,
		AppType:          appType,
		CurrentPage:      "ui-schema", // Set to "ui-schema" for menu highlighting
		TemplateName:     "schema_form",
		SchemaFormHTML:   formHTML,
		FormData:         formData,
		ValidationErrors: validationErrors,
		LocalURL:         LocalURL,
		MobileURL:        MobileURL,
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// renderSuccess renders a success page with the generated URL
func renderSuccess(w http.ResponseWriter, platform, appType, generatedURL string) {
	err := Templates.ExecuteTemplate(w, "base", PageData{
		Platform:     platform,
		AppType:      appType,
		CurrentPage:  "custom",
		TemplateName: "custom",
		GeneratedURL: generatedURL,
		LocalURL:     LocalURL,
		MobileURL:    MobileURL,
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}
