package server

import (
	"html/template"
)

// Package-level variables for templates and URLs
var (
	Templates *template.Template
	LocalURL  string
	MobileURL string
)

// PageData holds all data needed to render a page
type PageData struct {
	Platform         string
	AppType          string
	CurrentPage      string            // "custom" or "showcase"
	TemplateName     string            // Template to render within base
	IsStub           bool              // True for stub/placeholder pages
	GeneratedURL     string            // Generated deep link URL (for success pages)
	Event            interface{}       // Platform-specific event (for form pre-fill) - can be nil
	Error            string            // Error message (if any)
	TestCases        interface{}       // Platform-specific examples
	LocalURL         string            // Desktop URL for QR codes
	MobileURL        string            // Mobile URL for QR codes
	SchemaFormHTML   template.HTML     // Dynamically generated form HTML from JSON Schema
	FormData         map[string]interface{} // Form data for pre-filling after validation errors
	ValidationErrors ValidationErrors  // Field-level validation errors
}
