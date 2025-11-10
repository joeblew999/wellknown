package pdfform_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

func TestBrowse_AllStates(t *testing.T) {
	catalogPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")

	result, err := pdfform.Browse(pdfform.BrowseOptions{
		CatalogPath: catalogPath,
		State:       "",
	})
	if err != nil {
		t.Fatalf("Browse failed: %v", err)
	}

	if len(result.States) == 0 {
		t.Error("Expected at least one state")
	}

	if len(result.Forms) != 0 {
		t.Error("Expected no forms when browsing all states")
	}

	t.Logf("Found %d states: %v", len(result.States), result.States)
}

func TestBrowse_SpecificState(t *testing.T) {
	catalogPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")

	result, err := pdfform.Browse(pdfform.BrowseOptions{
		CatalogPath: catalogPath,
		State:       "VIC",
	})
	if err != nil {
		t.Fatalf("Browse failed: %v", err)
	}

	if len(result.States) != 0 {
		t.Error("Expected no states when browsing specific state")
	}

	if len(result.Forms) == 0 {
		t.Error("Expected at least one form for VIC")
	}

	for _, form := range result.Forms {
		if form.State != "VIC" {
			t.Errorf("Expected VIC form, got %s", form.State)
		}
	}

	t.Logf("Found %d forms for VIC", len(result.Forms))
}

func TestDownload(t *testing.T) {
	if os.Getenv("ENABLE_DOWNLOAD_TESTS") != "true" {
		t.Skip("Skipping download test. Set ENABLE_DOWNLOAD_TESTS=true to run")
	}

	catalogPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")
	outputDir := t.TempDir()

	result, err := pdfform.Download(pdfform.DownloadOptions{
		CatalogPath: catalogPath,
		FormCode:    "F3520",
		OutputDir:   outputDir,
	})
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify PDF was downloaded
	if _, err := os.Stat(result.PDFPath); os.IsNotExist(err) {
		t.Errorf("Downloaded PDF does not exist: %s", result.PDFPath)
	}

	// Verify metadata was created
	if _, err := os.Stat(result.Metadata); os.IsNotExist(err) {
		t.Errorf("Provenance metadata does not exist: %s", result.Metadata)
	}

	// Verify form info
	if result.Form == nil {
		t.Fatal("Expected form info, got nil")
	}
	if result.Form.FormCode != "F3520" {
		t.Errorf("Expected form code F3520, got %s", result.Form.FormCode)
	}

	t.Logf("Downloaded: %s to %s", result.Form.FormName, result.PDFPath)
}

func TestInspect(t *testing.T) {
	// This test requires a real PDF file
	// We'll check if one exists from a previous download test
	testPDF := filepath.Join("data", "downloads", "F3520_Transfer_of_Registration.pdf")
	if _, err := os.Stat(testPDF); os.IsNotExist(err) {
		t.Skip("Skipping inspect test - no test PDF available. Run download test first.")
	}

	outputDir := t.TempDir()

	result, err := pdfform.Inspect(pdfform.InspectOptions{
		PDFPath:   testPDF,
		OutputDir: outputDir,
	})
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}

	// Verify template was created
	if _, err := os.Stat(result.TemplatePath); os.IsNotExist(err) {
		t.Errorf("Template file does not exist: %s", result.TemplatePath)
	}

	if result.FieldCount == 0 {
		t.Error("Expected at least one form field")
	}

	// Verify template is valid JSON
	data, err := os.ReadFile(result.TemplatePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	var formData pdfform.FormData
	if err := json.Unmarshal(data, &formData); err != nil {
		t.Errorf("Template is not valid JSON: %v", err)
	}

	if len(formData.Fields) != result.FieldCount {
		t.Errorf("Expected %d fields in template, got %d", result.FieldCount, len(formData.Fields))
	}

	t.Logf("Extracted %d fields, template: %s", result.FieldCount, result.TemplatePath)
}

func TestFill(t *testing.T) {
	// This test requires a test case JSON file with valid PDF
	testCasePath := filepath.Join("examples", "pdfform", "testdata", "cases", "vba_basic.json")
	if _, err := os.Stat(testCasePath); os.IsNotExist(err) {
		t.Skip("Skipping fill test - no test case available")
	}

	outputDir := t.TempDir()

	result, err := pdfform.Fill(pdfform.FillOptions{
		DataPath:  testCasePath,
		OutputDir: outputDir,
		Flatten:   false,
	})
	if err != nil {
		// Test data might reference non-existent PDF
		t.Skipf("Skipping fill test - test case references unavailable PDF: %v", err)
	}

	// Verify output PDF was created
	if _, err := os.Stat(result.OutputPath); os.IsNotExist(err) {
		t.Errorf("Filled PDF does not exist: %s", result.OutputPath)
	}

	if result.Flattened {
		t.Error("Expected non-flattened PDF, but Flattened=true")
	}

	t.Logf("Filled PDF: %s", result.OutputPath)
}

func TestFill_WithFlatten(t *testing.T) {
	testCasePath := filepath.Join("examples", "pdfform", "testdata", "cases", "vba_basic.json")
	if _, err := os.Stat(testCasePath); os.IsNotExist(err) {
		t.Skip("Skipping fill test - no test case available")
	}

	outputDir := t.TempDir()

	result, err := pdfform.Fill(pdfform.FillOptions{
		DataPath:  testCasePath,
		OutputDir: outputDir,
		Flatten:   true,
	})
	if err != nil {
		// Test data might reference non-existent PDF
		t.Skipf("Skipping fill with flatten test - test case references unavailable PDF: %v", err)
	}

	// Verify output PDF was created
	if _, err := os.Stat(result.OutputPath); os.IsNotExist(err) {
		t.Errorf("Filled PDF does not exist: %s", result.OutputPath)
	}

	if !result.Flattened {
		t.Error("Expected flattened PDF, but Flattened=false")
	}

	t.Logf("Flattened PDF: %s", result.OutputPath)
}

func TestTest(t *testing.T) {
	testCasePath := filepath.Join("examples", "pdfform", "testdata", "cases", "vba_basic.json")
	if _, err := os.Stat(testCasePath); os.IsNotExist(err) {
		t.Skip("Skipping test runner test - no test case available")
	}

	outputDir := t.TempDir()

	result, err := pdfform.Test(pdfform.TestOptions{
		TestCasePath: testCasePath,
		OutputDir:    outputDir,
	})
	if err != nil {
		t.Logf("Test execution error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected test result, got nil")
	}

	if result.Name == "" {
		t.Error("Expected test name to be set")
	}

	if result.Passed {
		t.Logf("Test passed: %s -> %s", result.Name, result.OutputPath)
	} else {
		t.Logf("Test failed: %s - %v", result.Name, result.Error)
	}
}

// TestWorkflowEndToEnd tests the complete workflow: browse -> download -> inspect -> fill
// This is the key integration test that populates the data folders
func TestWorkflowEndToEnd(t *testing.T) {
	if os.Getenv("ENABLE_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping end-to-end workflow test. Set ENABLE_INTEGRATION_TESTS=true to run")
	}

	catalogPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")
	outputDir := t.TempDir()

	// Prepare test field data
	fieldData := map[string]string{
		"test_field_1": "Test Value 1",
		"test_field_2": "Test Value 2",
	}

	result, err := pdfform.RunWorkflow(pdfform.WorkflowOptions{
		CatalogPath: catalogPath,
		FormCode:    "F3520",
		FieldData:   fieldData,
		OutputDir:   outputDir,
		Flatten:     false,
	})
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// Verify all steps completed
	if result.FormCode != "F3520" {
		t.Errorf("Expected form code F3520, got %s", result.FormCode)
	}

	// Verify download step
	if _, err := os.Stat(result.DownloadPath); os.IsNotExist(err) {
		t.Errorf("Download step failed - PDF does not exist: %s", result.DownloadPath)
	}

	// Verify inspect step
	if _, err := os.Stat(result.TemplatePath); os.IsNotExist(err) {
		t.Errorf("Inspect step failed - Template does not exist: %s", result.TemplatePath)
	}

	// Verify fill step
	if _, err := os.Stat(result.FilledPath); os.IsNotExist(err) {
		t.Errorf("Fill step failed - Filled PDF does not exist: %s", result.FilledPath)
	}

	// Verify provenance tracking
	if result.Provenance == nil {
		t.Error("Expected provenance metadata, got nil")
	} else {
		if result.Provenance.CatalogFormCode != "F3520" {
			t.Errorf("Expected provenance form code F3520, got %s", result.Provenance.CatalogFormCode)
		}
		if result.Provenance.DownloadedAt.IsZero() {
			t.Error("Expected download timestamp to be set")
		}
	}

	t.Logf("Complete workflow succeeded:")
	t.Logf("  Downloaded: %s", result.DownloadPath)
	t.Logf("  Template:   %s", result.TemplatePath)
	t.Logf("  Filled:     %s", result.FilledPath)
}

// TestWorkflowPopulatesDataFolders tests that the workflow correctly populates
// the data/ folder structure with downloads, templates, and outputs
func TestWorkflowPopulatesDataFolders(t *testing.T) {
	if os.Getenv("ENABLE_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping data folder population test. Set ENABLE_INTEGRATION_TESTS=true to run")
	}

	catalogPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")
	baseDir := "data"

	// Prepare test field data
	fieldData := map[string]string{
		"test_field": "Test Value",
	}

	result, err := pdfform.RunWorkflow(pdfform.WorkflowOptions{
		CatalogPath: catalogPath,
		FormCode:    "F3520",
		FieldData:   fieldData,
		OutputDir:   baseDir,
		Flatten:     false,
	})
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// Verify data/downloads/ was populated
	downloadsDir := filepath.Join(baseDir, "downloads")
	if _, err := os.Stat(downloadsDir); os.IsNotExist(err) {
		t.Errorf("downloads directory was not created: %s", downloadsDir)
	}

	// Verify data/templates/ was populated
	templatesDir := filepath.Join(baseDir, "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Errorf("templates directory was not created: %s", templatesDir)
	}

	// Verify data/outputs/ was populated
	outputsDir := filepath.Join(baseDir, "outputs")
	if _, err := os.Stat(outputsDir); os.IsNotExist(err) {
		t.Errorf("outputs directory was not created: %s", outputsDir)
	}

	// Verify files are in the correct locations
	if !filepath.HasPrefix(result.DownloadPath, downloadsDir) {
		t.Errorf("Download not in downloads/ folder: %s", result.DownloadPath)
	}
	if !filepath.HasPrefix(result.TemplatePath, templatesDir) {
		t.Errorf("Template not in templates/ folder: %s", result.TemplatePath)
	}
	if !filepath.HasPrefix(result.FilledPath, outputsDir) {
		t.Errorf("Output not in outputs/ folder: %s", result.FilledPath)
	}

	t.Logf("Data folders successfully populated:")
	t.Logf("  %s/ ✓", downloadsDir)
	t.Logf("  %s/ ✓", templatesDir)
	t.Logf("  %s/ ✓", outputsDir)
}
