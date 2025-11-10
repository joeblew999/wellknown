package pdfform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CaseMetadata contains metadata about a case
type CaseMetadata struct {
	CaseID    string    `json:"case_id"`
	CaseName  string    `json:"case_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// FormReference contains information about the form to fill
type FormReference struct {
	FormCode     string `json:"form_code"`
	TemplatePath string `json:"template_path,omitempty"`
}

// ValidationStatus contains validation results for the case
type ValidationStatus struct {
	Valid          bool     `json:"valid"`
	MissingFields  []string `json:"missing_fields,omitempty"`
	InvalidFields  []string `json:"invalid_fields,omitempty"`
	ValidationTime time.Time `json:"validation_time,omitempty"`
}

// Case represents a complete case with metadata, form reference, and field data
type Case struct {
	Metadata      CaseMetadata      `json:"case_metadata"`
	FormReference FormReference     `json:"form_reference"`
	Fields        map[string]string `json:"fields"`
	Validation    *ValidationStatus `json:"validation,omitempty"`
}

// LoadCase loads a case from a JSON file
func LoadCase(casePath string) (*Case, error) {
	data, err := os.ReadFile(casePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read case file: %w", err)
	}

	var c Case
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse case JSON: %w", err)
	}

	return &c, nil
}

// SaveCase saves a case to a JSON file
func SaveCase(c *Case, casePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(casePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create case directory: %w", err)
	}

	// Update timestamp
	c.Metadata.UpdatedAt = time.Now()

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal case JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(casePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write case file: %w", err)
	}

	return nil
}

// CreateCase creates a new case with the given form code and saves it
func CreateCase(formCode, caseName, entityName string, dataDir string) (*Case, string, error) {
	// Generate case ID with microseconds for uniqueness
	timestamp := time.Now().Format("20060102_150405.000000")
	caseID := fmt.Sprintf("%s_%s_%s", entityName, formCode, timestamp)

	// Create case structure
	c := &Case{
		Metadata: CaseMetadata{
			CaseID:    caseID,
			CaseName:  caseName,
			CreatedAt: time.Now(),
		},
		FormReference: FormReference{
			FormCode: formCode,
		},
		Fields: make(map[string]string),
	}

	// Determine case file path
	casePath := filepath.Join(dataDir, "cases", entityName, caseID+".json")

	// Save the case
	if err := SaveCase(c, casePath); err != nil {
		return nil, "", err
	}

	return c, casePath, nil
}

// FillFromCase fills a PDF form using data from a case file
func FillFromCase(casePath, outputDir string, flatten bool) (*FillResult, error) {
	// Load the case
	c, err := LoadCase(casePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load case: %w", err)
	}

	// Create FormData from case
	formData := FormData{
		Fields: c.Fields,
	}

	// Determine PDF path from form code or template path
	var pdfPath string
	if c.FormReference.TemplatePath != "" {
		// Extract PDF path from template
		templateData, err := os.ReadFile(c.FormReference.TemplatePath)
		if err == nil {
			var template FormData
			if err := json.Unmarshal(templateData, &template); err == nil {
				pdfPath = template.PdfURL
			}
		}
	}

	if pdfPath == "" {
		return nil, fmt.Errorf("cannot determine PDF path from case")
	}

	formData.PdfURL = pdfPath

	// Create temporary JSON file for Fill function
	tempDir := filepath.Join(filepath.Dir(casePath), ".temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempJSON := filepath.Join(tempDir, "temp_fill.json")
	defer os.Remove(tempJSON)

	data, err := json.Marshal(formData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal form data: %w", err)
	}

	if err := os.WriteFile(tempJSON, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp JSON: %w", err)
	}

	// Use Fill function
	return Fill(FillOptions{
		DataPath:  tempJSON,
		OutputDir: outputDir,
		Flatten:   flatten,
	})
}

// ListCases lists all case files for a given entity (or all if entityName is empty)
func ListCases(dataDir, entityName string) ([]string, error) {
	casesDir := filepath.Join(dataDir, "cases")

	if entityName != "" {
		casesDir = filepath.Join(casesDir, entityName)
	}

	// Check if directory exists
	if _, err := os.Stat(casesDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	var cases []string

	// Walk the directory
	err := filepath.Walk(casesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-JSON files
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		cases = append(cases, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list cases: %w", err)
	}

	return cases, nil
}

// ValidateCase validates a case against its form template
func ValidateCase(c *Case, catalogPath string) error {
	if c.Validation == nil {
		c.Validation = &ValidationStatus{}
	}

	c.Validation.Valid = true
	c.Validation.MissingFields = []string{}
	c.Validation.InvalidFields = []string{}
	c.Validation.ValidationTime = time.Now()

	// Load template if path is provided
	if c.FormReference.TemplatePath != "" {
		templateData, err := os.ReadFile(c.FormReference.TemplatePath)
		if err != nil {
			return fmt.Errorf("failed to read template: %w", err)
		}

		var template FormData
		if err := json.Unmarshal(templateData, &template); err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		// Check for missing required fields
		for fieldName := range template.Fields {
			if _, exists := c.Fields[fieldName]; !exists {
				c.Validation.MissingFields = append(c.Validation.MissingFields, fieldName)
				c.Validation.Valid = false
			}
		}
	}

	return nil
}
