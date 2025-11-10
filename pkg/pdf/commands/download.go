package commands

import (
	"fmt"
	"os"
	"path/filepath"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
)

// DownloadOptions contains options for downloading a form
type DownloadOptions struct {
	CatalogPath string
	FormCode    string
	OutputDir   string
}

// DownloadResult contains the results of downloading a form
type DownloadResult struct {
	PDFPath  string
	Form     *pdfform.TransferForm
	Metadata string // Path to .meta.json file
}

// Download downloads a form PDF by its code from the catalog
// Emits events: download.started, download.progress, download.completed, download.error
func Download(opts DownloadOptions) (*DownloadResult, error) {
	// Emit started event
	Emit(EventDownloadStarted, map[string]interface{}{
		"form_code":  opts.FormCode,
		"output_dir": opts.OutputDir,
	})

	catalog, err := pdfform.LoadFormsCatalog(opts.CatalogPath)
	if err != nil {
		EmitError(EventDownloadError, err, map[string]interface{}{
			"form_code": opts.FormCode,
			"stage":     "load_catalog",
		})
		return nil, fmt.Errorf("failed to load forms catalog: %w", err)
	}

	form := catalog.GetFormByCode(opts.FormCode)
	if form == nil {
		err := fmt.Errorf("form with code '%s' not found", opts.FormCode)
		EmitError(EventDownloadError, err, map[string]interface{}{
			"form_code": opts.FormCode,
			"stage":     "find_form",
		})
		return nil, err
	}

	if form.DirectPDFURL == "" {
		err := fmt.Errorf("form '%s' does not have a direct PDF URL", opts.FormCode)
		EmitError(EventDownloadError, err, map[string]interface{}{
			"form_code": opts.FormCode,
			"stage":     "check_url",
		})
		return nil, err
	}

	// Emit progress - found form
	Emit(EventDownloadProgress, map[string]interface{}{
		"form_code": opts.FormCode,
		"form_name": form.FormName,
		"state":     form.State,
		"stage":     DownloadStageFoundForm,
		"progress":  ProgressFoundForm,
	})

	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, DefaultDirPerm); err != nil {
		EmitStageError(EventDownloadError, DownloadStageCreateDir, err, map[string]interface{}{
			"form_code": opts.FormCode,
		})
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Emit progress - starting download
	Emit(EventDownloadProgress, map[string]interface{}{
		"form_code": opts.FormCode,
		"stage":     DownloadStageDownloading,
		"progress":  ProgressDownloading,
	})

	// Download the form
	pdfPath, err := catalog.DownloadFormPDF(form, opts.OutputDir)
	if err != nil {
		EmitStageError(EventDownloadError, DownloadStageDownloadPDF, err, map[string]interface{}{
			"form_code": opts.FormCode,
		})
		return nil, fmt.Errorf("failed to download form: %w", err)
	}

	// Emit progress - download complete, saving metadata
	Emit(EventDownloadProgress, map[string]interface{}{
		"form_code": opts.FormCode,
		"pdf_path":  pdfPath,
		"stage":     DownloadStageSavingMeta,
		"progress":  ProgressMetadata,
	})

	// Save provenance metadata
	if err := pdfform.SaveProvenanceMetadata(pdfPath, form.FormCode, form.State, form.DirectPDFURL); err != nil {
		// Don't fail the download, just warn
		fmt.Printf("⚠️  Warning: Could not save provenance metadata: %v\n", err)
	}

	metadataPath := pdfPath[:len(pdfPath)-len(filepath.Ext(pdfPath))] + MetaJSONSuffix

	result := &DownloadResult{
		PDFPath:  pdfPath,
		Form:     form,
		Metadata: metadataPath,
	}

	// Emit completed event
	Emit(EventDownloadCompleted, map[string]interface{}{
		"form_code": opts.FormCode,
		"pdf_path":  pdfPath,
		"form_name": form.FormName,
		"state":     form.State,
		"progress":  ProgressComplete,
	})

	return result, nil
}
