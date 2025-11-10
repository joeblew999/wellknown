package pdfform

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TransferForm represents a government transfer form
type TransferForm struct {
	State           string
	FormName        string
	FormCode        string
	Description     string
	Format          string // PDF, DOCX, etc.
	DirectPDFURL    string
	InfoURL         string
	OnlineAvailable bool
	Notes           string
}

// FormsCatalog holds a collection of transfer forms
type FormsCatalog struct {
	Forms []TransferForm
}

// LoadFormsCatalog loads transfer forms from a CSV file
func LoadFormsCatalog(csvPath string) (*FormsCatalog, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or missing header")
	}

	// Skip header row
	catalog := &FormsCatalog{
		Forms: make([]TransferForm, 0, len(records)-1),
	}

	for i, record := range records[1:] {
		if len(record) < 9 {
			return nil, fmt.Errorf("row %d has insufficient columns", i+2)
		}

		onlineAvailable, _ := strconv.ParseBool(strings.ToLower(strings.TrimSpace(record[7])))

		form := TransferForm{
			State:           strings.TrimSpace(record[0]),
			FormName:        strings.TrimSpace(record[1]),
			FormCode:        strings.TrimSpace(record[2]),
			Description:     strings.TrimSpace(record[3]),
			Format:          strings.TrimSpace(record[4]),
			DirectPDFURL:    strings.TrimSpace(record[5]),
			InfoURL:         strings.TrimSpace(record[6]),
			OnlineAvailable: onlineAvailable,
			Notes:           strings.TrimSpace(record[8]),
		}

		catalog.Forms = append(catalog.Forms, form)
	}

	return catalog, nil
}

// GetFormsByState returns all forms for a specific state
func (c *FormsCatalog) GetFormsByState(state string) []TransferForm {
	state = strings.ToUpper(strings.TrimSpace(state))
	var forms []TransferForm
	for _, form := range c.Forms {
		if strings.ToUpper(form.State) == state {
			forms = append(forms, form)
		}
	}
	return forms
}

// GetFormByCode returns a form by its form code
func (c *FormsCatalog) GetFormByCode(code string) *TransferForm {
	code = strings.ToUpper(strings.TrimSpace(code))
	for _, form := range c.Forms {
		if strings.ToUpper(form.FormCode) == code {
			return &form
		}
	}
	return nil
}

// ListStates returns a list of all states in the catalog
func (c *FormsCatalog) ListStates() []string {
	stateMap := make(map[string]bool)
	for _, form := range c.Forms {
		stateMap[form.State] = true
	}

	states := make([]string, 0, len(stateMap))
	for state := range stateMap {
		states = append(states, state)
	}
	return states
}

// GetPDFForms returns only forms that are available as PDFs
func (c *FormsCatalog) GetPDFForms() []TransferForm {
	var pdfForms []TransferForm
	for _, form := range c.Forms {
		if strings.ToUpper(form.Format) == "PDF" {
			pdfForms = append(pdfForms, form)
		}
	}
	return pdfForms
}

// DownloadFormPDF downloads a form PDF to the specified directory
func (c *FormsCatalog) DownloadFormPDF(form *TransferForm, outputDir string) (string, error) {
	if form.DirectPDFURL == "" {
		return "", fmt.Errorf("form has no direct PDF URL")
	}

	// Create a filename from the form code or name
	filename := form.FormCode
	if filename == "" {
		filename = strings.ReplaceAll(form.FormName, " ", "_")
	}
	filename = strings.ToLower(filename) + ".pdf"

	outputPath := filepath.Join(outputDir, filename)
	if err := DownloadPDF(form.DirectPDFURL, outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}
