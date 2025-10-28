package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// ==================== BACKWARDS COMPATIBILITY ====================
// These functions maintain compatibility with existing server code that
// uses the old custom validator. Eventually, the server should be refactored
// to use ValidatorV6 directly instead of these legacy functions.

// ValidateAgainstSchema validates form data against a JSON Schema
// This is a backwards-compatible shim for existing server code.
//
// DEPRECATED: New code should use ValidatorV6 directly.
func ValidateAgainstSchema(data map[string]interface{}, customSchema *JSONSchema) ValidationErrors {
	// Convert the custom JSONSchema struct to actual JSON
	schemaBytes, err := json.Marshal(customSchema)
	if err != nil {
		return ValidationErrors{
			"_error": fmt.Sprintf("Failed to marshal schema: %v", err),
		}
	}

	// Create a new compiler and compile the schema
	compiler := jsonschema.NewCompiler()

	// Unmarshal the schema JSON into interface{} for the compiler
	var schemaDoc interface{}
	if err := json.Unmarshal(schemaBytes, &schemaDoc); err != nil {
		return ValidationErrors{
			"_error": fmt.Sprintf("Failed to unmarshal schema: %v", err),
		}
	}

	// Add the schema to compiler using the correct method
	if err := compiler.AddResource("schema.json", schemaDoc); err != nil {
		return ValidationErrors{
			"_error": fmt.Sprintf("Failed to add schema resource: %v", err),
		}
	}

	// Compile the schema
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return ValidationErrors{
			"_error": fmt.Sprintf("Failed to compile schema: %v", err),
		}
	}

	// Validate the data
	validationErr := schema.Validate(data)
	if validationErr == nil {
		return ValidationErrors{} // No errors
	}

	// Convert validation errors
	if valErr, ok := validationErr.(*jsonschema.ValidationError); ok {
		return convertValidationErrorV6(valErr)
	}

	return ValidationErrors{
		"_error": validationErr.Error(),
	}
}

// FormDataToMap converts HTML form data to a map suitable for validation
func FormDataToMap(formData map[string][]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, values := range formData {
		if len(values) == 0 {
			continue
		}

		value := values[0] // Take first value for now

		// Handle nested objects with dot notation (e.g., "organizer.name")
		if strings.Contains(key, ".") {
			parts := strings.Split(key, ".")
			setNestedValue(result, parts, value)
		} else {
			result[key] = value
		}
	}

	return result
}

// setNestedValue sets a value in a nested map using a path
func setNestedValue(m map[string]interface{}, path []string, value string) {
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
