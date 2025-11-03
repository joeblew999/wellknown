package schema

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// SchemaMetadata represents extracted schema information
// This is used for test generation and documentation
type SchemaMetadata struct {
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	RequiredFields []string          `json:"required_fields"`
	OptionalFields []string          `json:"optional_fields"`
	FieldTypes     map[string]string `json:"field_types"`
}

// ValidationResult contains detailed validation results
type ValidationResult struct {
	IsValid       bool     `json:"is_valid"`
	Errors        []string `json:"errors,omitempty"`
	MissingFields []string `json:"missing_required_fields,omitempty"`
}

// ExtractMetadata extracts metadata from a JSON Schema file
// This uses reflection-style analysis of the schema structure
func ExtractMetadata(schemaPath string) (SchemaMetadata, error) {
	// Read raw schema file
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return SchemaMetadata{}, fmt.Errorf("failed to read schema: %w", err)
	}

	var rawSchema map[string]interface{}
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return SchemaMetadata{}, fmt.Errorf("failed to parse schema: %w", err)
	}

	metadata := SchemaMetadata{
		FieldTypes: make(map[string]string),
	}

	// Extract title and description
	if title, ok := rawSchema["title"].(string); ok {
		metadata.Title = title
	}
	if desc, ok := rawSchema["description"].(string); ok {
		metadata.Description = desc
	}

	// Extract required fields
	if required, ok := rawSchema["required"].([]interface{}); ok {
		for _, field := range required {
			if fieldName, ok := field.(string); ok {
				metadata.RequiredFields = append(metadata.RequiredFields, fieldName)
			}
		}
	}

	// Extract properties and field types
	if properties, ok := rawSchema["properties"].(map[string]interface{}); ok {
		for fieldName, fieldDef := range properties {
			if def, ok := fieldDef.(map[string]interface{}); ok {
				// Get field type
				if fieldType, ok := def["type"].(string); ok {
					metadata.FieldTypes[fieldName] = fieldType
				}

				// Determine if optional
				isRequired := false
				for _, req := range metadata.RequiredFields {
					if req == fieldName {
						isRequired = true
						break
					}
				}
				if !isRequired {
					metadata.OptionalFields = append(metadata.OptionalFields, fieldName)
				}
			}
		}
	}

	return metadata, nil
}

// ValidateWithDetails validates data against a schema and returns detailed results
// This is used by test generators to capture validation status
func (v *ValidatorV6) ValidateWithDetails(data map[string]interface{}, schema *jsonschema.Schema, requiredFields []string) ValidationResult {
	result := ValidationResult{IsValid: true}

	// Perform validation using jsonschema
	err := schema.Validate(data)
	if err != nil {
		result.IsValid = false

		// Extract detailed error messages
		if valErr, ok := err.(*jsonschema.ValidationError); ok {
			for _, e := range flattenValidationErrors(valErr) {
				result.Errors = append(result.Errors, e.Error())
			}
		} else {
			result.Errors = []string{err.Error()}
		}
	}

	// Check for missing required fields
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			result.MissingFields = append(result.MissingFields, field)
		}
	}

	return result
}

// flattenValidationErrors recursively flattens validation error tree
func flattenValidationErrors(ve *jsonschema.ValidationError) []*jsonschema.ValidationError {
	var errors []*jsonschema.ValidationError
	errors = append(errors, ve)

	for _, cause := range ve.Causes {
		errors = append(errors, flattenValidationErrors(cause)...)
	}

	return errors
}
