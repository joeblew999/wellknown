package server

import (
	"log"
	"net/http"

	"github.com/joeblew999/wellknown/pkg/schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// loadAndCompileSchemas is now a thin wrapper around schema.LoadSchemasForRendering
// This maintains backwards compatibility with existing code while using the centralized loader
func loadAndCompileSchemas(platform, appType string) (string, *jsonschema.Schema, *schema.ValidatorV6, error) {
	return schema.LoadSchemasForRendering(platform, appType)
}

// HandlerContext wraps Server dependencies for handler functions
// This eliminates the need for package-level globals
type HandlerContext struct {
	server *Server
}

// newHandlerContext creates a context for handlers with server dependencies
func (s *Server) newHandlerContext() *HandlerContext {
	return &HandlerContext{server: s}
}

// renderPage is a helper to render a page with error handling
func (hc *HandlerContext) renderPage(w http.ResponseWriter, r *http.Request, platform, appType, currentPage, templateName string, data interface{}) {
	err := hc.server.templates.ExecuteTemplate(w, "base", PageData{
		Platform:     platform,
		AppType:      appType,
		CurrentPage:  currentPage,
		TemplateName: templateName,
		TestCases:    data,
		LocalURL:     hc.server.LocalURL,
		MobileURL:    hc.server.MobileURL,
		Navigation:   hc.server.registry.GetNavigation(r.URL.Path),
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// renderShowcase is a helper specifically for showcase pages
func (hc *HandlerContext) renderShowcase(w http.ResponseWriter, r *http.Request, platform, appType string, examples interface{}) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	hc.renderPage(w, r, platform, appType, "showcase", "showcase", examples)
}

// renderUISchemaBasedForm renders a form generated from compiled JSON Schema + UI Schema
func (hc *HandlerContext) renderUISchemaBasedForm(w http.ResponseWriter, r *http.Request, platform, appType string, compiledSchema *jsonschema.Schema, uiSchemaJSON string) {
	hc.renderUISchemaBasedFormWithErrors(w, r, platform, appType, compiledSchema, uiSchemaJSON, nil, nil)
}

// renderUISchemaBasedFormWithErrors renders a form with validation errors and pre-filled data
func (hc *HandlerContext) renderUISchemaBasedFormWithErrors(w http.ResponseWriter, r *http.Request, platform, appType string, compiledSchema *jsonschema.Schema, uiSchemaJSON string, formData map[string]interface{}, validationErrors schema.ValidationErrors) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)

	// Parse UI Schema
	uiSchema, err := schema.ParseUISchema(uiSchemaJSON)
	if err != nil {
		log.Printf("UI Schema parse error: %v", err)
		http.Error(w, "Failed to parse UI schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate form HTML from UI Schema + compiled JSON Schema with validation errors
	formHTML := uiSchema.GenerateFormHTMLWithData(compiledSchema, formData, validationErrors)

	// Render with UI schema-generated form
	err = hc.server.templates.ExecuteTemplate(w, "base", PageData{
		Platform:         platform,
		AppType:          appType,
		CurrentPage:      "custom", // Set to "custom" for menu highlighting (UI Schema forms are the "Custom" pages)
		TemplateName:     "schema_form",
		SchemaFormHTML:   formHTML,
		FormData:         formData,
		ValidationErrors: validationErrors,
		LocalURL:         hc.server.LocalURL,
		MobileURL:        hc.server.MobileURL,
		Navigation:       hc.server.registry.GetNavigation(r.URL.Path),
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// renderSuccess renders a success page with the generated URL
func (hc *HandlerContext) renderSuccess(w http.ResponseWriter, r *http.Request, platform, appType, generatedURL string) {
	err := hc.server.templates.ExecuteTemplate(w, "base", PageData{
		Platform:     platform,
		AppType:      appType,
		CurrentPage:  "custom",
		TemplateName: "success",
		GeneratedURL: generatedURL,
		LocalURL:     hc.server.LocalURL,
		MobileURL:    hc.server.MobileURL,
		Navigation:   hc.server.registry.GetNavigation(r.URL.Path),
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

