package commands

import (
	"fmt"
	"path/filepath"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

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
// Emits events: fill.started, fill.completed, fill.error
func Fill(opts FillOptions) (*FillResult, error) {
	// Emit started event
	Emit(EventFillStarted, map[string]interface{}{
		"data_path":  opts.DataPath,
		"output_dir": opts.OutputDir,
		"flatten":    opts.Flatten,
	})

	// Determine output path using helper
	outputPath := DetermineOutputPath(opts.DataPath, opts.OutputDir, FilledPDFSuffix)

	// Ensure output directory exists using helper
	if err := EnsureOutputDir(outputPath); err != nil {
		EmitStageError(EventFillError, StageCreateDir, err, map[string]interface{}{
			"data_path": opts.DataPath,
		})
		return nil, err
	}

	// Fill the PDF
	inputPDF, err := pdfform.FillPDFFromJSON(opts.DataPath, outputPath)
	if err != nil {
		EmitStageError(EventFillError, StageFillPDF, err, map[string]interface{}{
			"data_path": opts.DataPath,
		})
		return nil, fmt.Errorf("failed to fill PDF: %w", err)
	}

	result := &FillResult{
		OutputPath: outputPath,
		InputPDF:   inputPDF,
		Flattened:  false,
	}

	// Flatten if requested
	if opts.Flatten {
		flatPath := outputPath[:len(outputPath)-len(filepath.Ext(outputPath))] + FlatPDFSuffix
		if err := pdfform.FlattenPDF(outputPath, flatPath); err != nil {
			EmitStageError(EventFillError, StageFlatten, err, map[string]interface{}{
				"data_path": opts.DataPath,
			})
			return nil, fmt.Errorf("failed to flatten PDF: %w", err)
		}
		result.OutputPath = flatPath
		result.Flattened = true
	}

	// Emit completed event
	Emit(EventFillCompleted, map[string]interface{}{
		"data_path":   opts.DataPath,
		"output_path": result.OutputPath,
		"input_pdf":   result.InputPDF,
		"flattened":   result.Flattened,
	})

	return result, nil
}

// FillFromCaseOptions contains options for filling from a case
type FillFromCaseOptions struct {
	CasePath  string
	OutputDir string
	Flatten   bool
}

// FillFromCase fills a PDF form using data from a case file
// Emits events: fill.started, fill.completed, fill.error
func FillFromCase(casePath, outputDir string, flatten bool) (*FillResult, error) {
	// Emit started event
	Emit(EventFillStarted, map[string]interface{}{
		"case_path":  casePath,
		"output_dir": outputDir,
		"flatten":    flatten,
	})

	pdfResult, err := pdfform.FillFromCase(casePath, outputDir, flatten)
	if err != nil {
		EmitError(EventFillError, err, map[string]interface{}{
			"case_path": casePath,
			"stage":     "fill_from_case",
		})
		return nil, err
	}

	// Convert to commands.FillResult
	result := &FillResult{
		OutputPath: pdfResult.OutputPath,
		InputPDF:   pdfResult.InputPDF,
		Flattened:  pdfResult.Flattened,
	}

	// Emit completed event
	Emit(EventFillCompleted, map[string]interface{}{
		"case_path":   casePath,
		"output_path": result.OutputPath,
		"input_pdf":   result.InputPDF,
		"flattened":   result.Flattened,
	})

	return result, nil
}
