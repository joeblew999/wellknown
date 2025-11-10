package gui

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var templates *template.Template

// Handler handles HTTP requests for the PDF form web GUI
type Handler struct {
	config *pdfform.Config
}

// InitTemplates initializes the embedded templates
func InitTemplates() error {
	var err error
	templates, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return err
	}
	log.Printf("📄 Loaded %d GUI templates", len(templates.Templates()))
	return nil
}

// renderTemplate is a helper function to render a template to string
func renderTemplate(name string) (string, error) {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, name, nil)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// NewHandler creates a new GUI handler
func NewHandler(config *pdfform.Config) *Handler {
	return &Handler{config: config}
}

// HandleHome renders the home page with 5-step workflow
func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	html, err := renderTemplate("home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// HandleBrowse renders the browse forms page
func (h *Handler) HandleBrowse(w http.ResponseWriter, r *http.Request) {
	html, err := renderTemplate("browse.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// HandleDownload renders the download form page
func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	// Get forms data to include in page
	result, err := h.getFormsData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add title to the result map
	result["Title"] = "2️⃣ Download Form"

	var buf bytes.Buffer
	err = templates.ExecuteTemplate(&buf, "download.html", result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, buf.String())
}

// HandleInspect renders the inspect fields page
func (h *Handler) HandleInspect(w http.ResponseWriter, r *http.Request) {
	html, err := renderTemplate("inspect.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// HandleFill renders the fill form page
func (h *Handler) HandleFill(w http.ResponseWriter, r *http.Request) {
	html, err := renderTemplate("fill.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// HandleTest renders the test page
func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	html, err := renderTemplate("test.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// HandleDownloadData renders the download fragment with forms data
// This is called on page load via data-on-load="$$get('/gui/download-data')"
func (h *Handler) HandleDownloadData(w http.ResponseWriter, r *http.Request) {
	// Get forms data
	result, err := h.getFormsData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the fragment template
	var buf bytes.Buffer
	err = templates.ExecuteTemplate(&buf, "download_fragment", result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send as fragment merge (replaces the container content)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, buf.String())
}

// RegisterRoutes registers all GUI routes on the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Serve static files (Datastar, assets, etc.)
	staticSub, _ := fs.Sub(staticFS, "static")
	staticFileServer := http.FileServer(http.FS(staticSub))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFileServer))

	// Workflow pages
	mux.HandleFunc("/", h.HandleHome)
	mux.HandleFunc("/1-browse", h.HandleBrowse)
	mux.HandleFunc("/2-download", h.HandleDownload)
	mux.HandleFunc("/3-inspect", h.HandleInspect)
	mux.HandleFunc("/4-fill", h.HandleFill)
	mux.HandleFunc("/5-test", h.HandleTest)

	// GUI-specific API endpoints (use /gui/ prefix to avoid conflicts with /api/)
	mux.HandleFunc("/gui/events", h.HandleSSE)                 // SSE event stream
	mux.HandleFunc("/gui/forms", h.HandleGetForms)             // Get available forms (JSON)
	mux.HandleFunc("/gui/download-data", h.HandleDownloadData) // Get download fragment (HTML)
	mux.HandleFunc("/gui/download", h.HandleDownloadAction)    // Trigger download action
	mux.HandleFunc("/gui/inspect", h.HandleInspectAction)      // Trigger inspect action
	mux.HandleFunc("/gui/fill", h.HandleFillAction)            // Trigger fill action

	// Case management endpoints
	mux.HandleFunc("/gui/cases/list", h.HandleListCases)   // List all cases (JSON)
	mux.HandleFunc("/gui/cases/create", h.HandleCreateCase) // Create new case
	mux.HandleFunc("/gui/cases/load", h.HandleLoadCase)     // Load case data (JSON)
	mux.HandleFunc("/gui/cases/save", h.HandleSaveCase)     // Save case data
}
