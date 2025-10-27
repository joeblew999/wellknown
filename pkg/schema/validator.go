package schema

import (
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ValidationErrors maps field paths to error messages
type ValidationErrors map[string]string

// ValidateAgainstSchema validates form data against a JSON Schema
// Returns a map of field path -> error message
// Uses only stdlib - no external validation libraries!
func ValidateAgainstSchema(data map[string]interface{}, schema *JSONSchema) ValidationErrors {
	errors := make(ValidationErrors)

	// Check required fields
	for _, requiredField := range schema.Required {
		if _, exists := data[requiredField]; !exists {
			errors[requiredField] = "This field is required"
			continue
		}
		// Check if the value is empty string
		if str, ok := data[requiredField].(string); ok && strings.TrimSpace(str) == "" {
			errors[requiredField] = "This field is required"
		}
	}

	// Validate each property
	for fieldName, value := range data {
		prop, exists := schema.Properties[fieldName]
		if !exists {
			continue // Unknown field, skip
		}

		// Skip validation for empty optional fields
		if !contains(schema.Required, fieldName) {
			if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
				continue
			}
		}

		// Validate based on type
		if err := validateProperty(fieldName, value, prop); err != nil {
			errors[fieldName] = err.Error()
		}
	}

	return errors
}

// validateProperty validates a single property value against its schema
func validateProperty(fieldName string, value interface{}, prop Property) error {
	switch prop.Type {
	case "string":
		return validateString(fieldName, value, prop)
	case "integer", "number":
		return validateNumber(fieldName, value, prop)
	case "boolean":
		return validateBoolean(fieldName, value, prop)
	case "array":
		return validateArray(fieldName, value, prop)
	case "object":
		return validateObject(fieldName, value, prop)
	default:
		return nil
	}
}

// validateString validates string values
func validateString(fieldName string, value interface{}, prop Property) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("must be a string")
	}

	// Check minLength
	if prop.MinLength > 0 && len(str) < prop.MinLength {
		return fmt.Errorf("must be at least %d characters", prop.MinLength)
	}

	// Check maxLength
	if prop.MaxLength > 0 && len(str) > prop.MaxLength {
		return fmt.Errorf("must be at most %d characters", prop.MaxLength)
	}

	// Check enum
	if len(prop.Enum) > 0 {
		valid := false
		for _, enumVal := range prop.Enum {
			if str == enumVal {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("must be one of: %s", strings.Join(prop.Enum, ", "))
		}
	}

	// Check format
	if prop.Format != "" {
		if err := validateFormat(str, prop.Format); err != nil {
			return err
		}
	}

	return nil
}

// validateFormat validates string formats
func validateFormat(value string, format string) error {
	switch format {
	case "email":
		if _, err := mail.ParseAddress(value); err != nil {
			return fmt.Errorf("must be a valid email address")
		}
	case "uri", "url":
		if _, err := url.ParseRequestURI(value); err != nil {
			return fmt.Errorf("must be a valid URL")
		}
	case "date":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("must be a valid date (YYYY-MM-DD)")
		}
	case "datetime-local":
		// HTML datetime-local format: YYYY-MM-DDTHH:MM
		if _, err := time.Parse("2006-01-02T15:04", value); err != nil {
			return fmt.Errorf("must be a valid datetime")
		}
	case "time":
		if _, err := time.Parse("15:04", value); err != nil {
			return fmt.Errorf("must be a valid time (HH:MM)")
		}
	}
	return nil
}

// validateNumber validates integer and number values
func validateNumber(fieldName string, value interface{}, prop Property) error {
	var num float64

	// Try to parse as number
	switch v := value.(type) {
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		num = parsed
	case float64:
		num = v
	case int:
		num = float64(v)
	default:
		return fmt.Errorf("must be a number")
	}

	// For integer type, check if it's actually an integer
	if prop.Type == "integer" {
		if num != float64(int(num)) {
			return fmt.Errorf("must be an integer")
		}
	}

	// Check minimum
	if prop.Minimum > 0 && int(num) < prop.Minimum {
		return fmt.Errorf("must be at least %d", prop.Minimum)
	}

	// Check maximum
	if prop.Maximum > 0 && int(num) > prop.Maximum {
		return fmt.Errorf("must be at most %d", prop.Maximum)
	}

	return nil
}

// validateBoolean validates boolean values
func validateBoolean(fieldName string, value interface{}, prop Property) error {
	switch v := value.(type) {
	case bool:
		return nil
	case string:
		if v == "true" || v == "false" || v == "on" || v == "" {
			return nil
		}
		return fmt.Errorf("must be true or false")
	default:
		return fmt.Errorf("must be a boolean")
	}
}

// validateArray validates array values
func validateArray(fieldName string, value interface{}, prop Property) error {
	// Arrays from forms need special handling
	// For now, just check if it's a slice
	// TODO: Implement full array validation
	return nil
}

// validateObject validates nested object values
func validateObject(fieldName string, value interface{}, prop Property) error {
	// Objects from forms need special handling (dot notation)
	// TODO: Implement full object validation
	return nil
}

// FormDataToMap converts form values to a map structure
// Handles both simple fields and nested objects (dot notation)
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

// setNestedValue sets a value in a nested map structure
// e.g., ["organizer", "name"] with value "John" creates map["organizer"]["name"] = "John"
func setNestedValue(m map[string]interface{}, path []string, value interface{}) {
	if len(path) == 0 {
		return
	}

	if len(path) == 1 {
		m[path[0]] = value
		return
	}

	// Create nested map if it doesn't exist
	if _, exists := m[path[0]]; !exists {
		m[path[0]] = make(map[string]interface{})
	}

	// Recurse into nested map
	if nested, ok := m[path[0]].(map[string]interface{}); ok {
		setNestedValue(nested, path[1:], value)
	}
}

// GetNestedValue gets a value from a nested map using dot notation
// e.g., "organizer.name" returns map["organizer"]["name"]
func GetNestedValue(m map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	return getNestedValueHelper(m, parts)
}

func getNestedValueHelper(m map[string]interface{}, path []string) (interface{}, bool) {
	if len(path) == 0 {
		return nil, false
	}

	if len(path) == 1 {
		val, exists := m[path[0]]
		return val, exists
	}

	if nested, ok := m[path[0]].(map[string]interface{}); ok {
		return getNestedValueHelper(nested, path[1:])
	}

	return nil, false
}
