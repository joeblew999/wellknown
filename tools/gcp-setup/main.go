package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
)

//go:embed templates/*.html
var templatesFS embed.FS

var tmpl *template.Template

// SetupStatus tracks the current setup progress
type SetupStatus struct {
	ProjectID     string `json:"project_id"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	ProjectDone   bool   `json:"project_done"`
	APIDone       bool   `json:"api_done"`
	ConsentDone   bool   `json:"consent_done"`
	CredsDone     bool   `json:"creds_done"`
	EnvPath       string `json:"env_path"`
}

var currentStatus SetupStatus

func main() {
	// Parse flags
	webMode := flag.Bool("web", false, "Start web-based setup wizard")
	cliMode := flag.Bool("cli", false, "Run CLI-based automated setup (requires gcloud)")
	flag.Parse()

	// Default to web mode if no flags specified
	if !*webMode && !*cliMode {
		*webMode = true
	}

	// Run appropriate mode
	if *webMode {
		runWebMode()
	} else {
		runCLIMode()
	}
}

func runWebMode() {
	// Parse templates
	var err error
	tmpl, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatal("Failed to parse templates:", err)
	}

	// Load existing .env if exists
	loadEnvStatus()

	// Routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/save-project", handleSaveProject)
	http.HandleFunc("/api/save-creds", handleSaveCreds)
	http.HandleFunc("/api/generate-env", handleGenerateEnv)
	http.HandleFunc("/api/delete-project", handleDeleteProject)
	http.HandleFunc("/api/reset", handleReset)
	http.HandleFunc("/api/list-projects", handleListProjects)

	fmt.Println("🚀 Wellknown GCP Setup Web Interface")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("📱 Open: http://localhost:3030")
	fmt.Println()
	fmt.Println("The web interface will guide you through:")
	fmt.Println("  1. Creating GCP project")
	fmt.Println("  2. Enabling APIs")
	fmt.Println("  3. Configuring OAuth")
	fmt.Println("  4. Generating .env file")
	fmt.Println()

	if err := http.ListenAndServe(":3030", nil); err != nil {
		log.Fatal(err)
	}
}

func runCLIMode() {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("❌ Set GCP_PROJECT_ID environment variable")
	}

	ctx := context.Background()

	fmt.Println("🚀 Automated Google Cloud Project Setup (CLI Mode)")
	fmt.Println("===================================================")
	fmt.Printf("📁 Config file: %s\n\n", getEnvPath())

	// Step 1: Create project
	fmt.Println("📦 Step 1: Creating project...")
	if err := createProject(ctx, projectID); err != nil {
		log.Printf("⚠️  Project creation skipped (might already exist): %v\n", err)
	} else {
		fmt.Println("✅ Project created")
		time.Sleep(3 * time.Second) // Wait for project to propagate
	}
	fmt.Println()

	// Step 2: Enable APIs
	fmt.Println("📡 Step 2: Enabling required APIs...")
	if err := enableAPIs(ctx, projectID); err != nil {
		log.Fatalf("❌ Failed to enable APIs: %v\n", err)
	}
	fmt.Println("✅ All APIs enabled")
	fmt.Println()

	// Step 3: Create OAuth consent screen
	fmt.Println("🔐 Step 3: Configuring OAuth consent screen...")
	if err := configureOAuthConsent(ctx, projectID); err != nil {
		log.Printf("⚠️  OAuth consent configuration: %v\n", err)
		fmt.Println("   You may need to configure this manually in the console")
	} else {
		fmt.Println("✅ OAuth consent screen configured")
	}
	fmt.Println()

	// Step 4: Create OAuth client credentials
	fmt.Println("🔑 Step 4: Creating OAuth 2.0 client credentials...")
	clientID, clientSecret, err := createOAuthClient(ctx, projectID)
	if err != nil {
		log.Printf("⚠️  OAuth client creation: %v\n", err)
		printManualInstructions(projectID)
		return
	}
	fmt.Println("✅ OAuth client created")
	fmt.Println()

	// Step 5: Generate .env file
	fmt.Println("📝 Step 5: Generating .env file...")
	if err := generateEnvFile(projectID, clientID, clientSecret); err != nil {
		log.Printf("⚠️  Failed to generate .env: %v\n", err)
	} else {
		fmt.Println("✅ .env file created: pb/base/.env")
	}
	fmt.Println()

	// Success summary
	fmt.Println("🎉 Setup Complete!")
	fmt.Println("==================")
	fmt.Println("Next steps:")
	fmt.Println("1. cd pb/base")
	fmt.Println("2. source .env")
	fmt.Println("3. go run main.go serve")
	fmt.Println("4. Open http://localhost:8090")
	fmt.Println()
	fmt.Println("📋 Credentials:")
	fmt.Printf("   GOOGLE_CLIENT_ID=%s\n", clientID)
	fmt.Printf("   GOOGLE_CLIENT_SECRET=%s\n", clientSecret)
}

func createProject(ctx context.Context, projectID string) error {
	crmService, err := cloudresourcemanager.NewService(ctx,
		option.WithScopes("https://www.googleapis.com/auth/cloud-platform"))
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	project := &cloudresourcemanager.Project{
		ProjectId: projectID,
		Name:      "Wellknown Calendar OAuth",
	}

	_, err = crmService.Projects.Create(project).Context(ctx).Do()
	return err
}

func enableAPIs(ctx context.Context, projectID string) error {
	svc, err := serviceusage.NewService(ctx,
		option.WithScopes("https://www.googleapis.com/auth/cloud-platform"))
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	apis := []string{
		"calendar-json.googleapis.com",        // Google Calendar API
		"oauth2.googleapis.com",               // OAuth 2.0 API
		"iamcredentials.googleapis.com",       // IAM Credentials API
		"cloudresourcemanager.googleapis.com", // Resource Manager API
	}

	for _, api := range apis {
		serviceName := fmt.Sprintf("projects/%s/services/%s", projectID, api)
		fmt.Printf("   ⏳ Enabling %s...\n", api)

		_, err := svc.Services.Enable(serviceName, &serviceusage.EnableServiceRequest{}).
			Context(ctx).
			Do()
		if err != nil {
			return fmt.Errorf("failed to enable %s: %w", api, err)
		}

		time.Sleep(1 * time.Second) // Rate limiting
	}

	return nil
}

func configureOAuthConsent(ctx context.Context, projectID string) error {
	// Note: The OAuth consent screen configuration requires manual setup
	// via the GCP Console or using the Admin SDK which requires additional setup.
	// For now, we'll provide instructions if this fails.

	// This is a placeholder - the actual OAuth brand creation requires
	// the Cloud Identity API which has complex authentication requirements.
	// Most users will need to configure this manually once.

	return fmt.Errorf("OAuth consent screen must be configured manually (one-time setup)")
}

func createOAuthClient(ctx context.Context, projectID string) (clientID, clientSecret string, err error) {
	// Note: Creating OAuth clients programmatically is complex and requires
	// the OAuth2 API which is not fully exposed via the Go client libraries.
	//
	// The recommended approach is to use the gcloud CLI or REST API directly.
	// For this tool, we'll provide clear instructions for manual setup.

	return "", "", fmt.Errorf("OAuth client creation requires manual setup via GCP Console")
}

func generateEnvFile(projectID, clientID, clientSecret string) error {
	envPath := "../../pb/base/.env"

	content := fmt.Sprintf(`# Google OAuth Configuration
# Generated by: make gcp-setup
# Project: %s

GOOGLE_CLIENT_ID=%s
GOOGLE_CLIENT_SECRET=%s
GOOGLE_REDIRECT_URL=http://localhost:8090/auth/google/callback

# Pocketbase Admin (optional)
PB_ADMIN_EMAIL=admin@example.com
PB_ADMIN_PASSWORD=changeme123
`, projectID, clientID, clientSecret)

	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write .env: %w", err)
	}

	return nil
}

func printManualInstructions(projectID string) {
	fmt.Println()
	fmt.Println("📋 Manual OAuth Setup Required")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("⚠️  Automated OAuth client creation is not available.")
	fmt.Println("   Please follow these steps:")
	fmt.Println()
	fmt.Printf("1. Open: https://console.cloud.google.com/apis/credentials/consent?project=%s\n", projectID)
	fmt.Println("   - User Type: External")
	fmt.Println("   - App name: Wellknown Calendar")
	fmt.Println("   - Support email: Your email")
	fmt.Println("   - Click 'Save and Continue' through all steps")
	fmt.Println()
	fmt.Printf("2. Open: https://console.cloud.google.com/apis/credentials?project=%s\n", projectID)
	fmt.Println("   - Click 'Create Credentials' → 'OAuth client ID'")
	fmt.Println("   - Application type: Web application")
	fmt.Println("   - Name: Wellknown PB Server")
	fmt.Println("   - Authorized redirect URIs:")
	fmt.Println("     • http://localhost:8090/auth/google/callback")
	fmt.Println("     • http://127.0.0.1:8090/auth/google/callback")
	fmt.Println("   - Click 'Create'")
	fmt.Println()
	fmt.Println("3. Copy the credentials:")
	fmt.Println("   - Save Client ID as GOOGLE_CLIENT_ID")
	fmt.Println("   - Save Client Secret as GOOGLE_CLIENT_SECRET")
	fmt.Println()
	fmt.Println("4. Create pb/base/.env:")
	fmt.Println("   cp pb/base/.env.example pb/base/.env")
	fmt.Println("   # Edit .env with your credentials")
	fmt.Println()
	fmt.Println("5. Run the server:")
	fmt.Println("   cd pb/base")
	fmt.Println("   source .env")
	fmt.Println("   go run main.go serve")
}

// Web server handlers

func handleHome(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "index.html", currentStatus); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStatus)
}

func handleSaveProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.FormValue("project_id")
	if projectID == "" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	currentStatus.ProjectID = projectID
	currentStatus.ProjectDone = true
	saveEnvFile()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStatus)
}

func handleSaveCreds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentStatus.ClientID = r.FormValue("client_id")
	currentStatus.ClientSecret = r.FormValue("client_secret")
	currentStatus.CredsDone = true
	saveEnvFile()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStatus)
}

func handleGenerateEnv(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := saveEnvFile(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"path":   currentStatus.EnvPath,
	})
}

func loadEnvStatus() {
	envPath := getEnvPath()
	currentStatus.EnvPath = envPath

	file, err := os.Open(envPath)
	if err != nil {
		// File doesn't exist yet, use defaults
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "GCP_PROJECT_ID":
			currentStatus.ProjectID = value
			if value != "" && value != "your-project-id" {
				currentStatus.ProjectDone = true
			}
		case "GOOGLE_CLIENT_ID":
			currentStatus.ClientID = value
			if value != "" && !strings.Contains(value, "your-client") {
				currentStatus.CredsDone = true
			}
		case "GOOGLE_CLIENT_SECRET":
			currentStatus.ClientSecret = value
		}
	}
}

func saveEnvFile() error {
	envPath := getEnvPath()

	// Ensure directory exists
	dir := filepath.Dir(envPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	clientID := currentStatus.ClientID
	if clientID == "" {
		clientID = "your-client-id.apps.googleusercontent.com"
	}

	clientSecret := currentStatus.ClientSecret
	if clientSecret == "" {
		clientSecret = "your-client-secret"
	}

	projectID := currentStatus.ProjectID
	if projectID == "" {
		projectID = "your-project-id"
	}

	content := fmt.Sprintf(`# Google OAuth Configuration
# Generated by: tools/gcp-setup
# Project: %s

GCP_PROJECT_ID=%s
GOOGLE_CLIENT_ID=%s
GOOGLE_CLIENT_SECRET=%s
GOOGLE_REDIRECT_URL=http://localhost:8090/auth/google/callback

# Pocketbase Admin (optional - change these!)
PB_ADMIN_EMAIL=admin@example.com
PB_ADMIN_PASSWORD=changeme123
`, projectID, projectID, clientID, clientSecret)

	return os.WriteFile(envPath, []byte(content), 0600)
}

func getEnvPath() string {
	// Try to find project root by looking for go.mod
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
		return "pb/base/.env"
	}

	// Walk up directories to find project root (where go.mod exists)
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Found project root
			return filepath.Join(dir, "pb", "base", ".env")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			log.Printf("Warning: Could not find project root (go.mod), using relative path")
			return filepath.Join(cwd, "..", "..", "pb", "base", ".env")
		}
		dir = parent
	}
}

func deleteProject(ctx context.Context, projectID string) error {
	crmService, err := cloudresourcemanager.NewService(ctx,
		option.WithScopes("https://www.googleapis.com/auth/cloud-platform"))
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Delete the project
	_, err = crmService.Projects.Delete(projectID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

func handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := currentStatus.ProjectID
	if projectID == "" {
		http.Error(w, "No project ID set", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := deleteProject(ctx, projectID); err != nil {
		log.Printf("Failed to delete project: %v", err)

		// Return friendly error for authentication issues
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "not_authenticated",
			"message": "Project deletion requires gcloud authentication.",
			"hint":    "To delete projects programmatically, run 'gcloud auth application-default login', or delete manually via GCP Console.",
			"console": fmt.Sprintf("https://console.cloud.google.com/home/dashboard?project=%s", projectID),
		})
		return
	}

	// Reset status after deletion
	currentStatus = SetupStatus{
		EnvPath: getEnvPath(),
	}

	// Remove .env file
	os.Remove(getEnvPath())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Project %s deleted", projectID),
	})
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Reset status
	currentStatus = SetupStatus{
		EnvPath: getEnvPath(),
	}

	// Remove .env file
	envPath := getEnvPath()
	if err := os.Remove(envPath); err != nil {
		log.Printf("Failed to remove .env file: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Setup reset, .env file removed",
	})
}

type ProjectInfo struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	State       string `json:"state"`
	CreateTime  string `json:"create_time"`
}

func listProjects(ctx context.Context) ([]ProjectInfo, error) {
	crmService, err := cloudresourcemanager.NewService(ctx,
		option.WithScopes("https://www.googleapis.com/auth/cloud-platform"))
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	projects, err := crmService.Projects.List().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var result []ProjectInfo
	for _, p := range projects.Projects {
		result = append(result, ProjectInfo{
			ProjectID:   p.ProjectId,
			ProjectName: p.Name,
			State:       p.LifecycleState,
			CreateTime:  p.CreateTime,
		})
	}

	return result, nil
}

func handleListProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	projects, err := listProjects(ctx)
	if err != nil {
		log.Printf("Failed to list projects: %v", err)

		// Return friendly error message for missing credentials
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "not_authenticated",
			"message": "GCP API access requires gcloud authentication. This feature is only available in CLI mode.",
			"hint":    "Run 'gcloud auth application-default login' to authenticate, or use the manual web-based workflow instead.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"projects": projects,
		"count":    len(projects),
	})
}
