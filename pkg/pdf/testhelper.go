package pdfform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TestCase represents a test scenario for PDF filling
type TestCase struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	PdfURL      string   `json:"pdf_url"`
	Fields      map[string]string `json:"fields"`
	ExpectError bool     `json:"expect_error,omitempty"`
}

// TestSuite represents a collection of test cases
type TestSuite struct {
	Name  string     `json:"name"`
	Cases []TestCase `json:"cases"`
}

// LoadTestCase loads a test case from a JSON file
func LoadTestCase(testFile string) (*TestCase, error) {
	data, err := os.ReadFile(testFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read test file: %w", err)
	}

	var testCase TestCase
	if err := json.Unmarshal(data, &testCase); err != nil {
		return nil, fmt.Errorf("failed to parse test case: %w", err)
	}

	return &testCase, nil
}

// LoadTestSuite loads a test suite from a JSON file
func LoadTestSuite(suiteFile string) (*TestSuite, error) {
	data, err := os.ReadFile(suiteFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read test suite: %w", err)
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse test suite: %w", err)
	}

	return &suite, nil
}

// SaveTestCase saves a test case to a JSON file
func SaveTestCase(testCase *TestCase, outputFile string) error {
	data, err := json.MarshalIndent(testCase, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test case: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write test case: %w", err)
	}

	return nil
}

// RunTestCase executes a test case and returns the output PDF path
func RunTestCase(testCase *TestCase, outputDir string) (string, error) {
	// Create FormData from TestCase
	formData := FormData{
		PdfURL: testCase.PdfURL,
		Fields: testCase.Fields,
	}

	// Create temp JSON file with test data
	tempJSON := filepath.Join(os.TempDir(), "test_case.json")
	data, err := json.Marshal(formData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal form data: %w", err)
	}
	if err := os.WriteFile(tempJSON, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp JSON: %w", err)
	}
	defer os.Remove(tempJSON)

	// Generate output filename
	outputPDF := filepath.Join(outputDir, testCase.Name+"_filled.pdf")

	// Fill the PDF
	_, err = FillPDFFromJSON(tempJSON, outputPDF)
	if err != nil {
		if testCase.ExpectError {
			// Expected error - return it
			return "", err
		}
		return "", fmt.Errorf("failed to fill PDF: %w", err)
	}

	if testCase.ExpectError {
		return outputPDF, fmt.Errorf("expected error but got success")
	}

	return outputPDF, nil
}

// ListTestCases lists all test case files in a directory
func ListTestCases(testDir string) ([]string, error) {
	pattern := filepath.Join(testDir, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list test cases: %w", err)
	}
	return matches, nil
}
