// Package testgen provides robust schema-driven test data generation for calendar platforms.
//
// This package generates Go-verified test expectations for Playwright E2E tests by:
// 1. Reading data-examples.json files
// 2. Running ACTUAL Go generator functions (GenerateURL, GenerateICS, etc.)
// 3. Capturing expected outputs
// 4. Writing test suites with expected results
//
// Key features:
// ✅ Zero code duplication (generic ProcessPlatform)
// ✅ Registry-based (easy to add platforms)
// ✅ Reflection for type-safe calls
// ✅ Schema validation integration
package testgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/joeblew999/wellknown/pkg/schema"
	googlecal "github.com/joeblew999/wellknown/pkg/google/calendar"
	applecal "github.com/joeblew999/wellknown/pkg/apple/calendar"
)

// Types (shared with old main.go)
type Example struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
}

type ExamplesFile struct {
	Examples []Example `json:"examples"`
}

type TestCase struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Data        map[string]interface{}  `json:"data"`
	Expected    ExpectedResult          `json:"expected"`
	Validation  schema.ValidationResult `json:"validation"`
	Tags        []string                `json:"tags,omitempty"`
}

type ExpectedResult struct {
	URL         string   `json:"url,omitempty"`
	ICS         string   `json:"ics,omitempty"`
	ICSContains []string `json:"ics_contains,omitempty"`
	Error       string   `json:"error,omitempty"`
}

type TestSuite struct {
	Platform       string                 `json:"platform"`
	AppType        string                 `json:"app_type"`
	GeneratedAt    string                 `json:"generated_at"`
	SourceFile     string                 `json:"source_file"`
	TestCases      []TestCase             `json:"test_cases"`
	Metadata       map[string]interface{} `json:"metadata"`
	SchemaMetadata schema.SchemaMetadata  `json:"schema_metadata"`
}

// PlatformConfig: Registry pattern for zero duplication
type PlatformConfig struct {
	Platform      string
	AppType       string
	BasePath      string
	GeneratorFunc interface{}
	ProcessResult func(interface{}, map[string]interface{}) (ExpectedResult, error)
}

// GenerateOptions configures test data generation
type GenerateOptions struct {
	OutputDir string
	Verbose   bool
}

// DefaultGenerateOptions returns default generation options
func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		OutputDir: "tests/e2e/generated",
		Verbose:   false,
	}
}

// Generate creates test data for all registered platforms
func Generate(opts GenerateOptions) error {
	// REGISTRY: All platforms in one place (data-driven!)
	registry := []PlatformConfig{
		{
			Platform:      "google",
			AppType:       "calendar",
			BasePath:      "pkg/google/calendar",
			GeneratorFunc: googlecal.GenerateURL,
			ProcessResult: processGoogle,
		},
		{
			Platform:      "apple",
			AppType:       "calendar",
			BasePath:      "pkg/apple/calendar",
			GeneratorFunc: applecal.GenerateICS,
			ProcessResult: processApple,
		},
		// Adding Maps? Just add one line here!
	}

	// Process each using GENERIC function
	for _, config := range registry {
		suite, err := ProcessPlatform(config)
		if err != nil {
			return fmt.Errorf("%s/%s: %w", config.Platform, config.AppType, err)
		}

		outPath := filepath.Join(opts.OutputDir, fmt.Sprintf("%s-%s-tests.json", config.Platform, config.AppType))
		if err := saveSuite(suite, outPath); err != nil {
			return fmt.Errorf("save failed: %w", err)
		}

		if opts.Verbose {
			fmt.Printf("✅ %s: %d tests (%d valid) → %s\n",
				strings.Title(config.Platform),
				len(suite.TestCases),
				countValid(suite.TestCases),
				outPath)
			fmt.Printf("   Schema: %s\n", suite.SchemaMetadata.Title)
			fmt.Printf("   Required: %v\n", suite.SchemaMetadata.RequiredFields)
		}
	}

	return nil
}

// ProcessPlatform: GENERIC function (works for ALL platforms!)
// This eliminates 100+ lines of duplication
func ProcessPlatform(config PlatformConfig) (*TestSuite, error) {
	// Load examples (using shared constant from pkg/schema)
	examplesPath := filepath.Join(config.BasePath, schema.ExamplesFilename)
	examples, err := loadExamples(examplesPath)
	if err != nil {
		return nil, err
	}

	// Schema metadata (using shared constant from pkg/schema)
	schemaPath := filepath.Join(config.BasePath, schema.SchemaFilename)
	schemaMeta, err := schema.ExtractMetadata(schemaPath)
	if err != nil {
		return nil, err
	}

	// Validator
	validator := schema.NewValidatorV6()
	compiled, err := validator.LoadSchemaFromFile(schemaPath)
	if err != nil {
		return nil, err
	}

	suite := &TestSuite{
		Platform:       config.Platform,
		AppType:        config.AppType,
		GeneratedAt:    time.Now().Format(time.RFC3339),
		SourceFile:     examplesPath,
		SchemaMetadata: schemaMeta,
		Metadata: map[string]interface{}{
			"generator_signature": reflect.TypeOf(config.GeneratorFunc).String(),
			"total_examples":      len(examples.Examples),
		},
	}

	// Process examples
	for _, ex := range examples.Examples {
		tc := TestCase{
			Name:        ex.Name,
			Description: ex.Description,
			Data:        ex.Data,
			Tags:        inferTags(ex.Data),
		}

		// Validate
		tc.Validation = validator.ValidateWithDetails(ex.Data, compiled, schemaMeta.RequiredFields)

		// Generate using reflection
		result, err := callGenerator(config.GeneratorFunc, ex.Data)
		if err != nil {
			tc.Expected.Error = err.Error()
		} else {
			expected, err := config.ProcessResult(result, ex.Data)
			if err != nil {
				tc.Expected.Error = err.Error()
			} else {
				tc.Expected = expected
			}
		}

		suite.TestCases = append(suite.TestCases, tc)
	}

	return suite, nil
}

// callGenerator: Reflection-based function call (type-safe!)
func callGenerator(fn interface{}, data map[string]interface{}) (interface{}, error) {
	fnVal := reflect.ValueOf(fn)
	results := fnVal.Call([]reflect.Value{reflect.ValueOf(data)})

	if len(results) != 2 {
		return nil, fmt.Errorf("generator must return (result, error)")
	}

	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface(), nil
}

// Platform-specific processors
func processGoogle(result interface{}, data map[string]interface{}) (ExpectedResult, error) {
	url, ok := result.(string)
	if !ok {
		return ExpectedResult{}, fmt.Errorf("expected string, got %T", result)
	}
	return ExpectedResult{URL: url}, nil
}

func processApple(result interface{}, data map[string]interface{}) (ExpectedResult, error) {
	ics, ok := result.([]byte)
	if !ok {
		return ExpectedResult{}, fmt.Errorf("expected []byte, got %T", result)
	}
	return ExpectedResult{
		ICS:         string(ics),
		ICSContains: extractICSKeywords(data, string(ics)),
	}, nil
}

// Helpers
func inferTags(data map[string]interface{}) []string {
	tags := []string{}
	for _, f := range applecal.AdvancedFeatures {
		if _, ok := data[f]; ok {
			if len(tags) == 0 {
				tags = append(tags, "advanced")
			}
			tags = append(tags, f)
		}
	}
	if len(tags) == 0 {
		tags = append(tags, "basic")
	}
	return tags
}

func extractICSKeywords(data map[string]interface{}, ics string) []string {
	// Required ICS keywords from pkg/apple/calendar
	keywords := []string{
		applecal.ICSBeginCalendar,
		applecal.ICSEndCalendar,
		applecal.ICSBeginEvent,
		applecal.ICSEndEvent,
	}

	if title, ok := data[applecal.FieldTitle].(string); ok {
		keywords = append(keywords, fmt.Sprintf("%s:%s", applecal.ICSSummary, title))
	}
	if _, ok := data[applecal.FieldRecurrence]; ok {
		keywords = append(keywords, applecal.ICSRule)
	}
	if _, ok := data[applecal.FieldAttendees]; ok {
		keywords = append(keywords, applecal.ICSAttendee)
	}

	found := []string{}
	for _, kw := range keywords {
		if strings.Contains(ics, kw) {
			found = append(found, kw)
		}
	}
	return found
}

func loadExamples(path string) (*ExamplesFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var examples ExamplesFile
	if err := json.Unmarshal(data, &examples); err != nil {
		return nil, err
	}
	return &examples, nil
}

func countValid(testCases []TestCase) int {
	count := 0
	for _, tc := range testCases {
		if tc.Validation.IsValid {
			count++
		}
	}
	return count
}

func saveSuite(suite *TestSuite, path string) error {
	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
