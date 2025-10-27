package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// GCPSetupStatus tracks the current setup progress
type GCPSetupStatus struct {
	ProjectID    string `json:"project_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	ProjectDone  bool   `json:"project_done"`
	APIDone      bool   `json:"api_done"`
	ConsentDone  bool   `json:"consent_done"`
	CredsDone    bool   `json:"creds_done"`
	EnvPath      string `json:"env_path"`
}

var gcpSetupStatus GCPSetupStatus

// RegisterGCPSetupRoutes registers all GCP setup routes with the given mux
func RegisterGCPSetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/tools/gcp-setup", handleGCPSetup)
	mux.HandleFunc("/api/gcp-setup/status", handleGCPSetupStatus)
	mux.HandleFunc("/api/gcp-setup/save-project", handleGCPSaveProject)
	mux.HandleFunc("/api/gcp-setup/save-creds", handleGCPSaveCreds)
	mux.HandleFunc("/api/gcp-setup/reset", handleGCPReset)
}

// handleGCPSetup renders the GCP setup page
func handleGCPSetup(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: GET %s", r.URL.Path)

	// Load current status from .env
	loadGCPEnvStatus()

	log.Printf("GCP Status loaded: ProjectID=%s, EnvPath=%s", gcpSetupStatus.ProjectID, gcpSetupStatus.EnvPath)

	// Render template using base template with navigation
	err := Templates.ExecuteTemplate(w, "base", PageData{
		Platform:     "tools",
		AppType:      "gcp-setup",
		CurrentPage:  "gcp-setup",
		TemplateName: "gcp_tool",
		GCPStatus:    gcpSetupStatus,
		LocalURL:     LocalURL,
		MobileURL:    MobileURL,
		Navigation:   GetNavigation(r.URL.Path),
	})
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// handleGCPSetupStatus returns current setup status as JSON
func handleGCPSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	loadGCPEnvStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gcpSetupStatus)
}

// handleGCPSaveProject saves project ID and name
func handleGCPSaveProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProjectID   string `json:"project_id"`
		ProjectName string `json:"project_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update status
	gcpSetupStatus.ProjectID = req.ProjectID
	gcpSetupStatus.ProjectDone = true

	// Save to .env
	if err := saveGCPEnvFile(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleGCPSaveCreds saves OAuth credentials
func handleGCPSaveCreds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update status
	gcpSetupStatus.ClientID = req.ClientID
	gcpSetupStatus.ClientSecret = req.ClientSecret
	gcpSetupStatus.CredsDone = true

	// Save to .env
	if err := saveGCPEnvFile(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleGCPReset resets the setup (deletes .env file)
func handleGCPReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	envPath := getGCPEnvPath()
	if err := os.Remove(envPath); err != nil && !os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reset status
	gcpSetupStatus = GCPSetupStatus{EnvPath: envPath}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// getGCPEnvPath returns the path to the .env file
func getGCPEnvPath() string {
	// Try to find project root by looking for go.mod
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
		return ".env"
	}

	// Walk up directories to find project root (where go.mod exists)
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Found project root - use root .env for all cloud providers
			return filepath.Join(dir, ".env")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			log.Printf("Warning: Could not find project root (go.mod), using current directory")
			return filepath.Join(cwd, ".env")
		}
		dir = parent
	}
}

// loadGCPEnvStatus loads status from .env file using EnvManager
func loadGCPEnvStatus() {
	em, err := NewEnvManager()
	if err != nil {
		log.Printf("Warning: Could not create EnvManager: %v", err)
		return
	}

	gcpSetupStatus.EnvPath = em.GetFilePath()

	// Get GCP section variables
	vars := em.GetSection("Google Cloud Platform (GCP)")

	if projectID, ok := vars["GCP_PROJECT_ID"]; ok && projectID != "" {
		gcpSetupStatus.ProjectID = projectID
		gcpSetupStatus.ProjectDone = true
	}

	if clientID, ok := vars["GOOGLE_CLIENT_ID"]; ok {
		gcpSetupStatus.ClientID = clientID
	}

	if clientSecret, ok := vars["GOOGLE_CLIENT_SECRET"]; ok {
		gcpSetupStatus.ClientSecret = clientSecret
		gcpSetupStatus.CredsDone = (gcpSetupStatus.ClientID != "" && clientSecret != "")
	}
}

// saveGCPEnvFile saves the current status to .env file using EnvManager
// Only updates the GCP section, leaves other sections (Cloudflare, Fly.io) untouched
func saveGCPEnvFile() error {
	em, err := NewEnvManager()
	if err != nil {
		return fmt.Errorf("failed to create EnvManager: %w", err)
	}

	// Build GCP variables map
	gcpVars := map[string]string{
		"GCP_PROJECT_ID":          gcpSetupStatus.ProjectID,
		"GOOGLE_CLIENT_ID":        gcpSetupStatus.ClientID,
		"GOOGLE_CLIENT_SECRET":    gcpSetupStatus.ClientSecret,
		"GOOGLE_REDIRECT_URL":     "http://localhost:8090/auth/google/callback",
	}

	// Update only the GCP section
	if err := em.UpdateSection("Google Cloud Platform (GCP)", gcpVars); err != nil {
		return fmt.Errorf("failed to update GCP section: %w", err)
	}

	return nil
}
