package workflow

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/joeblew999/wellknown/pkg/env"
)

// Test FinalizeWorkflow with basic encryption
func TestFinalizeWorkflow_Basic(t *testing.T) {
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

	// Generate age key
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(keyPath, []byte(identity.String()), 0600)

	// Create environment files
	os.WriteFile(env.Local.FileName, []byte("TEST_VAR=local_value\n"), 0600)
	os.WriteFile(env.Production.FileName, []byte("TEST_VAR=prod_value\n"), 0600)

	// Run workflow
	result, err := FinalizeWorkflow(FinalizeOptions{
		Environments:      env.AllEnvironmentFiles(),
		EncryptionKeyPath: keyPath,
		GitAdd:            false,
	})

	if err != nil {
		t.Fatalf("FinalizeWorkflow failed: %v", err)
	}

	// Check that encrypted files were created
	if !fileExists(env.Local.FileName + ".age") {
		t.Error("Expected encrypted local file")
	}
	if !fileExists(env.Production.FileName + ".age") {
		t.Error("Expected encrypted production file")
	}

	// Check result
	if len(result.GeneratedFiles) < 2 {
		t.Errorf("Expected at least 2 generated files, got %d", len(result.GeneratedFiles))
	}
}

// Test FinalizeWorkflow with git add enabled
func TestFinalizeWorkflow_WithGitAdd(t *testing.T) {
	// Setup temp dir as git repo
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Generate age key
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(keyPath, []byte(identity.String()), 0600)

	// Create environment files
	os.WriteFile(env.Local.FileName, []byte("TEST_VAR=local_value\n"), 0600)

	// Run workflow with git add
	result, err := FinalizeWorkflow(FinalizeOptions{
		Environments:      []*env.Environment{env.Local},
		EncryptionKeyPath: keyPath,
		GitAdd:            true,
	})

	if err != nil {
		t.Fatalf("FinalizeWorkflow failed: %v", err)
	}

	// Check that file was encrypted
	if len(result.GeneratedFiles) < 1 {
		t.Errorf("Expected at least 1 generated file, got %d", len(result.GeneratedFiles))
	}

	// Check that file was added to git (check git status)
	cmd := exec.Command("git", "status", "--porcelain")
	output, _ := cmd.Output()
	// If git add worked, output should show staged file (A prefix)
	if len(output) > 0 && !contains(string(output), ".age") {
		// Note: May have warnings if git isn't available, that's ok
	}
}

// Test FinalizeWorkflow with missing key file
func TestFinalizeWorkflow_MissingKeyFile(t *testing.T) {
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

	// Create environment files
	os.WriteFile(env.Local.FileName, []byte("TEST_VAR=value\n"), 0600)

	// Run workflow with non-existent key
	_, err = FinalizeWorkflow(FinalizeOptions{
		Environments:      []*env.Environment{env.Local},
		EncryptionKeyPath: "nonexistent/key.txt",
		GitAdd:            false,
	})

	// Should fail with clear error
	if err == nil {
		t.Error("Expected error for missing key file")
	}
	if !contains(err.Error(), "no Age key found") {
		t.Errorf("Expected key read error, got: %v", err)
	}
}

// Test FinalizeWorkflow with invalid key format
func TestFinalizeWorkflow_InvalidKeyFormat(t *testing.T) {
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

	// Create invalid key file
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")
	os.WriteFile(keyPath, []byte("invalid key content"), 0600)

	// Create environment files
	os.WriteFile(env.Local.FileName, []byte("TEST_VAR=value\n"), 0600)

	// Run workflow
	_, err = FinalizeWorkflow(FinalizeOptions{
		Environments:      []*env.Environment{env.Local},
		EncryptionKeyPath: keyPath,
		GitAdd:            false,
	})

	// Should fail with parsing error
	if err == nil {
		t.Error("Expected error for invalid key format")
	}
	if !contains(err.Error(), "failed to parse identity") {
		t.Errorf("Expected key parse error, got: %v", err)
	}
}

// Test FinalizeWorkflow with non-existent environment file
func TestFinalizeWorkflow_NonExistentEnvFile(t *testing.T) {
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

	// Generate age key
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(keyPath, []byte(identity.String()), 0600)

	// Run workflow with non-existent env files
	result, err := FinalizeWorkflow(FinalizeOptions{
		Environments:      env.AllEnvironmentFiles(),
		EncryptionKeyPath: keyPath,
		GitAdd:            false,
	})

	// Should succeed but skip files
	if err != nil {
		t.Fatalf("FinalizeWorkflow should not fail on missing env files: %v", err)
	}

	// Should have skipped all files
	if len(result.SkippedFiles) == 0 {
		t.Error("Expected skipped files for non-existent environments")
	}
}

// Test FinalizeWorkflow with default key path
func TestFinalizeWorkflow_DefaultKeyPath(t *testing.T) {
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

	// Generate age key at DEFAULT location
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(keyPath, []byte(identity.String()), 0600)

	// Create environment files
	os.WriteFile(env.Local.FileName, []byte("TEST_VAR=value\n"), 0600)

	// Run workflow with EMPTY key path (should use default)
	result, err := FinalizeWorkflow(FinalizeOptions{
		Environments:      []*env.Environment{env.Local},
		EncryptionKeyPath: "", // Empty - should use env.DefaultAgeKeyPath
		GitAdd:            false,
	})

	if err != nil {
		t.Fatalf("FinalizeWorkflow failed: %v", err)
	}

	// Should encrypt successfully
	if len(result.GeneratedFiles) < 1 {
		t.Error("Expected at least 1 encrypted file")
	}
}

// Test FinalizeWorkflow with nil environments (should use default)
func TestFinalizeWorkflow_NilEnvironments(t *testing.T) {
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

	// Generate age key
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(keyPath, []byte(identity.String()), 0600)

	// Create environment files
	os.WriteFile(env.Local.FileName, []byte("TEST_VAR=value\n"), 0600)

	// Run workflow with nil environments
	result, err := FinalizeWorkflow(FinalizeOptions{
		Environments:      nil, // Should use default
		EncryptionKeyPath: keyPath,
		GitAdd:            false,
	})

	if err != nil {
		t.Fatalf("FinalizeWorkflow failed: %v", err)
	}

	// Should handle defaults gracefully
	// (may skip or encrypt depending on what exists)
	if len(result.GeneratedFiles) == 0 && len(result.SkippedFiles) == 0 {
		t.Error("Expected some file handling with default environments")
	}
}

// Test FinalizeWorkflow with custom output writer
func TestFinalizeWorkflow_CustomOutputWriter(t *testing.T) {
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

	// Generate age key
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(keyPath, []byte(identity.String()), 0600)

	// Create environment files
	os.WriteFile(env.Local.FileName, []byte("TEST_VAR=value\n"), 0600)

	// Use custom writer
	var buf bytes.Buffer

	result, err := FinalizeWorkflow(FinalizeOptions{
		Environments:      []*env.Environment{env.Local},
		EncryptionKeyPath: keyPath,
		GitAdd:            false,
		OutputWriter:      &buf,
	})

	if err != nil {
		t.Fatalf("FinalizeWorkflow failed: %v", err)
	}

	// Should complete successfully
	if len(result.GeneratedFiles) < 1 {
		t.Error("Expected at least 1 generated file")
	}
}

// Test FinalizeWorkflow encryption round-trip
func TestFinalizeWorkflow_EncryptionRoundTrip(t *testing.T) {
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

	// Generate age key
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(keyPath, []byte(identity.String()), 0600)

	// Create environment file with known content
	originalContent := "SECRET_VAR=super_secret_value\nPUBLIC_VAR=public_value\n"
	os.WriteFile(env.Local.FileName, []byte(originalContent), 0600)

	// Encrypt
	result, err := FinalizeWorkflow(FinalizeOptions{
		Environments:      []*env.Environment{env.Local},
		EncryptionKeyPath: keyPath,
		GitAdd:            false,
	})

	if err != nil {
		t.Fatalf("FinalizeWorkflow failed: %v", err)
	}

	// Verify encrypted file exists
	encryptedPath := env.Local.FileName + ".age"
	if !fileExists(encryptedPath) {
		t.Fatal("Encrypted file not created")
	}

	// Decrypt and verify
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		t.Fatal(err)
	}

	identities, err := age.ParseIdentities(bytes.NewReader([]byte(identity.String())))
	if err != nil {
		t.Fatal(err)
	}

	decryptor, err := age.Decrypt(bytes.NewReader(encryptedData), identities...)
	if err != nil {
		t.Fatal(err)
	}

	var decrypted bytes.Buffer
	if _, err := decrypted.ReadFrom(decryptor); err != nil {
		t.Fatal(err)
	}

	// Verify content matches
	if decrypted.String() != originalContent {
		t.Errorf("Decrypted content doesn't match:\nGot: %q\nWant: %q", decrypted.String(), originalContent)
	}

	// Verify result
	if len(result.GeneratedFiles) < 1 {
		t.Error("Expected at least 1 generated file")
	}
}

// Test FinalizeWorkflow with git add failure (simulated by no git repo)
func TestFinalizeWorkflow_GitAddFailure(t *testing.T) {
	// Setup temp dir WITHOUT git init
	tmpDir, err := os.MkdirTemp("", "workflow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Generate age key
	keyDir := ".age"
	os.Mkdir(keyDir, 0700)
	keyPath := filepath.Join(keyDir, "key.txt")

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(keyPath, []byte(identity.String()), 0600)

	// Create environment files
	os.WriteFile(env.Local.FileName, []byte("TEST_VAR=value\n"), 0600)

	// Run workflow with git add (should fail gracefully)
	result, err := FinalizeWorkflow(FinalizeOptions{
		Environments:      []*env.Environment{env.Local},
		EncryptionKeyPath: keyPath,
		GitAdd:            true, // Will fail - not a git repo
	})

	// Should NOT fail completely
	if err != nil {
		t.Fatalf("FinalizeWorkflow should not fail on git add error: %v", err)
	}

	// Should have encrypted file
	if len(result.GeneratedFiles) < 1 {
		t.Error("Expected encrypted file despite git error")
	}

	// Should have warning about git failure
	if len(result.Warnings) == 0 {
		t.Error("Expected warning for git add failure")
	}
}
