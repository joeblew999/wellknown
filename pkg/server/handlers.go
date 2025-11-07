package server

import (
	"html/template"

	"github.com/joeblew999/wellknown/pkg/schema"
)

// NavLink represents a single navigation link
type NavLink struct {
	Label    string
	URL      string
	IsActive bool
}

// NavSection represents a navigation section with multiple links
type NavSection struct {
	Title string
	Links []NavLink
}

// PageData holds all data needed to render a page
type PageData struct {
	Platform         string
	AppType          string
	CurrentPage      string            // "custom" or "examples"
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
	ValidationErrors schema.ValidationErrors // Field-level validation errors
	Navigation       []NavSection      // Server-generated navigation
	GCPStatus        GCPSetupStatus    // GCP setup status (for tools/gcp-setup page)
	URLPrefix        string            // URL prefix when embedded (e.g., "/demo" in PocketBase)
}
