package pdfform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateCase(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create a case
	c, casePath, err := CreateCase("F3520", "Test Vehicle Sale", "test_user", tempDir)
	if err != nil {
		t.Fatalf("CreateCase failed: %v", err)
	}

	// Verify case was created
	if c == nil {
		t.Fatal("Case is nil")
	}

	if c.Metadata.CaseName != "Test Vehicle Sale" {
		t.Errorf("Expected case name 'Test Vehicle Sale', got '%s'", c.Metadata.CaseName)
	}

	if c.FormReference.FormCode != "F3520" {
		t.Errorf("Expected form code 'F3520', got '%s'", c.FormReference.FormCode)
	}

	// Verify file was created
	if _, err := os.Stat(casePath); os.IsNotExist(err) {
		t.Errorf("Case file was not created at %s", casePath)
	}
}

func TestLoadCase(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create a case
	original, casePath, err := CreateCase("F3520", "Test Vehicle Sale", "test_user", tempDir)
	if err != nil {
		t.Fatalf("CreateCase failed: %v", err)
	}

	// Load the case
	loaded, err := LoadCase(casePath)
	if err != nil {
		t.Fatalf("LoadCase failed: %v", err)
	}

	// Verify loaded case matches original
	if loaded.Metadata.CaseID != original.Metadata.CaseID {
		t.Errorf("Expected case ID '%s', got '%s'", original.Metadata.CaseID, loaded.Metadata.CaseID)
	}

	if loaded.Metadata.CaseName != original.Metadata.CaseName {
		t.Errorf("Expected case name '%s', got '%s'", original.Metadata.CaseName, loaded.Metadata.CaseName)
	}

	if loaded.FormReference.FormCode != original.FormReference.FormCode {
		t.Errorf("Expected form code '%s', got '%s'", original.FormReference.FormCode, loaded.FormReference.FormCode)
	}
}

func TestSaveCase(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create a case
	c, casePath, err := CreateCase("F3520", "Test Vehicle Sale", "test_user", tempDir)
	if err != nil {
		t.Fatalf("CreateCase failed: %v", err)
	}

	// Modify the case
	c.Fields["Text1"] = "John"
	c.Fields["Text2"] = "Smith"

	// Save the case
	if err := SaveCase(c, casePath); err != nil {
		t.Fatalf("SaveCase failed: %v", err)
	}

	// Load the case again
	loaded, err := LoadCase(casePath)
	if err != nil {
		t.Fatalf("LoadCase failed: %v", err)
	}

	// Verify fields were saved
	if loaded.Fields["Text1"] != "John" {
		t.Errorf("Expected field Text1 to be 'John', got '%s'", loaded.Fields["Text1"])
	}

	if loaded.Fields["Text2"] != "Smith" {
		t.Errorf("Expected field Text2 to be 'Smith', got '%s'", loaded.Fields["Text2"])
	}
}

func TestListCases(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create multiple cases
	c1, p1, err := CreateCase("F3520", "Test Sale 1", "user1", tempDir)
	if err != nil {
		t.Fatalf("CreateCase 1 failed: %v", err)
	}
	t.Logf("Created case 1: %s at %s", c1.Metadata.CaseID, p1)

	c2, p2, err := CreateCase("F3520", "Test Sale 2", "user1", tempDir)
	if err != nil {
		t.Fatalf("CreateCase 2 failed: %v", err)
	}
	t.Logf("Created case 2: %s at %s", c2.Metadata.CaseID, p2)

	c3, p3, err := CreateCase("F4101", "Test Transfer", "user2", tempDir)
	if err != nil {
		t.Fatalf("CreateCase 3 failed: %v", err)
	}
	t.Logf("Created case 3: %s at %s", c3.Metadata.CaseID, p3)

	// List all cases
	allCases, err := ListCases(tempDir, "")
	if err != nil {
		t.Fatalf("ListCases failed: %v", err)
	}
	t.Logf("Found %d total cases: %v", len(allCases), allCases)

	if len(allCases) != 3 {
		t.Errorf("Expected 3 cases, got %d", len(allCases))
	}

	// List cases for user1
	user1Cases, err := ListCases(tempDir, "user1")
	if err != nil {
		t.Fatalf("ListCases for user1 failed: %v", err)
	}
	t.Logf("Found %d user1 cases: %v", len(user1Cases), user1Cases)

	if len(user1Cases) != 2 {
		t.Errorf("Expected 2 cases for user1, got %d", len(user1Cases))
	}

	// List cases for user2
	user2Cases, err := ListCases(tempDir, "user2")
	if err != nil {
		t.Fatalf("ListCases for user2 failed: %v", err)
	}
	t.Logf("Found %d user2 cases: %v", len(user2Cases), user2Cases)

	if len(user2Cases) != 1 {
		t.Errorf("Expected 1 case for user2, got %d", len(user2Cases))
	}
}

func TestListCases_EmptyDirectory(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// List cases in empty directory
	cases, err := ListCases(tempDir, "")
	if err != nil {
		t.Fatalf("ListCases failed: %v", err)
	}

	if len(cases) != 0 {
		t.Errorf("Expected 0 cases in empty directory, got %d", len(cases))
	}
}

func TestFillFromCase_MissingCase(t *testing.T) {
	// Try to fill from non-existent case
	_, err := FillFromCase("/nonexistent/case.json", t.TempDir(), false)
	if err == nil {
		t.Error("Expected error for non-existent case, got nil")
	}
}

func TestValidateCase(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create a case
	c, _, err := CreateCase("F3520", "Test Vehicle Sale", "test_user", tempDir)
	if err != nil {
		t.Fatalf("CreateCase failed: %v", err)
	}

	// Validate without template
	catalogPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")
	err = ValidateCase(c, catalogPath)
	if err != nil {
		t.Fatalf("ValidateCase failed: %v", err)
	}

	// Validation should mark as valid (no template to compare against)
	if c.Validation == nil {
		t.Error("Expected validation to be set")
	}

	if !c.Validation.Valid {
		t.Error("Expected case to be valid")
	}
}
