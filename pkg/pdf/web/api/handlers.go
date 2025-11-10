package api

import (
	"log"
	"net/http"
	"path/filepath"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
	"github.com/joeblew999/wellknown/pkg/pdf/commands"
	"github.com/joeblew999/wellknown/pkg/pdf/web/httputil"
)

// Handler handles HTTP requests for the PDF form API
type Handler struct {
	config *pdfform.Config
}

// NewHandler creates a new API handler
func NewHandler(config *pdfform.Config) *Handler {
	return &Handler{config: config}
}

// HandleBrowse handles the browse API request
func (h *Handler) HandleBrowse(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "GET") {
		return
	}

	state := r.URL.Query().Get("state")

	result, err := commands.Browse(commands.BrowseOptions{
		CatalogPath: h.config.CatalogFilePath(),
		State:       state,
	})

	if err != nil {
		log.Printf("Browse error: %v", err)
		httputil.RespondInternalError(w, err)
		return
	}

	httputil.RespondJSONOK(w, result)
}

// HandleDownload handles the download API request (JSON only)
func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	// Parse JSON request body
	var req struct {
		FormCode string `json:"form_code"`
	}
	if err := httputil.DecodeJSONBody(w, r, &req); err != nil {
		return
	}

	if req.FormCode == "" {
		httputil.RespondBadRequest(w, "form_code is required")
		return
	}

	outputDir := h.config.DownloadsPath()

	result, err := commands.Download(commands.DownloadOptions{
		CatalogPath: h.config.CatalogFilePath(),
		FormCode:    req.FormCode,
		OutputDir:   outputDir,
	})

	if err != nil {
		log.Printf("Download error: %v", err)
		httputil.RespondInternalError(w, err)
		return
	}

	httputil.RespondJSONOK(w, result)
}

// HandleInspect handles the inspect API request (JSON only)
func (h *Handler) HandleInspect(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	// Parse JSON request body
	var req struct {
		PDFPath string `json:"pdf_path"`
	}
	if err := httputil.DecodeJSONBody(w, r, &req); err != nil {
		return
	}

	if req.PDFPath == "" {
		httputil.RespondBadRequest(w, "pdf_path is required")
		return
	}

	pdfPath := req.PDFPath
	// If relative path, make it relative to downloads dir
	if !filepath.IsAbs(pdfPath) {
		pdfPath = filepath.Join(h.config.DownloadsPath(), pdfPath)
	}

	outputDir := h.config.TemplatesPath()

	result, err := commands.Inspect(commands.InspectOptions{
		PDFPath:   pdfPath,
		OutputDir: outputDir,
	})

	if err != nil {
		log.Printf("Inspect error: %v", err)
		httputil.RespondInternalError(w, err)
		return
	}

	httputil.RespondJSONOK(w, result)
}

// HandleFill handles the fill API request (JSON only)
func (h *Handler) HandleFill(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	// Parse JSON request body
	var req struct {
		CaseID  string `json:"case_id"`
		Flatten bool   `json:"flatten"`
	}
	if err := httputil.DecodeJSONBody(w, r, &req); err != nil {
		return
	}

	if req.CaseID == "" {
		httputil.RespondBadRequest(w, "case_id is required")
		return
	}

	// Find case file
	casePath, err := commands.FindCaseByID(req.CaseID, h.config.DataDir)
	if err != nil {
		httputil.RespondInternalError(w, err)
		return
	}

	if casePath == "" {
		httputil.RespondNotFound(w, "Case not found")
		return
	}

	outputDir := h.config.OutputsPath()

	result, err := commands.FillFromCase(casePath, outputDir, req.Flatten)
	if err != nil {
		log.Printf("Fill error: %v", err)
		httputil.RespondInternalError(w, err)
		return
	}

	httputil.RespondJSONOK(w, result)
}

// HandleListCases lists all available cases
func (h *Handler) HandleListCases(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "GET") {
		return
	}

	entityName := r.URL.Query().Get("entity")

	cases, err := commands.ListCases(h.config.DataDir, entityName)
	if err != nil {
		httputil.RespondInternalError(w, err)
		return
	}

	// Load case metadata for each case
	type CaseInfo struct {
		Path     string               `json:"path"`
		Metadata pdfform.CaseMetadata `json:"metadata"`
		FormCode string               `json:"form_code"`
	}

	var caseInfos []CaseInfo
	for _, casePath := range cases {
		c, err := commands.LoadCase(casePath)
		if err != nil {
			log.Printf("Failed to load case %s: %v", casePath, err)
			continue
		}

		caseInfos = append(caseInfos, CaseInfo{
			Path:     casePath,
			Metadata: c.Metadata,
			FormCode: c.FormReference.FormCode,
		})
	}

	httputil.RespondJSONOK(w, caseInfos)
}

// HandleCreateCase creates a new case (JSON only)
func (h *Handler) HandleCreateCase(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	// Parse JSON request body
	var req struct {
		FormCode   string `json:"form_code"`
		CaseName   string `json:"case_name"`
		EntityName string `json:"entity_name"`
	}
	if err := httputil.DecodeJSONBody(w, r, &req); err != nil {
		return
	}

	if req.FormCode == "" {
		httputil.RespondBadRequest(w, "form_code is required")
		return
	}
	if req.CaseName == "" {
		httputil.RespondBadRequest(w, "case_name is required")
		return
	}
	if req.EntityName == "" {
		httputil.RespondBadRequest(w, "entity_name is required")
		return
	}

	c, casePath, err := commands.CreateCase(req.FormCode, req.CaseName, req.EntityName, h.config.DataDir)
	if err != nil {
		log.Printf("Create case error: %v", err)
		httputil.RespondInternalError(w, err)
		return
	}

	response := map[string]interface{}{
		"case":      c,
		"case_path": casePath,
	}

	httputil.RespondJSONOK(w, response)
}

// HandleLoadCase loads a case and returns its data
func (h *Handler) HandleLoadCase(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "GET") {
		return
	}

	caseID, ok := httputil.GetRequiredQueryParam(w, r, "case_id")
	if !ok {
		return
	}

	// Find case file
	casePath, err := commands.FindCaseByID(caseID, h.config.DataDir)
	if err != nil {
		httputil.RespondInternalError(w, err)
		return
	}

	if casePath == "" {
		httputil.RespondNotFound(w, "Case not found")
		return
	}

	c, err := commands.LoadCase(casePath)
	if err != nil {
		httputil.RespondInternalError(w, err)
		return
	}

	httputil.RespondJSONOK(w, c)
}

// RegisterRoutes registers all API routes on the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/browse", h.HandleBrowse)
	mux.HandleFunc("/api/download", h.HandleDownload)
	mux.HandleFunc("/api/inspect", h.HandleInspect)
	mux.HandleFunc("/api/fill", h.HandleFill)
	mux.HandleFunc("/api/cases/list", h.HandleListCases)
	mux.HandleFunc("/api/cases/create", h.HandleCreateCase)
	mux.HandleFunc("/api/cases/load", h.HandleLoadCase)
}
