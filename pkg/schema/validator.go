package schema

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ValidationErrors maps field names to error messages
type ValidationErrors map[string]string

// ValidatorV6 handles JSON Schema validation using santhosh-tekuri/jsonschema v6
type ValidatorV6 struct {
	compiler *jsonschema.Compiler
	schemas  map[string]*jsonschema.Schema
}

// NewValidatorV6 creates a new validator instance using jsonschema v6
func NewValidatorV6() *ValidatorV6 {
	compiler := jsonschema.NewCompiler()

	return &ValidatorV6{
		compiler: compiler,
		schemas:  make(map[string]*jsonschema.Schema),
	}
}

// LoadSchemaFromFile loads and compiles a JSON Schema file
func (v *ValidatorV6) LoadSchemaFromFile(schemaPath string) (*jsonschema.Schema, error) {
	// Check cache first
	if cached, ok := v.schemas[schemaPath]; ok {
		return cached, nil
	}

	// Read file
	absPath, err := filepath.Abs(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %s: %w", schemaPath, err)
	}

	// Check file exists
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("schema file not found: %s", absPath)
	}

	// Compile schema (library handles file loading)
	fileURL := "file:///" + absPath
	schema, err := v.compiler.Compile(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema %s: %w", schemaPath, err)
	}

	// Cache it
	v.schemas[schemaPath] = schema
	return schema, nil
}

// Validate validates data against a compiled schema
// Returns ValidationErrors for backwards compatibility
func (v *ValidatorV6) Validate(data map[string]interface{}, schema *jsonschema.Schema) ValidationErrors {
	errors := make(ValidationErrors)

	// Use library's validation
	err := schema.Validate(data)
	if err == nil {
		return errors // No errors
	}

	// Convert validation errors to our format
	if valErr, ok := err.(*jsonschema.ValidationError); ok {
		errors = convertValidationErrorV6(valErr)
	} else {
		// Generic error
		errors["_error"] = err.Error()
	}

	return errors
}

// convertValidationErrorV6 converts jsonschema.ValidationError to our format
func convertValidationErrorV6(err *jsonschema.ValidationError) ValidationErrors {
	errors := make(ValidationErrors)

	// Get the instance path (which field failed)
	// InstanceLocation is []string like ["fieldName"] or ["fieldName", "subField"]
	instancePath := err.InstanceLocation

	// Convert to simple field name
	fieldName := strings.Join(instancePath, "/")
	if fieldName == "" {
		fieldName = "_root"
	}

	// Get the error message from ErrorKind
	message := fmt.Sprintf("%v", err.ErrorKind)

	// Store error
	errors[fieldName] = message

	// Also add any sub-errors
	for _, cause := range err.Causes {
		subErrors := convertValidationErrorV6(cause)
		for k, v := range subErrors {
			errors[k] = v
		}
	}

	return errors
}

// ValidateWithContext validates data with context support
func (v *ValidatorV6) ValidateWithContext(ctx context.Context, data map[string]interface{}, schema *jsonschema.Schema) ValidationErrors {
	errors := make(ValidationErrors)

	err := schema.Validate(data)
	if err == nil {
		return errors
	}

	if valErr, ok := err.(*jsonschema.ValidationError); ok {
		errors = convertValidationErrorV6(valErr)
	} else {
		errors["_error"] = err.Error()
	}

	return errors
}

// FormDataToMap converts HTML form data to a map suitable for validation
// Supports arrays (field[0]), nested objects (field.subfield), and array of objects (field[0].subfield)
func FormDataToMap(formData map[string][]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, values := range formData {
		if len(values) == 0 {
			continue
		}

		value := values[0] // Take first value

		// Type coercion: HTML forms send everything as strings
		typedValue := coerceType(value)

		// Parse the key to handle arrays and nested objects
		setValueByPath(result, key, typedValue)
	}

	return result
}

// coerceType converts string values to appropriate types
func coerceType(value string) interface{} {
	// Convert boolean strings
	if value == "true" {
		return true
	} else if value == "false" {
		return false
	}

	// Try to convert to number - always return float64 like JSON unmarshaling does
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num
	}

	// Return as string
	return value
}

// setValueByPath sets a value in a nested structure based on a path
// Supports: "field", "field.nested", "field[0]", "field[0].nested"
func setValueByPath(m map[string]interface{}, path string, value interface{}) {
	// Check for array notation: field[index] or field[index].nested
	arrayRegex := regexp.MustCompile(`^([^\[]+)\[(\d+)\](.*)$`)
	if matches := arrayRegex.FindStringSubmatch(path); matches != nil {
		fieldName := matches[1]
		index, _ := strconv.Atoi(matches[2])
		remainder := strings.TrimPrefix(matches[3], ".")

		// Ensure the field is an array
		if _, exists := m[fieldName]; !exists {
			m[fieldName] = make([]interface{}, 0)
		}

		arr, ok := m[fieldName].([]interface{})
		if !ok {
			arr = make([]interface{}, 0)
		}

		// Expand array if necessary
		for len(arr) <= index {
			arr = append(arr, nil)
		}

		if remainder == "" {
			// Simple array element: field[0]
			arr[index] = value
		} else {
			// Array of objects: field[0].nested
			if arr[index] == nil {
				arr[index] = make(map[string]interface{})
			}
			obj, ok := arr[index].(map[string]interface{})
			if !ok {
				obj = make(map[string]interface{})
				arr[index] = obj
			}
			setValueByPath(obj, remainder, value)
		}

		m[fieldName] = arr
		return
	}

	// Check for dot notation: field.nested
	if strings.Contains(path, ".") {
		parts := strings.Split(path, ".")
		setNestedValue(m, parts, value)
		return
	}

	// Simple field
	m[path] = value
}

// setNestedValue sets a value in a nested map using a path
func setNestedValue(m map[string]interface{}, path []string, value interface{}) {
	if len(path) == 0 {
		return
	}

	if len(path) == 1 {
		m[path[0]] = value
		return
	}

	// Create or get nested map
	key := path[0]
	nested, ok := m[key].(map[string]interface{})
	if !ok {
		nested = make(map[string]interface{})
		m[key] = nested
	}

	setNestedValue(nested, path[1:], value)
}

// ============================================================================
// Schema Loading and Caching
// ============================================================================

// uiSchemaCache caches loaded UI schemas (JSON strings) in memory
var uiSchemaCache = struct {
	sync.RWMutex
	schemas map[string]string
}{
	schemas: make(map[string]string),
}

// LoadUISchemaFromFile loads a UI Schema JSON file with caching
func LoadUISchemaFromFile(platform, appType string) (string, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("%s/%s/uischema", platform, appType)

	// Check cache first (read lock)
	uiSchemaCache.RLock()
	if cached, exists := uiSchemaCache.schemas[cacheKey]; exists {
		uiSchemaCache.RUnlock()
		return cached, nil
	}
	uiSchemaCache.RUnlock()

	// Not in cache, load from file (write lock)
	uiSchemaCache.Lock()
	defer uiSchemaCache.Unlock()

	// Double-check cache (another goroutine might have loaded it)
	if cached, exists := uiSchemaCache.schemas[cacheKey]; exists {
		return cached, nil
	}

	// Try relative path from project root first
	path := fmt.Sprintf("pkg/%s/%s/%s", platform, appType, UISchemaFilename)
	content, err := os.ReadFile(path)
	if err != nil {
		// If that fails, try from cmd/server/ directory (Air case)
		path = fmt.Sprintf("../../pkg/%s/%s/%s", platform, appType, UISchemaFilename)
		content, err = os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read UI schema file: %w", err)
		}
	}

	// Store in cache
	uiSchemaStr := string(content)
	uiSchemaCache.schemas[cacheKey] = uiSchemaStr

	return uiSchemaStr, nil
}

// LoadSchemasForRendering loads everything needed for form rendering and validation
// Returns: (uiSchemaJSON, compiledSchema, validator, error)
// This is the SINGLE function that should be called to get schemas for a platform/appType
func LoadSchemasForRendering(platform, appType string) (string, *jsonschema.Schema, *ValidatorV6, error) {
	// Load UI Schema JSON (cached)
	uiSchemaJSON, err := LoadUISchemaFromFile(platform, appType)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to load UI schema: %w", err)
	}

	// Create validator and load compiled schema (cached internally by validator)
	validator := NewValidatorV6()

	// Try multiple paths (project root, then from cmd/server, then from pkg/server for tests)
	schemaPaths := []string{
		fmt.Sprintf("pkg/%s/%s/%s", platform, appType, SchemaFilename),           // From project root
		fmt.Sprintf("../../pkg/%s/%s/%s", platform, appType, SchemaFilename),     // From cmd/server
		fmt.Sprintf("../%s/%s/%s", platform, appType, SchemaFilename),            // From pkg/server (tests)
	}

	var compiledSchema *jsonschema.Schema
	var lastErr error
	for _, schemaPath := range schemaPaths {
		compiledSchema, err = validator.LoadSchemaFromFile(schemaPath)
		if err == nil {
			return uiSchemaJSON, compiledSchema, validator, nil
		}
		lastErr = err
	}

	return "", nil, nil, fmt.Errorf("failed to compile schema (tried %d paths): %w", len(schemaPaths), lastErr)
}
