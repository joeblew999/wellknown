package gui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/joeblew999/wellknown/pkg/pdf/commands"
	"github.com/joeblew999/wellknown/pkg/pdf/web/httputil"
	"github.com/starfederation/datastar-go/datastar"
)

// HandleSSE streams events to the browser using Server-Sent Events
// This endpoint streams all events from the commands event bus to connected clients
// It intelligently converts command events into appropriate Datastar signal updates
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Subscribe to all events
	eventChan := commands.Subscribe("*")
	defer commands.Unsubscribe(eventChan)

	// Create context with cancellation
	ctx := r.Context()

	// Create SSE generator
	sse := datastar.NewSSE(w, r)
	log.Printf("🔌 SSE client connected from %s", r.RemoteAddr)

	// Keep connection alive with periodic heartbeat
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("🔌 SSE client disconnected from %s", r.RemoteAddr)
			return

		case event := <-eventChan:
			// Convert event to Datastar signals based on event type
			signals := h.eventToSignals(event)
			if signals == nil {
				// Event doesn't need UI update
				continue
			}

			// Send signals to browser - Datastar will merge these into existing signals
			// without affecting form inputs that aren't in this update
			err := sse.MarshalAndPatchSignals(signals)
			if err != nil {
				log.Printf("❌ Error sending SSE signals: %v", err)
				return
			}
			log.Printf("📡 Sent SSE signal update: %s", event.Type)

		case <-ticker.C:
			// Send heartbeat comment to keep connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// eventToSignals converts a command event into appropriate Datastar signals
// Returns nil if the event doesn't require UI updates
func (h *Handler) eventToSignals(event *commands.Event) map[string]interface{} {
	switch event.Type {
	// Download events
	case commands.EventDownloadStarted:
		formCode := getStringFromData(event.Data, "form_code")
		return map[string]interface{}{
			"downloading": true,
			"status":      fmt.Sprintf("Downloading %s...", formCode),
			"error":       "",
		}
	case commands.EventDownloadProgress:
		formCode := getStringFromData(event.Data, "form_code")
		stage := getStringFromData(event.Data, "stage")
		progress := getFloatFromData(event.Data, "progress")
		return map[string]interface{}{
			"downloading": true,
			"status":      fmt.Sprintf("Downloading %s: %s (%.0f%%)", formCode, stage, progress*100),
			"error":       "",
		}
	case commands.EventDownloadCompleted:
		pdfPath := getStringFromData(event.Data, "pdf_path")
		formName := getStringFromData(event.Data, "form_name")
		return map[string]interface{}{
			"downloading": false,
			"status":      fmt.Sprintf("Downloaded %s successfully! PDF saved to: %s", formName, pdfPath),
			"error":       "",
		}
	case commands.EventDownloadError:
		errorMsg := ""
		if event.Error != nil {
			errorMsg = event.Error.Error()
		}
		return map[string]interface{}{
			"downloading": false,
			"status":      "",
			"error":       fmt.Sprintf("Download failed: %s", errorMsg),
		}

	// Inspect events
	case commands.EventInspectStarted:
		pdfPath := getStringFromData(event.Data, "pdf_path")
		return map[string]interface{}{
			"inspecting": true,
			"status":     fmt.Sprintf("Inspecting %s...", filepath.Base(pdfPath)),
			"error":      "",
		}
	case commands.EventInspectCompleted:
		templatePath := getStringFromData(event.Data, "template_path")
		fieldCount := getIntFromData(event.Data, "field_count")
		return map[string]interface{}{
			"inspecting":   false,
			"status":       fmt.Sprintf("Inspection complete! Found %d fields. Template saved to: %s", fieldCount, templatePath),
			"error":        "",
			"fieldCount":   fieldCount,
			"templatePath": templatePath,
		}
	case commands.EventInspectError:
		errorMsg := ""
		if event.Error != nil {
			errorMsg = event.Error.Error()
		}
		return map[string]interface{}{
			"inspecting":   false,
			"status":       "",
			"error":        fmt.Sprintf("Inspect failed: %s", errorMsg),
			"fieldCount":   0,
			"templatePath": "",
		}

	// Fill events
	case commands.EventFillStarted:
		dataPath := getStringFromData(event.Data, "data_path")
		return map[string]interface{}{
			"filling": true,
			"status":  fmt.Sprintf("Filling from %s...", filepath.Base(dataPath)),
			"error":   "",
		}
	case commands.EventFillCompleted:
		outputPath := getStringFromData(event.Data, "output_path")
		flattened := getBoolFromData(event.Data, "flattened")
		flattenMsg := ""
		if flattened {
			flattenMsg = " (flattened)"
		}
		return map[string]interface{}{
			"filling":    false,
			"status":     fmt.Sprintf("Fill complete%s! Output saved to: %s", flattenMsg, outputPath),
			"error":      "",
			"outputPath": outputPath,
		}
	case commands.EventFillError:
		errorMsg := ""
		if event.Error != nil {
			errorMsg = event.Error.Error()
		}
		return map[string]interface{}{
			"filling":    false,
			"status":     "",
			"error":      fmt.Sprintf("Fill failed: %s", errorMsg),
			"outputPath": "",
		}

	// Case events
	case commands.EventCaseCreated:
		caseID := getStringFromData(event.Data, "case_id")
		caseName := getStringFromData(event.Data, "case_name")
		casePath := getStringFromData(event.Data, "case_path")
		return map[string]interface{}{
			"creating": false,
			"status":   fmt.Sprintf("Case '%s' created successfully! ID: %s", caseName, caseID),
			"error":    "",
			"caseId":   caseID,
			"casePath": casePath,
		}
	case commands.EventCaseError:
		errorMsg := ""
		if event.Error != nil {
			errorMsg = event.Error.Error()
		}
		return map[string]interface{}{
			"creating": false,
			"status":   "",
			"error":    fmt.Sprintf("Case operation failed: %s", errorMsg),
		}

	default:
		// Event doesn't need UI update
		return nil
	}
}

// Helper functions to safely extract values from event data
func getStringFromData(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntFromData(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

func getFloatFromData(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

func getBoolFromData(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// getSignalName converts event type to camelCase signal name
// e.g., "download.progress" -> "downloadProgress"
func getSignalName(eventType string) string {
	// Simple conversion for now - can be enhanced later
	// This maps event types like "download.progress" to "downloadProgress"
	result := ""
	capitalizeNext := false

	for i, c := range eventType {
		if c == '.' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			result += string(c - 32) // Convert to uppercase
			capitalizeNext = false
		} else if i == 0 {
			result += string(c)
		} else {
			result += string(c)
		}
	}

	return result
}

// HandleDownloadAction handles the download form submission
// Triggers download asynchronously - UI updates come via SSE from event system
func (h *Handler) HandleDownloadAction(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	formCode, ok := httputil.GetRequiredFormValue(w, r, "selectedForm")
	if !ok {
		return
	}

	// Get catalog path from config
	catalogPath := h.config.CatalogFilePath()
	outputDir := filepath.Join(h.config.DownloadsPath(), formCode)

	// Execute download asynchronously - events will update UI via SSE
	go func() {
		opts := commands.DownloadOptions{
			CatalogPath: catalogPath,
			FormCode:    formCode,
			OutputDir:   outputDir,
		}
		_, err := commands.Download(opts)
		if err != nil {
			log.Printf("❌ Download failed for %s: %v", formCode, err)
		} else {
			log.Printf("✅ Download completed for %s", formCode)
		}
	}()

	// Respond immediately - SSE will handle UI updates
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Download started")
}

// getFormsData is a helper to get forms data for both JSON and HTML responses
func (h *Handler) getFormsData(r *http.Request) (map[string]interface{}, error) {
	state := r.URL.Query().Get("state")

	// If no state specified, get all forms from all states (zero-input workflow)
	if state == "" {
		// First, get list of states
		statesResult, err := commands.Browse(commands.BrowseOptions{
			CatalogPath: h.config.CatalogFilePath(),
			State:       "",
		})
		if err != nil {
			return nil, err
		}

		// Now get forms from all states and combine them
		var allForms []interface{}
		for _, st := range statesResult.States {
			stateResult, err := commands.Browse(commands.BrowseOptions{
				CatalogPath: h.config.CatalogFilePath(),
				State:       st,
			})
			if err != nil {
				continue // Skip states with errors
			}
			// Convert each form to interface{} and append
			for _, form := range stateResult.Forms {
				allForms = append(allForms, form)
			}
		}

		return map[string]interface{}{
			"Success": true,
			"Count":   len(allForms),
			"Forms":   allForms,
		}, nil
	}

	// If state specified, get forms for that state only
	result, err := commands.Browse(commands.BrowseOptions{
		CatalogPath: h.config.CatalogFilePath(),
		State:       state,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"Success": true,
		"Count":   len(result.Forms),
		"Forms":   result.Forms,
	}, nil
}

// HandleGetForms returns available forms from the catalog for auto-populating selects
// This enables zero-input workflow - user just selects from available options
func (h *Handler) HandleGetForms(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "GET") {
		return
	}

	result, err := h.getFormsData(r)
	if err != nil {
		log.Printf("❌ Failed to browse forms: %v", err)
		httputil.RespondInternalError(w, err)
		return
	}

	// Return JSON response (lowercase keys for JSON)
	httputil.RespondJSONOK(w, map[string]interface{}{
		"success": result["Success"],
		"count":   result["Count"],
		"forms":   result["Forms"],
	})
}

// HandleInspectAction handles PDF inspection
// Triggers inspect asynchronously - UI updates come via SSE from event system
func (h *Handler) HandleInspectAction(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	pdfPath, ok := httputil.GetRequiredFormValue(w, r, "pdfPath")
	if !ok {
		return
	}

	// Get output directory from config
	outputDir := h.config.TemplatesPath()

	// Execute inspect asynchronously - events will update UI via SSE
	go func() {
		opts := commands.InspectOptions{
			PDFPath:   pdfPath,
			OutputDir: outputDir,
		}
		_, err := commands.Inspect(opts)
		if err != nil {
			log.Printf("❌ Inspect failed for %s: %v", pdfPath, err)
		} else {
			log.Printf("✅ Inspect completed for %s", pdfPath)
		}
	}()

	// Respond immediately - SSE will handle UI updates
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Inspect started")
}

// HandleFillAction handles PDF filling
// Triggers fill asynchronously - UI updates come via SSE from event system
func (h *Handler) HandleFillAction(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	// Get parameters from form
	dataPath, ok := httputil.GetRequiredFormValue(w, r, "dataPath")
	if !ok {
		return
	}

	// Optional flatten parameter
	flatten := r.FormValue("flatten") == "true"

	// Get output directory from config
	outputDir := h.config.DownloadsPath()

	// Execute fill asynchronously - events will update UI via SSE
	go func() {
		opts := commands.FillOptions{
			DataPath:  dataPath,
			OutputDir: outputDir,
			Flatten:   flatten,
		}
		_, err := commands.Fill(opts)
		if err != nil {
			log.Printf("❌ Fill failed for %s: %v", dataPath, err)
		} else {
			log.Printf("✅ Fill completed for %s", dataPath)
		}
	}()

	// Respond immediately - SSE will handle UI updates
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Fill started")
}

// HandleListCases returns list of available cases for selection
func (h *Handler) HandleListCases(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "GET") {
		return
	}

	// Get cases directory from config
	casesDir := h.config.CasesPath()
	entityName := r.URL.Query().Get("entity") // Optional filter

	// Call commands to list cases
	caseIDs, err := commands.ListCases(casesDir, entityName)
	if err != nil {
		log.Printf("❌ Failed to list cases: %v", err)
		httputil.RespondInternalError(w, err)
		return
	}

	// Return JSON response
	httputil.RespondJSONOK(w, map[string]interface{}{
		"success": true,
		"count":   len(caseIDs),
		"cases":   caseIDs,
	})
}

// HandleCreateCase creates a new test case
func (h *Handler) HandleCreateCase(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	caseName, ok := httputil.GetRequiredFormValue(w, r, "caseName")
	if !ok {
		return
	}

	formCode := r.FormValue("formCode")
	entityName := r.FormValue("entityName")

	// Get cases directory from config
	casesDir := h.config.CasesPath()

	// Create SSE connection for this request
	sse := datastar.NewSSE(w, r)

	// Send "creating" signal immediately
	signals := map[string]interface{}{
		"creating": true,
		"status":   fmt.Sprintf("Creating case %s...", caseName),
		"error":    "",
	}
	sse.MarshalAndPatchSignals(signals)

	// Call commands to create case (returns 3 values)
	caseObj, casePath, err := commands.CreateCase(formCode, caseName, entityName, casesDir)
	if err != nil {
		log.Printf("❌ Failed to create case %s: %v", caseName, err)
		// Send error signal
		signals = map[string]interface{}{
			"creating": false,
			"status":   "",
			"error":    err.Error(),
			"casePath": "",
		}
		sse.MarshalAndPatchSignals(signals)
	} else {
		log.Printf("✅ Case created: %s at %s", caseName, casePath)
		// Send success signal
		signals = map[string]interface{}{
			"creating": false,
			"status":   fmt.Sprintf("Case created successfully! Path: %s", casePath),
			"error":    "",
			"caseId":   caseObj.Metadata.CaseID,
			"casePath": casePath,
		}
		sse.MarshalAndPatchSignals(signals)
	}
}

// HandleLoadCase loads an existing test case
func (h *Handler) HandleLoadCase(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "GET") {
		return
	}

	casePath, ok := httputil.GetRequiredFormValue(w, r, "casePath")
	if !ok {
		return
	}

	// Call commands to load case (takes just casePath)
	caseObj, err := commands.LoadCase(casePath)
	if err != nil {
		log.Printf("❌ Failed to load case %s: %v", casePath, err)
		httputil.RespondInternalError(w, err)
		return
	}

	// Return JSON response with case data
	httputil.RespondJSONOK(w, map[string]interface{}{
		"success": true,
		"case":    caseObj,
	})
}

// HandleSaveCase saves changes to an existing test case
func (h *Handler) HandleSaveCase(w http.ResponseWriter, r *http.Request) {
	if !httputil.ValidateMethod(w, r, "POST") {
		return
	}

	casePath, ok := httputil.GetRequiredFormValue(w, r, "casePath")
	if !ok {
		return
	}

	// First load the existing case
	caseObj, err := commands.LoadCase(casePath)
	if err != nil {
		httputil.RespondInternalError(w, err)
		return
	}

	// Parse updates from request body
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		httputil.RespondBadRequest(w, "Invalid JSON data")
		return
	}

	// TODO: Apply updates to caseObj fields
	// This requires knowing the Case struct layout
	// For now, we'll just save the existing case

	// Create SSE connection for this request
	sse := datastar.NewSSE(w, r)

	// Send "saving" signal immediately
	signals := map[string]interface{}{
		"saving": true,
		"status": "Saving case...",
		"error":  "",
	}
	sse.MarshalAndPatchSignals(signals)

	// Call commands to save case (returns just error)
	err = commands.SaveCase(caseObj, casePath)
	if err != nil {
		log.Printf("❌ Failed to save case %s: %v", casePath, err)
		// Send error signal
		signals = map[string]interface{}{
			"saving": false,
			"status": "",
			"error":  err.Error(),
		}
		sse.MarshalAndPatchSignals(signals)
	} else {
		log.Printf("✅ Case saved: %s", casePath)
		// Send success signal
		signals = map[string]interface{}{
			"saving": false,
			"status": fmt.Sprintf("Case saved successfully! Path: %s", casePath),
			"error":  "",
		}
		sse.MarshalAndPatchSignals(signals)
	}
}
