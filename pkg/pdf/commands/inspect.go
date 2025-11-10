package commands

import (
	"fmt"
	"os"
	"path/filepath"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

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
// Emits events: inspect.started, inspect.completed, inspect.error
func Inspect(opts InspectOptions) (*InspectResult, error) {
	// Emit started event
	Emit(EventInspectStarted, map[string]interface{}{
		"pdf_path":   opts.PDFPath,
		"output_dir": opts.OutputDir,
	})

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
		EmitError(EventInspectError, err, map[string]interface{}{
			"pdf_path": opts.PDFPath,
			"stage":    "create_dir",
		})
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract form fields
	fields, err := pdfform.ListFormFields(opts.PDFPath)
	if err != nil {
		EmitError(EventInspectError, err, map[string]interface{}{
			"pdf_path": opts.PDFPath,
			"stage":    "list_fields",
		})
		return nil, fmt.Errorf("failed to list form fields: %w", err)
	}

	// Export to JSON template
	if err := pdfform.ExportFormFieldsToJSON(opts.PDFPath, outputPath); err != nil {
		EmitError(EventInspectError, err, map[string]interface{}{
			"pdf_path": opts.PDFPath,
			"stage":    "export_json",
		})
		return nil, fmt.Errorf("failed to export form fields: %w", err)
	}

	// Update provenance metadata with inspection timestamp
	if err := pdfform.AddInspectedTimestamp(opts.PDFPath); err != nil {
		// Ignore error if no metadata exists
	}

	// Collect field names
	fieldNames := make([]string, len(fields))
	for i, field := range fields {
		fieldNames[i] = field.Name
	}

	result := &InspectResult{
		TemplatePath: outputPath,
		FieldCount:   len(fields),
		Fields:       fieldNames,
	}

	// Emit completed event
	Emit(EventInspectCompleted, map[string]interface{}{
		"pdf_path":      opts.PDFPath,
		"template_path": outputPath,
		"field_count":   len(fields),
	})

	return result, nil
}
