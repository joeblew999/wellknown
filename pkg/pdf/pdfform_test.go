package pdfform_test

import (
	"os"
	"path/filepath"
	"testing"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

func TestLoadTestCase(t *testing.T) {
	// Create a temporary test case
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.json")

	testCase := &pdfform.TestCase{
		Name:        "test_case",
		Description: "Test description",
		PdfURL:      "http://example.com/form.pdf",
		Fields: map[string]string{
			"field1": "value1",
			"field2": "value2",
		},
		ExpectError: false,
	}

	// Save test case
	if err := pdfform.SaveTestCase(testCase, testFile); err != nil {
		t.Fatalf("Failed to save test case: %v", err)
	}

	// Load it back
	loaded, err := pdfform.LoadTestCase(testFile)
	if err != nil {
		t.Fatalf("Failed to load test case: %v", err)
	}

	if loaded.Name != testCase.Name {
		t.Errorf("Expected name %s, got %s", testCase.Name, loaded.Name)
	}

	if len(loaded.Fields) != len(testCase.Fields) {
		t.Errorf("Expected %d fields, got %d", len(testCase.Fields), len(loaded.Fields))
	}
}

func TestListFormFields(t *testing.T) {
	// This test requires an actual PDF file
	// Skip if no test PDF available
	testPDF := "testdata/test_form.pdf"
	if _, err := os.Stat(testPDF); os.IsNotExist(err) {
		t.Skip("Skipping test: no test PDF available")
	}

	fields, err := pdfform.ListFormFields(testPDF)
	if err != nil {
		t.Fatalf("Failed to list form fields: %v", err)
	}

	if len(fields) == 0 {
		t.Error("Expected at least one form field")
	}
}

func TestExportFormFieldsToJSON(t *testing.T) {
	testPDF := "testdata/test_form.pdf"
	if _, err := os.Stat(testPDF); os.IsNotExist(err) {
		t.Skip("Skipping test: no test PDF available")
	}

	tempDir := t.TempDir()
	outputJSON := filepath.Join(tempDir, "fields.json")

	if err := pdfform.ExportFormFieldsToJSON(testPDF, outputJSON); err != nil {
		t.Fatalf("Failed to export form fields: %v", err)
	}

	// Verify JSON file was created
	if _, err := os.Stat(outputJSON); os.IsNotExist(err) {
		t.Error("Expected JSON file to be created")
	}

	// Load the exported template
	formData, err := pdfform.LoadTestCase(outputJSON)
	if err != nil {
		t.Fatalf("Failed to load exported template: %v", err)
	}

	if len(formData.Fields) == 0 {
		t.Error("Expected at least one field in exported template")
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/form.pdf", true},
		{"http://example.com/form.pdf", true},
		{"/path/to/local/file.pdf", false},
		{"./relative/path.pdf", false},
		{"file.pdf", false},
		{"", false},
	}

	// Note: isURL is not exported, so we test it indirectly through FormData behavior
	// This is a placeholder for documentation
	_ = tests
}
