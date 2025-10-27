package server

import (
	"log"
	"net/http"
)

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
	schema, err := ParseSchema(schemaJSON)
	if err != nil {
		log.Printf("Schema parse error: %v", err)
		http.Error(w, "Failed to parse schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate form HTML from schema
	formHTML := schema.GenerateFormHTML()

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
func renderUISchemaBasedFormWithErrors(w http.ResponseWriter, r *http.Request, platform, appType, schemaJSON, uiSchemaJSON string, formData map[string]interface{}, validationErrors ValidationErrors) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse JSON Schema
	schema, err := ParseSchema(schemaJSON)
	if err != nil {
		log.Printf("JSON Schema parse error: %v", err)
		http.Error(w, "Failed to parse JSON schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse UI Schema
	uiSchema, err := ParseUISchema(uiSchemaJSON)
	if err != nil {
		log.Printf("UI Schema parse error: %v", err)
		http.Error(w, "Failed to parse UI schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate form HTML from UI Schema + JSON Schema
	// TODO: Pass formData and validationErrors to pre-fill form and show errors
	formHTML := uiSchema.GenerateFormHTML(schema)

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
