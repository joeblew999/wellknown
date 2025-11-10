package pdfform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/benoitkugler/pdf/formfill"
	"github.com/benoitkugler/pdf/reader"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/form"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// FormData represents the JSON structure with optional PDF URL and form field data
type FormData struct {
	PdfURL     string            `json:"pdf_url,omitempty"`
	Provenance *Provenance       `json:"provenance,omitempty"`
	Fields     map[string]string `json:"fields"`
}

// isURL checks if a string is a valid HTTP/HTTPS URL
func isURL(str string) bool {
	str = strings.TrimSpace(str)
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// DownloadPDF downloads a PDF from a URL to a local file
func DownloadPDF(pdfURL, outputPath string) error {
	resp, err := http.Get(pdfURL)
	if err != nil {
		return fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download PDF: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	return nil
}

// FillPDF fills an existing fillable PDF form with data from a JSON file
// Uses pdfcpu library
func FillPDF(inputPDF, jsonFile, outputPDF string) error {
	conf := model.NewDefaultConfiguration()
	if err := api.FillFormFile(inputPDF, jsonFile, outputPDF, conf); err != nil {
		return fmt.Errorf("failed to fill PDF: %w", err)
	}
	return nil
}

// FillPDFBenoitkugler fills a PDF using the benoitkugler/pdf library
// This library may work better with some PDFs, especially signed ones
func FillPDFBenoitkugler(inputPDF string, fields map[string]string, outputPDF string) error {
	// Open and parse the PDF
	f, err := os.Open(inputPDF)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	doc, _, err := reader.ParsePDFReader(f, reader.Options{})
	if err != nil {
		return fmt.Errorf("failed to parse PDF: %w", err)
	}

	// Create FDF dictionary from fields map
	var fdfFields []formfill.FDFField
	for name, value := range fields {
		fdfFields = append(fdfFields, formfill.FDFField{
			T: name,
			Values: formfill.Values{
				V: formfill.FDFText(value),
			},
		})
	}

	fdfDict := formfill.FDFDict{
		Fields: fdfFields,
	}

	// Fill the form (lockForm=false to keep it editable)
	if err := formfill.FillForm(&doc, fdfDict, false); err != nil {
		return fmt.Errorf("failed to fill form: %w", err)
	}

	// Write the filled PDF
	if err := doc.WriteFile(outputPDF, nil); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	return nil
}

// FillPDFWithFallback tries to fill a PDF using pdfcpu first, then falls back to benoitkugler
func FillPDFWithFallback(inputPDF string, fields map[string]string, outputPDF string) error {
	// Create temporary JSON for pdfcpu
	tempJSON := filepath.Join(os.TempDir(), "form_fields.json")
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}
	if err := os.WriteFile(tempJSON, fieldsJSON, 0644); err != nil {
		return fmt.Errorf("failed to write temp JSON: %w", err)
	}
	defer os.Remove(tempJSON)

	// Try pdfcpu first
	err = FillPDF(inputPDF, tempJSON, outputPDF)
	if err == nil {
		return nil // Success with pdfcpu
	}

	// If pdfcpu failed, try benoitkugler
	fmt.Printf("⚠️  pdfcpu failed (%v), trying benoitkugler library...\n", err)
	return FillPDFBenoitkugler(inputPDF, fields, outputPDF)
}

// FillPDFFromJSON fills a PDF using structured JSON data (with optional PDF URL or local path)
func FillPDFFromJSON(jsonFile, outputPDF string) (inputPDF string, err error) {
	// Read and parse JSON
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return "", fmt.Errorf("failed to read JSON file: %w", err)
	}

	var formData FormData
	if err := json.Unmarshal(data, &formData); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Handle pdf_url field - can be URL or local file path
	if formData.PdfURL != "" {
		if isURL(formData.PdfURL) {
			// It's a URL - download it
			inputPDF = filepath.Join(os.TempDir(), "form_template.pdf")
			if err := DownloadPDF(formData.PdfURL, inputPDF); err != nil {
				return "", err
			}
		} else {
			// It's a local file path - use it directly
			inputPDF = formData.PdfURL
			// Verify file exists
			if _, err := os.Stat(inputPDF); err != nil {
				return "", fmt.Errorf("PDF file not found: %s: %w", inputPDF, err)
			}
		}
	} else {
		return "", fmt.Errorf("pdf_url is required in JSON data")
	}

	// Create a temporary JSON file with just the fields for pdfcpu
	tempJSON := filepath.Join(os.TempDir(), "form_data.json")
	fieldsJSON, err := json.Marshal(formData.Fields)
	if err != nil {
		return inputPDF, fmt.Errorf("failed to marshal fields: %w", err)
	}
	if err := os.WriteFile(tempJSON, fieldsJSON, 0644); err != nil {
		return inputPDF, fmt.Errorf("failed to write temp JSON: %w", err)
	}
	defer os.Remove(tempJSON)

	// Fill the PDF using fallback approach
	if err := FillPDFWithFallback(inputPDF, formData.Fields, outputPDF); err != nil {
		return inputPDF, fmt.Errorf("failed to fill PDF: %w", err)
	}

	return inputPDF, nil
}

// FlattenPDF flattens a filled PDF (makes it read-only by locking all form fields)
func FlattenPDF(inputPDF, outputPDF string) error {
	conf := model.NewDefaultConfiguration()
	// Lock all form fields (pass nil to lock all fields)
	if err := api.LockFormFieldsFile(inputPDF, outputPDF, nil, conf); err != nil {
		return fmt.Errorf("failed to flatten PDF: %w", err)
	}
	return nil
}

// DefaultOutputName returns the default name for the filled PDF
func DefaultOutputName(inputPDF string) string {
	base := filepath.Base(inputPDF)
	name := base[:len(base)-len(filepath.Ext(base))]
	return name + "_filled.pdf"
}

// ListFormFields extracts all form fields from a PDF
func ListFormFields(inputPDF string) ([]form.Field, error) {
	f, err := os.Open(inputPDF)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	conf := model.NewDefaultConfiguration()
	fields, err := api.FormFields(f, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to extract form fields: %w", err)
	}

	return fields, nil
}

// ExportFormFieldsToJSON extracts form fields and exports them as a JSON template
// If provenance metadata exists, it will be included in the template
func ExportFormFieldsToJSON(inputPDF, outputJSON string) error {
	fields, err := ListFormFields(inputPDF)
	if err != nil {
		return err
	}

	// Create a map with field names as keys and empty strings as values
	fieldMap := make(map[string]string)
	for _, field := range fields {
		// Use the field name (full path) as the key
		fieldMap[field.Name] = ""
	}

	// Try to load provenance metadata if it exists
	prov, _ := LoadProvenanceMetadata(inputPDF)

	// Create the template structure
	template := FormData{
		PdfURL:     "", // User should fill this in
		Provenance: prov,
		Fields:     fieldMap,
	}

	// Write to JSON file
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputJSON, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}
