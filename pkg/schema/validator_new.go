package schema

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

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
