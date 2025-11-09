package workflow

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/joeblew999/wellknown/pkg/env"
)

// Test SyncRegistryWorkflow with minimal config
func TestSyncRegistryWorkflow_Minimal(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create test registry
	registry := env.NewRegistry([]env.EnvVar{
		{Name: "TEST_VAR", Description: "Test variable", Default: "test", Group: "Test"},
	})

	// Run workflow
	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry:           registry,
		AppName:            "Test App",
		CreateSecretsFiles: false,
	})

	if err != nil {
		t.Fatalf("SyncRegistryWorkflow failed: %v", err)
	}

	// Check that env files were created
	if !fileExists(env.Local.FileName) {
		t.Errorf("Expected %s to be created", env.Local.FileName)
	}
	if !fileExists(env.Production.FileName) {
		t.Errorf("Expected %s to be created", env.Production.FileName)
	}

	// Check result
	if len(result.UpdatedFiles) != 2 {
		t.Errorf("Expected 2 updated files, got %d", len(result.UpdatedFiles))
	}
}

// Test SyncRegistryWorkflow with secrets file creation
func TestSyncRegistryWorkflow_WithSecrets(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create test registry with secrets
	registry := env.NewRegistry([]env.EnvVar{
		{Name: "PUBLIC_VAR", Description: "Public", Default: "public", Secret: false},
		{Name: "SECRET_VAR", Description: "Secret", Default: "secret", Secret: true},
	})

	// Run workflow with secrets creation
	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry:           registry,
		AppName:            "Test App",
		CreateSecretsFiles: true,
	})

	if err != nil {
		t.Fatalf("SyncRegistryWorkflow failed: %v", err)
	}

	// Check that secrets files were created
	if !fileExists(env.SecretsLocal.FileName) {
		t.Errorf("Expected %s to be created", env.SecretsLocal.FileName)
	}
	if !fileExists(env.SecretsProduction.FileName) {
		t.Errorf("Expected %s to be created", env.SecretsProduction.FileName)
	}

	// Check result
	if len(result.GeneratedFiles) != 2 {
		t.Errorf("Expected 2 generated files, got %d", len(result.GeneratedFiles))
	}
}

// Test SyncRegistryWorkflow with deployment configs
func TestSyncRegistryWorkflow_WithDeploymentConfigs(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create test file with markers
	testFile := "test-config.txt"
	initialContent := `Header line
# === AUTO-GENERATED ===
old content
# === END ===
Footer line`
	os.WriteFile(testFile, []byte(initialContent), 0600)

	// Create test registry
	registry := env.NewRegistry([]env.EnvVar{
		{Name: "TEST_VAR", Description: "Test", Default: "value"},
	})

	// Run workflow with deployment config
	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry: registry,
		AppName:  "Test App",
		DeploymentConfigs: []DeploymentConfig{
			{
				FilePath:    testFile,
				StartMarker: "# === AUTO-GENERATED ===",
				EndMarker:   "# === END ===",
				Generator: func(r *env.Registry) (string, error) {
					return "new generated content", nil
				},
			},
		},
		CreateSecretsFiles: false,
	})

	if err != nil {
		t.Fatalf("SyncRegistryWorkflow failed: %v", err)
	}

	// Read updated file
	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check content was updated
	content := string(updated)
	if !contains(content, "new generated content") {
		t.Error("Expected new generated content in file")
	}
	if contains(content, "old content") {
		t.Error("Old content should have been replaced")
	}

	// Check result
	if len(result.UpdatedFiles) < 1 {
		t.Errorf("Expected at least 1 updated file in deployment configs")
	}
}

// Test SyncRegistryWorkflow with generator error
func TestSyncRegistryWorkflow_GeneratorError(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create test file
	testFile := "test-config.txt"
	os.WriteFile(testFile, []byte("test"), 0600)

	registry := env.NewRegistry([]env.EnvVar{
		{Name: "TEST", Description: "Test"},
	})

	// Run workflow with failing generator
	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry: registry,
		AppName:  "Test App",
		DeploymentConfigs: []DeploymentConfig{
			{
				FilePath:    testFile,
				StartMarker: "# === START ===",
				EndMarker:   "# === END ===",
				Generator: func(r *env.Registry) (string, error) {
					return "", fmt.Errorf("generator error")
				},
			},
		},
		CreateSecretsFiles: false,
	})

	// Should not fail completely, but add warning
	if err != nil {
		t.Fatalf("SyncRegistryWorkflow should not fail on generator error: %v", err)
	}

	// Check warning was added
	if len(result.Warnings) == 0 {
		t.Error("Expected warning for generator error")
	}
}

// Test SyncRegistryWorkflow with nil registry
func TestSyncRegistryWorkflow_NilRegistry(t *testing.T) {
	_, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry: nil,
		AppName:  "Test App",
	})

	if err == nil {
		t.Error("Expected error for nil registry")
	}
	if !contains(err.Error(), "registry cannot be nil") {
		t.Errorf("Expected nil registry error, got: %v", err)
	}
}

// Test SyncRegistryWorkflow with empty app name (should use default)
func TestSyncRegistryWorkflow_EmptyAppName(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	registry := env.NewRegistry([]env.EnvVar{
		{Name: "TEST", Description: "Test"},
	})

	// Run with empty app name
	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry: registry,
		AppName:  "", // Empty - should use default
	})

	if err != nil {
		t.Fatalf("SyncRegistryWorkflow failed: %v", err)
	}

	// Should still create files
	if len(result.UpdatedFiles) != 2 {
		t.Errorf("Expected 2 updated files, got %d", len(result.UpdatedFiles))
	}
}

// Test SyncRegistryWorkflow with custom output writer
func TestSyncRegistryWorkflow_CustomOutputWriter(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	registry := env.NewRegistry([]env.EnvVar{
		{Name: "TEST", Description: "Test"},
	})

	// Use custom writer (even though current implementation discards it)
	var buf bytes.Buffer

	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry:     registry,
		AppName:      "Test App",
		OutputWriter: &buf,
	})

	if err != nil {
		t.Fatalf("SyncRegistryWorkflow failed: %v", err)
	}

	// Workflow should complete successfully
	if len(result.UpdatedFiles) != 2 {
		t.Errorf("Expected 2 updated files, got %d", len(result.UpdatedFiles))
	}
}

// Test SyncRegistryWorkflow skips existing secrets files
func TestSyncRegistryWorkflow_SkipsExistingSecrets(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Pre-create secrets files
	os.WriteFile(env.SecretsLocal.FileName, []byte("existing local"), 0600)
	os.WriteFile(env.SecretsProduction.FileName, []byte("existing prod"), 0600)

	registry := env.NewRegistry([]env.EnvVar{
		{Name: "SECRET", Description: "Secret", Secret: true},
	})

	// Run workflow
	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry:           registry,
		AppName:            "Test App",
		CreateSecretsFiles: true,
	})

	if err != nil {
		t.Fatalf("SyncRegistryWorkflow failed: %v", err)
	}

	// Check that secrets were skipped
	if len(result.SkippedFiles) != 2 {
		t.Errorf("Expected 2 skipped files, got %d", len(result.SkippedFiles))
	}
	if len(result.GeneratedFiles) != 0 {
		t.Errorf("Expected 0 generated files, got %d", len(result.GeneratedFiles))
	}

	// Verify content wasn't overwritten
	content, _ := os.ReadFile(env.SecretsLocal.FileName)
	if string(content) != "existing local" {
		t.Error("Existing secrets file was overwritten")
	}
}

// Test SyncRegistryWorkflow with SyncOnlyConfigs filter
func TestSyncRegistryWorkflow_FilterSingleConfig(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create test files to sync
	os.WriteFile("Dockerfile", []byte("# === START ===\n# === END ===\n"), 0644)
	os.WriteFile("fly.toml", []byte("# === START ===\n# === END ===\n"), 0644)

	// Create test registry
	registry := env.NewRegistry([]env.EnvVar{
		{Name: "TEST_VAR", Description: "Test variable", Default: "test", Group: "Test"},
	})

	// Run workflow with filter for only Dockerfile
	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry: registry,
		AppName:  "Test App",
		DeploymentConfigs: []DeploymentConfig{
			{
				FilePath:    "Dockerfile",
				StartMarker: "# === START ===",
				EndMarker:   "# === END ===",
				Generator:   func(r *env.Registry) (string, error) { return "dockerfile content", nil },
			},
			{
				FilePath:    "fly.toml",
				StartMarker: "# === START ===",
				EndMarker:   "# === END ===",
				Generator:   func(r *env.Registry) (string, error) { return "fly content", nil },
			},
		},
		SyncOnlyConfigs:    []string{"Dockerfile"}, // Only sync Dockerfile
		CreateSecretsFiles: false,
	})

	if err != nil {
		t.Fatalf("SyncRegistryWorkflow failed: %v", err)
	}

	// Check that only Dockerfile was updated
	dockerContent := readFile("Dockerfile")
	if !contains(dockerContent, "dockerfile content") {
		t.Error("Dockerfile should have been synced")
	}

	// fly.toml should NOT have been synced
	flyContent := readFile("fly.toml")
	if contains(flyContent, "fly content") {
		t.Error("fly.toml should NOT have been synced (filtered out)")
	}

	// Check result - should only show Dockerfile as updated
	if len(result.UpdatedFiles) < 1 || !contains(result.UpdatedFiles[0], "Dockerfile") {
		t.Error("Expected Dockerfile in updated files")
	}
}

// Test SyncRegistryWorkflow with SkipEnvironments flag
func TestSyncRegistryWorkflow_SkipEnvironments(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create test registry
	registry := env.NewRegistry([]env.EnvVar{
		{Name: "TEST_VAR", Description: "Test variable", Default: "test", Group: "Test"},
	})

	// Run workflow with SkipEnvironments
	result, err := SyncRegistryWorkflow(RegistrySyncOptions{
		Registry:           registry,
		AppName:            "Test App",
		SkipEnvironments:   true, // Skip .env file generation
		CreateSecretsFiles: false,
	})

	if err != nil {
		t.Fatalf("SyncRegistryWorkflow failed: %v", err)
	}

	// Check that env files were NOT created
	if fileExists(env.Local.FileName) {
		t.Error(".env.local should NOT have been created (SkipEnvironments = true)")
	}
	if fileExists(env.Production.FileName) {
		t.Error(".env.production should NOT have been created (SkipEnvironments = true)")
	}

	// Result should not include env files
	for _, file := range result.UpdatedFiles {
		if file == env.Local.FileName || file == env.Production.FileName {
			t.Errorf("Environment file %s should not be in updated files", file)
		}
	}
}

// Helper functions

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

func readFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}
