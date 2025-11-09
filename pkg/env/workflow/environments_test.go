package workflow

import (
	"bytes"
	"os"
	"testing"

	"github.com/joeblew999/wellknown/pkg/env"
)

// Test SyncEnvironmentsWorkflow with basic setup
func TestSyncEnvironmentsWorkflow_Basic(t *testing.T) {
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
		{Name: "PUBLIC_VAR", Description: "Public", Default: "public_value", Secret: false},
		{Name: "SECRET_VAR", Description: "Secret", Default: "secret_value", Secret: true, Required: true},
	})

	// Create environment templates
	localTemplate := env.Local.Generate(registry, "Test App")
	os.WriteFile(env.Local.FileName, []byte(localTemplate), 0600)

	prodTemplate := env.Production.Generate(registry, "Test App")
	os.WriteFile(env.Production.FileName, []byte(prodTemplate), 0600)

	// Create secrets files
	secretsContent := "SECRET_VAR=my_secret_value\n"
	os.WriteFile(env.SecretsLocal.FileName, []byte(secretsContent), 0600)
	os.WriteFile(env.SecretsProduction.FileName, []byte(secretsContent), 0600)

	// Run workflow
	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:          registry,
		AppName:           "Test App",
		LocalEnv:          env.Local,
		ProductionEnv:     env.Production,
		LocalSecrets:      env.SecretsLocal,
		ProductionSecrets: env.SecretsProduction,
		ValidateRequired:  true,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Check that files were updated
	if len(result.UpdatedFiles) != 2 {
		t.Errorf("Expected 2 updated files, got %d", len(result.UpdatedFiles))
	}

	// Check that merged env files contain the secret
	localContent := readFile(env.Local.FileName)
	if !contains(localContent, "SECRET_VAR=my_secret_value") {
		t.Error("Local env should contain merged secret")
	}

	prodContent := readFile(env.Production.FileName)
	if !contains(prodContent, "SECRET_VAR=my_secret_value") {
		t.Error("Production env should contain merged secret")
	}
}

// Test SyncEnvironmentsWorkflow with plaintext secrets (simpler test)
func TestSyncEnvironmentsWorkflow_PlaintextSecrets(t *testing.T) {
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
		{Name: "SECRET_VAR", Description: "Secret", Default: "secret_value", Secret: true},
	})

	// Create environment templates
	localTemplate := env.Local.Generate(registry, "Test App")
	os.WriteFile(env.Local.FileName, []byte(localTemplate), 0600)

	prodTemplate := env.Production.Generate(registry, "Test App")
	os.WriteFile(env.Production.FileName, []byte(prodTemplate), 0600)

	// Create plaintext secrets files
	secretsContent := "SECRET_VAR=plaintext_secret\n"
	os.WriteFile(env.SecretsLocal.FileName, []byte(secretsContent), 0600)
	os.WriteFile(env.SecretsProduction.FileName, []byte(secretsContent), 0600)

	// Run workflow
	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:          registry,
		AppName:           "Test App",
		LocalEnv:          env.Local,
		ProductionEnv:     env.Production,
		ValidateRequired:  false,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Should succeed
	if len(result.UpdatedFiles) != 2 {
		t.Errorf("Expected 2 updated files, got %d", len(result.UpdatedFiles))
	}
}

// Test SyncEnvironmentsWorkflow with missing secrets file
func TestSyncEnvironmentsWorkflow_MissingSecrets(t *testing.T) {
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
		{Name: "TEST_VAR", Description: "Test", Default: "value"},
	})

	// Create environment templates but NO secrets files
	localTemplate := env.Local.Generate(registry, "Test App")
	os.WriteFile(env.Local.FileName, []byte(localTemplate), 0600)

	// Run workflow
	_, err = SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:      registry,
		AppName:       "Test App",
		LocalEnv:      env.Local,
		ProductionEnv: env.Production,
	})

	// Should fail with clear error
	if err == nil {
		t.Error("Expected error for missing secrets file")
	}
	if !contains(err.Error(), "failed to load secrets") {
		t.Errorf("Expected secrets loading error, got: %v", err)
	}
}

// Test SyncEnvironmentsWorkflow with validation enabled
func TestSyncEnvironmentsWorkflow_ValidationEnabled(t *testing.T) {
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

	// Create test registry with REQUIRED variable
	registry := env.NewRegistry([]env.EnvVar{
		{Name: "REQUIRED_VAR", Description: "Required", Default: "", Secret: true, Required: true},
	})

	// Create environment templates
	localTemplate := env.Local.Generate(registry, "Test App")
	os.WriteFile(env.Local.FileName, []byte(localTemplate), 0600)

	prodTemplate := env.Production.Generate(registry, "Test App")
	os.WriteFile(env.Production.FileName, []byte(prodTemplate), 0600)

	// Create secrets files with EMPTY value for required var
	secretsContent := "REQUIRED_VAR=\n"
	os.WriteFile(env.SecretsLocal.FileName, []byte(secretsContent), 0600)
	os.WriteFile(env.SecretsProduction.FileName, []byte(secretsContent), 0600)

	// Run workflow with validation
	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:         registry,
		AppName:          "Test App",
		LocalEnv:         env.Local,
		ProductionEnv:    env.Production,
		ValidateRequired: true,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Should have warning about validation failure
	if len(result.Warnings) == 0 {
		t.Error("Expected validation warning for empty required variable")
	}
}

// Test SyncEnvironmentsWorkflow with validation disabled
func TestSyncEnvironmentsWorkflow_ValidationDisabled(t *testing.T) {
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

	// Create test registry with required variable
	registry := env.NewRegistry([]env.EnvVar{
		{Name: "REQUIRED_VAR", Description: "Required", Default: "", Secret: true, Required: true},
	})

	// Create environment templates
	localTemplate := env.Local.Generate(registry, "Test App")
	os.WriteFile(env.Local.FileName, []byte(localTemplate), 0600)

	prodTemplate := env.Production.Generate(registry, "Test App")
	os.WriteFile(env.Production.FileName, []byte(prodTemplate), 0600)

	// Create secrets files with empty value
	secretsContent := "REQUIRED_VAR=\n"
	os.WriteFile(env.SecretsLocal.FileName, []byte(secretsContent), 0600)
	os.WriteFile(env.SecretsProduction.FileName, []byte(secretsContent), 0600)

	// Run workflow WITHOUT validation
	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:         registry,
		AppName:          "Test App",
		LocalEnv:         env.Local,
		ProductionEnv:    env.Production,
		ValidateRequired: false,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Should have NO warnings
	if len(result.Warnings) > 0 {
		t.Error("Should not have warnings when validation disabled")
	}
}

// Test SyncEnvironmentsWorkflow with nil registry
func TestSyncEnvironmentsWorkflow_NilRegistry(t *testing.T) {
	_, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
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

// Test SyncEnvironmentsWorkflow with empty app name (should use default)
func TestSyncEnvironmentsWorkflow_EmptyAppName(t *testing.T) {
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
		{Name: "TEST", Description: "Test", Secret: true},
	})

	// Create templates and secrets
	localTemplate := env.Local.Generate(registry, "Test App")
	os.WriteFile(env.Local.FileName, []byte(localTemplate), 0600)

	prodTemplate := env.Production.Generate(registry, "Test App")
	os.WriteFile(env.Production.FileName, []byte(prodTemplate), 0600)

	secretsContent := "TEST=value\n"
	os.WriteFile(env.SecretsLocal.FileName, []byte(secretsContent), 0600)
	os.WriteFile(env.SecretsProduction.FileName, []byte(secretsContent), 0600)

	// Run with empty app name
	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:      registry,
		AppName:       "", // Empty - should use default
		LocalEnv:      env.Local,
		ProductionEnv: env.Production,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Should still work
	if len(result.UpdatedFiles) != 2 {
		t.Errorf("Expected 2 updated files, got %d", len(result.UpdatedFiles))
	}
}

// Test SyncEnvironmentsWorkflow with custom output writer
func TestSyncEnvironmentsWorkflow_CustomOutputWriter(t *testing.T) {
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
		{Name: "TEST", Description: "Test", Secret: true},
	})

	// Create templates and secrets
	localTemplate := env.Local.Generate(registry, "Test App")
	os.WriteFile(env.Local.FileName, []byte(localTemplate), 0600)

	prodTemplate := env.Production.Generate(registry, "Test App")
	os.WriteFile(env.Production.FileName, []byte(prodTemplate), 0600)

	secretsContent := "TEST=value\n"
	os.WriteFile(env.SecretsLocal.FileName, []byte(secretsContent), 0600)
	os.WriteFile(env.SecretsProduction.FileName, []byte(secretsContent), 0600)

	// Use custom writer
	var buf bytes.Buffer

	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:      registry,
		AppName:       "Test App",
		LocalEnv:      env.Local,
		ProductionEnv: env.Production,
		OutputWriter:  &buf,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Should complete successfully
	if len(result.UpdatedFiles) != 2 {
		t.Errorf("Expected 2 updated files, got %d", len(result.UpdatedFiles))
	}
}

// Test SyncEnvironmentsWorkflow with both secrets files present
func TestSyncEnvironmentsWorkflow_BothSecretsPresent(t *testing.T) {
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
		{Name: "SECRET", Description: "Secret", Secret: true},
	})

	// Create templates
	localTemplate := env.Local.Generate(registry, "Test App")
	os.WriteFile(env.Local.FileName, []byte(localTemplate), 0600)

	prodTemplate := env.Production.Generate(registry, "Test App")
	os.WriteFile(env.Production.FileName, []byte(prodTemplate), 0600)

	// Create BOTH secrets files with different values
	os.WriteFile(env.SecretsLocal.FileName, []byte("SECRET=local_value\n"), 0600)
	os.WriteFile(env.SecretsProduction.FileName, []byte("SECRET=prod_value\n"), 0600)

	// Run workflow
	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:      registry,
		AppName:       "Test App",
		LocalEnv:      env.Local,
		ProductionEnv: env.Production,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Should succeed with no fallback warnings
	if len(result.UpdatedFiles) != 2 {
		t.Errorf("Expected 2 updated files, got %d", len(result.UpdatedFiles))
	}

	// Verify different values were used
	localContent := readFile(env.Local.FileName)
	prodContent := readFile(env.Production.FileName)

	if !contains(localContent, "SECRET=local_value") {
		t.Error("Local env should contain local secret value")
	}
	if !contains(prodContent, "SECRET=prod_value") {
		t.Error("Production env should contain production secret value")
	}
}

// Test SyncEnvironmentsWorkflow with only local environment
func TestSyncEnvironmentsWorkflow_LocalOnly(t *testing.T) {
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
		{Name: "SECRET", Description: "Secret variable", Secret: true, Group: "Secrets"},
	})

	// Create local secrets file
	os.WriteFile(env.SecretsLocal.FileName, []byte("SECRET=local_secret_value\n"), 0600)

	// Run workflow with ONLY LocalEnv (ProductionEnv = nil)
	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:         registry,
		AppName:          "Test App",
		LocalEnv:         env.Local,      // Sync local
		ProductionEnv:    nil,             // Skip production
		ValidateRequired: false,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Check that ONLY local was updated
	if !fileExists(env.Local.FileName) {
		t.Error("Local environment should have been created")
	}
	if fileExists(env.Production.FileName) {
		t.Error("Production environment should NOT have been created")
	}

	// Result should only contain local
	if len(result.UpdatedFiles) != 1 {
		t.Errorf("Expected 1 updated file (local only), got %d", len(result.UpdatedFiles))
	}
	if result.UpdatedFiles[0] != env.Local.FileName {
		t.Errorf("Expected local env in updated files, got %s", result.UpdatedFiles[0])
	}

	// Verify local content
	localContent := readFile(env.Local.FileName)
	if !contains(localContent, "SECRET=local_secret_value") {
		t.Error("Local env should contain secret value")
	}
}

// Test SyncEnvironmentsWorkflow with only production environment
func TestSyncEnvironmentsWorkflow_ProductionOnly(t *testing.T) {
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
		{Name: "SECRET", Description: "Secret variable", Secret: true, Group: "Secrets"},
	})

	// Create production secrets file
	os.WriteFile(env.SecretsProduction.FileName, []byte("SECRET=prod_secret_value\n"), 0600)

	// Run workflow with ONLY ProductionEnv (LocalEnv = nil)
	result, err := SyncEnvironmentsWorkflow(EnvironmentsSyncOptions{
		Registry:         registry,
		AppName:          "Test App",
		LocalEnv:         nil,               // Skip local
		ProductionEnv:    env.Production,    // Sync production
		ValidateRequired: false,
	})

	if err != nil {
		t.Fatalf("SyncEnvironmentsWorkflow failed: %v", err)
	}

	// Check that ONLY production was updated
	if fileExists(env.Local.FileName) {
		t.Error("Local environment should NOT have been created")
	}
	if !fileExists(env.Production.FileName) {
		t.Error("Production environment should have been created")
	}

	// Result should only contain production
	if len(result.UpdatedFiles) != 1 {
		t.Errorf("Expected 1 updated file (production only), got %d", len(result.UpdatedFiles))
	}
	if result.UpdatedFiles[0] != env.Production.FileName {
		t.Errorf("Expected production env in updated files, got %s", result.UpdatedFiles[0])
	}

	// Verify production content
	prodContent := readFile(env.Production.FileName)
	if !contains(prodContent, "SECRET=prod_secret_value") {
		t.Error("Production env should contain secret value")
	}
}
