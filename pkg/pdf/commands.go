package pdfform

import (
	"fmt"
	"os"
	"path/filepath"
)

// BrowseOptions contains options for browsing the forms catalog
type BrowseOptions struct {
	CatalogPath string
	State       string
}

// BrowseResult contains the results of browsing the forms catalog
type BrowseResult struct {
	States []string
	Forms  []TransferForm
}

// Browse loads the forms catalog and returns either all states or forms for a specific state
func Browse(opts BrowseOptions) (*BrowseResult, error) {
	catalog, err := LoadFormsCatalog(opts.CatalogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load forms catalog: %w", err)
	}

	result := &BrowseResult{}

	if opts.State == "" {
		// Return all states
		result.States = catalog.ListStates()
	} else {
		// Return forms for specific state
		result.Forms = catalog.GetFormsByState(opts.State)
		if len(result.Forms) == 0 {
			return nil, fmt.Errorf("no forms found for state: %s", opts.State)
		}
	}

	return result, nil
}

// DownloadOptions contains options for downloading a form
type DownloadOptions struct {
	CatalogPath string
	FormCode    string
	OutputDir   string
}

// DownloadResult contains the results of downloading a form
type DownloadResult struct {
	PDFPath  string
	Form     *TransferForm
	Metadata string // Path to .meta.json file
}

// Download downloads a form PDF by its code from the catalog
func Download(opts DownloadOptions) (*DownloadResult, error) {
	catalog, err := LoadFormsCatalog(opts.CatalogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load forms catalog: %w", err)
	}

	form := catalog.GetFormByCode(opts.FormCode)
	if form == nil {
		return nil, fmt.Errorf("form with code '%s' not found", opts.FormCode)
	}

	if form.DirectPDFURL == "" {
		return nil, fmt.Errorf("form '%s' does not have a direct PDF URL", opts.FormCode)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Download the form
	pdfPath, err := catalog.DownloadFormPDF(form, opts.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download form: %w", err)
	}

	// Save provenance metadata
	if err := SaveProvenanceMetadata(pdfPath, form.FormCode, form.State, form.DirectPDFURL); err != nil {
		// Don't fail the download, just warn
		fmt.Printf("⚠️  Warning: Could not save provenance metadata: %v\n", err)
	}

	metadataPath := pdfPath[:len(pdfPath)-len(filepath.Ext(pdfPath))] + ".meta.json"

	return &DownloadResult{
		PDFPath:  pdfPath,
		Form:     form,
		Metadata: metadataPath,
	}, nil
}

// InspectOptions contains options for inspecting a PDF form
type InspectOptions struct {
	PDFPath   string
	OutputDir string
}

// InspectResult contains the results of inspecting a PDF form
type InspectResult struct {
	TemplatePath string
	FieldCount   int
	Fields       []string // Field names
}

// Inspect extracts form fields from a PDF and creates a JSON template
func Inspect(opts InspectOptions) (*InspectResult, error) {
	// Determine output path
	outputPath := opts.OutputDir
	if outputPath == "" {
		base := filepath.Base(opts.PDFPath)
		name := base[:len(base)-len(filepath.Ext(base))]
		outputPath = name + "_template.json"
	} else {
		// If OutputDir is a directory, create template filename
		if info, err := os.Stat(opts.OutputDir); err == nil && info.IsDir() {
			base := filepath.Base(opts.PDFPath)
			name := base[:len(base)-len(filepath.Ext(base))]
			outputPath = filepath.Join(opts.OutputDir, name+"_template.json")
		}
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract form fields
	fields, err := ListFormFields(opts.PDFPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list form fields: %w", err)
	}

	// Export to JSON template
	if err := ExportFormFieldsToJSON(opts.PDFPath, outputPath); err != nil {
		return nil, fmt.Errorf("failed to export form fields: %w", err)
	}

	// Update provenance metadata with inspection timestamp
	if err := AddInspectedTimestamp(opts.PDFPath); err != nil {
		// Ignore error if no metadata exists
	}

	// Collect field names
	fieldNames := make([]string, len(fields))
	for i, field := range fields {
		fieldNames[i] = field.Name
	}

	return &InspectResult{
		TemplatePath: outputPath,
		FieldCount:   len(fields),
		Fields:       fieldNames,
	}, nil
}

// FillOptions contains options for filling a PDF form
type FillOptions struct {
	DataPath  string
	OutputDir string
	Flatten   bool
}

// FillResult contains the results of filling a PDF form
type FillResult struct {
	OutputPath string
	InputPDF   string
	Flattened  bool
}

// Fill fills a PDF form using JSON data
func Fill(opts FillOptions) (*FillResult, error) {
	// Determine output path
	outputPath := opts.OutputDir
	if outputPath == "" {
		base := filepath.Base(opts.DataPath)
		name := base[:len(base)-len(filepath.Ext(base))]
		outputPath = name + "_filled.pdf"
	} else {
		// If OutputDir is a directory, create output filename
		if info, err := os.Stat(opts.OutputDir); err != nil || info.IsDir() {
			base := filepath.Base(opts.DataPath)
			name := base[:len(base)-len(filepath.Ext(base))]
			outputPath = filepath.Join(opts.OutputDir, name+"_filled.pdf")
		}
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Fill the PDF
	inputPDF, err := FillPDFFromJSON(opts.DataPath, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to fill PDF: %w", err)
	}

	result := &FillResult{
		OutputPath: outputPath,
		InputPDF:   inputPDF,
		Flattened:  false,
	}

	// Flatten if requested
	if opts.Flatten {
		flatPath := outputPath[:len(outputPath)-len(filepath.Ext(outputPath))] + "_flat.pdf"
		if err := FlattenPDF(outputPath, flatPath); err != nil {
			return nil, fmt.Errorf("failed to flatten PDF: %w", err)
		}
		result.OutputPath = flatPath
		result.Flattened = true
	}

	return result, nil
}

// TestOptions contains options for running a test case
type TestOptions struct {
	TestCasePath string
	OutputDir    string
}

// TestResult contains the results of running a test case
type TestResult struct {
	Name       string
	Passed     bool
	OutputPath string
	Error      error
}

// Test runs a test case and returns the result
func Test(opts TestOptions) (*TestResult, error) {
	// Load test case
	testCase, err := LoadTestCase(opts.TestCasePath)
	if err != nil {
		return &TestResult{
			Name:   filepath.Base(opts.TestCasePath),
			Passed: false,
			Error:  fmt.Errorf("failed to load test case: %w", err),
		}, err
	}

	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return &TestResult{
			Name:   testCase.Name,
			Passed: false,
			Error:  fmt.Errorf("failed to create output directory: %w", err),
		}, err
	}

	// Run test case
	outputPath, err := RunTestCase(testCase, opts.OutputDir)
	if err != nil {
		return &TestResult{
			Name:       testCase.Name,
			Passed:     false,
			OutputPath: outputPath,
			Error:      err,
		}, nil
	}

	return &TestResult{
		Name:       testCase.Name,
		Passed:     true,
		OutputPath: outputPath,
		Error:      nil,
	}, nil
}
