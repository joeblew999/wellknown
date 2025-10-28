package calendar

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/joeblew999/wellknown/pkg/types"
)

// NOTE: Filenames must match schema.ExamplesFilename and schema.FailuresFilename constants
// but go:embed requires literal strings (can't use constants)
//
//go:embed data-examples.json
var examplesData []byte

//go:embed data-failures.json
var failuresData []byte

// TestGenerateURL_ValidExamples tests all valid examples from data-examples.json
func TestGenerateURL_ValidExamples(t *testing.T) {
	var examples struct {
		Examples []types.ShowcaseExample `json:"examples"`
	}

	if err := json.Unmarshal(examplesData, &examples); err != nil {
		t.Fatalf("Failed to parse data-examples.json: %v", err)
	}

	for _, example := range examples.Examples {
		t.Run(example.Name, func(t *testing.T) {
			// Generate URL from example data
			url, err := GenerateURL(example.Data)
			if err != nil {
				t.Fatalf("GenerateURL failed: %v", err)
			}

			// Verify URL structure
			if !strings.HasPrefix(url, BaseURL + "?") {
				t.Errorf("URL should start with Google Calendar base URL\nGot: %s", url)
			}

			// Verify required parameters
			if !strings.Contains(url, QueryParamAction + "=" + ActionParam) {
				t.Errorf("URL missing action=TEMPLATE\nGot: %s", url)
			}

			// Verify title is in URL
			if _, ok := example.Data["title"].(string); ok {
				if !strings.Contains(url, FieldMapping["title"] + "=") {
					t.Errorf("URL missing title parameter\nGot: %s", url)
				}
			}

			t.Logf("✅ Generated URL (%d bytes): %s", len(url), url)
		})
	}
}

// TestGenerateURL_InvalidCases tests all invalid cases from data-failures.json
func TestGenerateURL_InvalidCases(t *testing.T) {
	var failures struct {
		InvalidCases []struct {
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			Input       map[string]interface{} `json:"input"`
			ExpectError string                 `json:"expect_error"`
		} `json:"invalid_cases"`
	}

	if err := json.Unmarshal(failuresData, &failures); err != nil {
		t.Fatalf("Failed to parse data-failures.json: %v", err)
	}

	for _, testCase := range failures.InvalidCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// Attempt to generate URL
			_, err := GenerateURL(testCase.Input)

			// Should fail
			if err == nil {
				t.Fatalf("Expected error but got success")
			}

			// Check error message contains expected text
			if !strings.Contains(err.Error(), testCase.ExpectError) {
				t.Errorf("Expected error containing %q\nGot: %v", testCase.ExpectError, err)
			}

			t.Logf("✅ Correctly rejected invalid data: %v", err)
		})
	}
}

// TestGenerateURL_EdgeCases tests all edge cases from data-failures.json
func TestGenerateURL_EdgeCases(t *testing.T) {
	var failures struct {
		EdgeCases []struct {
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			Input       map[string]interface{} `json:"input"`
			Expect      struct {
				URLContains []string `json:"url_contains"`
			} `json:"expect"`
		} `json:"edge_cases"`
	}

	if err := json.Unmarshal(failuresData, &failures); err != nil {
		t.Fatalf("Failed to parse data-failures.json: %v", err)
	}

	for _, testCase := range failures.EdgeCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// Generate URL
			url, err := GenerateURL(testCase.Input)
			if err != nil {
				t.Fatalf("GenerateURL failed: %v", err)
			}

			// Verify expected strings are in URL
			for _, expected := range testCase.Expect.URLContains {
				if !strings.Contains(url, expected) {
					t.Errorf("URL missing expected string: %q\nGot: %s", expected, url)
				}
			}

			t.Logf("✅ Edge case handled: %s", testCase.Description)
		})
	}
}
