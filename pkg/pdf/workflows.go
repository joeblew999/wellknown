package pdfform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WorkflowOptions contains options for running the complete workflow
type WorkflowOptions struct {
	CatalogPath string
	FormCode    string
	FieldData   map[string]string
	OutputDir   string
	Flatten     bool
}

// WorkflowResult contains the results of running the complete workflow
type WorkflowResult struct {
	FormCode     string
	DownloadPath string
	TemplatePath string
	FilledPath   string
	Provenance   *Provenance
}

// RunWorkflow executes the complete 5-step workflow: browse → download → inspect → fill
// This is a higher-level orchestration that combines individual commands
func RunWorkflow(opts WorkflowOptions) (*WorkflowResult, error) {
	result := &WorkflowResult{
		FormCode: opts.FormCode,
	}

	// Step 1: Browse - verify form exists
	catalog, err := LoadFormsCatalog(opts.CatalogPath)
	if err != nil {
		return nil, fmt.Errorf("step 1 (browse) failed: %w", err)
	}

	form := catalog.GetFormByCode(opts.FormCode)
	if form == nil {
		return nil, fmt.Errorf("step 1 (browse) failed: form '%s' not found", opts.FormCode)
	}

	// Step 2: Download
	downloadDir := filepath.Join(opts.OutputDir, "downloads")
	downloadResult, err := Download(DownloadOptions{
		CatalogPath: opts.CatalogPath,
		FormCode:    opts.FormCode,
		OutputDir:   downloadDir,
	})
	if err != nil {
		return nil, fmt.Errorf("step 2 (download) failed: %w", err)
	}
	result.DownloadPath = downloadResult.PDFPath

	// Step 3: Inspect
	templateDir := filepath.Join(opts.OutputDir, "templates")
	inspectResult, err := Inspect(InspectOptions{
		PDFPath:   downloadResult.PDFPath,
		OutputDir: templateDir,
	})
	if err != nil {
		return nil, fmt.Errorf("step 3 (inspect) failed: %w", err)
	}
	result.TemplatePath = inspectResult.TemplatePath

	// Step 4: Fill - Create temporary JSON with field data
	tempDir := filepath.Join(opts.OutputDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("step 4 (fill) failed to create temp dir: %w", err)
	}

	// Create FormData with the field data
	formData := FormData{
		PdfURL: downloadResult.PDFPath,
		Fields: opts.FieldData,
	}

	// Save temporary JSON
	tempJSON := filepath.Join(tempDir, "temp_data.json")
	data, err := json.Marshal(formData)
	if err != nil {
		return nil, fmt.Errorf("step 4 (fill) failed to marshal data: %w", err)
	}
	if err := os.WriteFile(tempJSON, data, 0644); err != nil {
		return nil, fmt.Errorf("step 4 (fill) failed to write temp JSON: %w", err)
	}
	defer os.Remove(tempJSON)

	// Fill the form
	outputsDir := filepath.Join(opts.OutputDir, "outputs")
	fillResult, err := Fill(FillOptions{
		DataPath:  tempJSON,
		OutputDir: outputsDir,
		Flatten:   opts.Flatten,
	})
	if err != nil {
		return nil, fmt.Errorf("step 4 (fill) failed: %w", err)
	}
	result.FilledPath = fillResult.OutputPath

	// Load provenance
	if prov, err := LoadProvenanceMetadata(downloadResult.PDFPath); err == nil {
		result.Provenance = prov
	}

	return result, nil
}

// BulkWorkflowOptions contains options for processing multiple forms
type BulkWorkflowOptions struct {
	CatalogPath string
	FormCodes   []string
	FieldData   map[string]map[string]string // FormCode -> Fields
	OutputDir   string
	Flatten     bool
}

// BulkWorkflowResult contains results for each form processed
type BulkWorkflowResult struct {
	Results  map[string]*WorkflowResult // FormCode -> Result
	Errors   map[string]error           // FormCode -> Error
	Total    int
	Success  int
	Failed   int
}

// RunBulkWorkflow processes multiple forms in a single operation
func RunBulkWorkflow(opts BulkWorkflowOptions) (*BulkWorkflowResult, error) {
	result := &BulkWorkflowResult{
		Results: make(map[string]*WorkflowResult),
		Errors:  make(map[string]error),
		Total:   len(opts.FormCodes),
	}

	for _, formCode := range opts.FormCodes {
		// Get field data for this form (or use empty map)
		fieldData := opts.FieldData[formCode]
		if fieldData == nil {
			fieldData = make(map[string]string)
		}

		// Run workflow for this form
		workflowResult, err := RunWorkflow(WorkflowOptions{
			CatalogPath: opts.CatalogPath,
			FormCode:    formCode,
			FieldData:   fieldData,
			OutputDir:   opts.OutputDir,
			Flatten:     opts.Flatten,
		})

		if err != nil {
			result.Errors[formCode] = err
			result.Failed++
		} else {
			result.Results[formCode] = workflowResult
			result.Success++
		}
	}

	return result, nil
}

// UpdateWorkflowOptions contains options for updating an existing form
type UpdateWorkflowOptions struct {
	CatalogPath    string
	FormCode       string
	ExistingPDFDir string
	OutputDir      string
	ForceUpdate    bool
}

// UpdateWorkflowResult contains results of the update workflow
type UpdateWorkflowResult struct {
	FormCode     string
	Updated      bool
	OldPath      string
	NewPath      string
	OldProvenance *Provenance
	NewProvenance *Provenance
}

// RunUpdateWorkflow checks if a form needs updating and re-downloads if necessary
func RunUpdateWorkflow(opts UpdateWorkflowOptions) (*UpdateWorkflowResult, error) {
	result := &UpdateWorkflowResult{
		FormCode: opts.FormCode,
	}

	// Find existing PDF
	catalog, err := LoadFormsCatalog(opts.CatalogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load catalog: %w", err)
	}

	form := catalog.GetFormByCode(opts.FormCode)
	if form == nil {
		return nil, fmt.Errorf("form '%s' not found in catalog", opts.FormCode)
	}

	// Look for existing PDF in the directory
	existingPDFPath := filepath.Join(opts.ExistingPDFDir, form.FormCode+".pdf")
	if _, err := os.Stat(existingPDFPath); err == nil {
		result.OldPath = existingPDFPath
		// Load old provenance
		if prov, err := LoadProvenanceMetadata(existingPDFPath); err == nil {
			result.OldProvenance = prov
		}
	}

	// Check if update is needed
	needsUpdate := opts.ForceUpdate || result.OldPath == ""

	if !needsUpdate {
		result.Updated = false
		result.NewPath = result.OldPath
		result.NewProvenance = result.OldProvenance
		return result, nil
	}

	// Download new version
	downloadDir := opts.OutputDir
	if downloadDir == "" {
		downloadDir = opts.ExistingPDFDir
	}

	downloadResult, err := Download(DownloadOptions{
		CatalogPath: opts.CatalogPath,
		FormCode:    opts.FormCode,
		OutputDir:   downloadDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download updated form: %w", err)
	}

	result.Updated = true
	result.NewPath = downloadResult.PDFPath

	// Load new provenance
	if prov, err := LoadProvenanceMetadata(downloadResult.PDFPath); err == nil {
		result.NewProvenance = prov
	}

	return result, nil
}
