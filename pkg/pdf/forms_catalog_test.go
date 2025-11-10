package pdfform_test

import (
	"os"
	"path/filepath"
	"testing"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

func TestLoadFormsCatalog(t *testing.T) {
	// Use the actual CSV file in the package
	csvPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")

	catalog, err := pdfform.LoadFormsCatalog(csvPath)
	if err != nil {
		t.Fatalf("Failed to load forms catalog: %v", err)
	}

	if len(catalog.Forms) == 0 {
		t.Error("Expected at least one form in catalog")
	}

	t.Logf("Loaded %d forms from catalog", len(catalog.Forms))
}

func TestGetFormsByState(t *testing.T) {
	csvPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")

	catalog, err := pdfform.LoadFormsCatalog(csvPath)
	if err != nil {
		t.Fatalf("Failed to load forms catalog: %v", err)
	}

	// Test VIC forms
	vicForms := catalog.GetFormsByState("VIC")
	if len(vicForms) == 0 {
		t.Error("Expected at least one VIC form")
	}

	for _, form := range vicForms {
		if form.State != "VIC" {
			t.Errorf("Expected state VIC, got %s", form.State)
		}
	}

	t.Logf("Found %d forms for VIC", len(vicForms))
}

func TestGetFormByCode(t *testing.T) {
	csvPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")

	catalog, err := pdfform.LoadFormsCatalog(csvPath)
	if err != nil {
		t.Fatalf("Failed to load forms catalog: %v", err)
	}

	// Test known form code
	form := catalog.GetFormByCode("F3520")
	if form == nil {
		t.Error("Expected to find form F3520 (QLD)")
	} else {
		if form.State != "QLD" {
			t.Errorf("Expected QLD form, got %s", form.State)
		}
		t.Logf("Found form: %s - %s", form.FormCode, form.FormName)
	}

	// Test non-existent form
	form = catalog.GetFormByCode("NONEXISTENT")
	if form != nil {
		t.Error("Expected nil for non-existent form code")
	}
}

func TestListStates(t *testing.T) {
	csvPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")

	catalog, err := pdfform.LoadFormsCatalog(csvPath)
	if err != nil {
		t.Fatalf("Failed to load forms catalog: %v", err)
	}

	states := catalog.ListStates()
	if len(states) < 6 {
		t.Errorf("Expected at least 6 states, got %d", len(states))
	}

	t.Logf("States in catalog: %v", states)
}

func TestGetPDFForms(t *testing.T) {
	csvPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")

	catalog, err := pdfform.LoadFormsCatalog(csvPath)
	if err != nil {
		t.Fatalf("Failed to load forms catalog: %v", err)
	}

	pdfForms := catalog.GetPDFForms()
	if len(pdfForms) == 0 {
		t.Error("Expected at least one PDF form")
	}

	for _, form := range pdfForms {
		if form.Format != "PDF" {
			t.Errorf("Expected PDF format, got %s", form.Format)
		}
	}

	t.Logf("Found %d PDF forms", len(pdfForms))
}

func TestDownloadFormPDF(t *testing.T) {
	// Skip this test by default as it downloads from the internet
	if os.Getenv("ENABLE_DOWNLOAD_TESTS") != "true" {
		t.Skip("Skipping download test. Set ENABLE_DOWNLOAD_TESTS=true to run")
	}

	csvPath := filepath.Join("data", "catalog", "australian_transfer_forms.csv")

	catalog, err := pdfform.LoadFormsCatalog(csvPath)
	if err != nil {
		t.Fatalf("Failed to load forms catalog: %v", err)
	}

	// Find a form with a direct PDF URL
	var testForm *pdfform.TransferForm
	for i := range catalog.Forms {
		if catalog.Forms[i].DirectPDFURL != "" {
			testForm = &catalog.Forms[i]
			break
		}
	}

	if testForm == nil {
		t.Skip("No forms with direct PDF URLs found")
	}

	// Create temp directory
	tempDir := t.TempDir()

	// Download the form
	outputPath, err := catalog.DownloadFormPDF(testForm, tempDir)
	if err != nil {
		t.Fatalf("Failed to download form: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Downloaded file does not exist: %s", outputPath)
	}

	t.Logf("Successfully downloaded form to: %s", outputPath)
}
