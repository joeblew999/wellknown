package schema

import (
	"encoding/json"
	"fmt"
	"html/template"
)

// JSONSchema represents a simplified JSON Schema for form generation
type JSONSchema struct {
	Schema      string                 `json:"$schema"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Properties  map[string]Property    `json:"properties"`
	Required    []string               `json:"required"`
}

// Property represents a schema property
type Property struct {
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Format      string      `json:"format,omitempty"`
	MinLength   int         `json:"minLength,omitempty"`
	MaxLength   int         `json:"maxLength,omitempty"`
	Minimum     int         `json:"minimum,omitempty"`
	Maximum     int         `json:"maximum,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Examples    []string    `json:"examples,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Items       *Property   `json:"items,omitempty"` // For arrays
	Properties  map[string]Property `json:"properties,omitempty"` // For nested objects
}

// ParseSchema parses a JSON Schema string into a JSONSchema struct
func ParseSchema(schemaJSON string) (*JSONSchema, error) {
	var schema JSONSchema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}
	return &schema, nil
}

// GenerateFormHTML generates HTML form fields from a JSON Schema
func (s *JSONSchema) GenerateFormHTML() template.HTML {
	var html string

	for fieldName, prop := range s.Properties {
		isRequired := contains(s.Required, fieldName)
		html += s.generateFieldHTML(fieldName, prop, isRequired, 0)
	}

	return template.HTML(html)
}

// generateFieldHTML generates HTML for a single field
func (s *JSONSchema) generateFieldHTML(fieldName string, prop Property, required bool, depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	requiredAttr := ""
	if required {
		requiredAttr = " required"
	}

	placeholder := ""
	if len(prop.Examples) > 0 {
		placeholder = fmt.Sprintf(` placeholder="%s"`, prop.Examples[0])
	}

	title := prop.Title
	if title == "" {
		title = fieldName
	}

	var html string

	// Start form group
	html += fmt.Sprintf(`%s<div class="form-group">
%s  <label for="%s">%s`, indent, indent, fieldName, title)

	if required {
		html += " *"
	}
	html += "</label>\n"

	if prop.Description != "" {
		html += fmt.Sprintf(`%s  <p class="field-description">%s</p>
`, indent, prop.Description)
	}

	// Generate input based on type
	switch prop.Type {
	case "string":
		html += s.generateStringInput(fieldName, prop, requiredAttr, placeholder, indent)
	case "boolean":
		html += s.generateBooleanInput(fieldName, prop, indent)
	case "integer", "number":
		html += s.generateNumberInput(fieldName, prop, requiredAttr, indent)
	case "array":
		html += s.generateArrayInput(fieldName, prop, indent)
	case "object":
		html += s.generateObjectInput(fieldName, prop, indent)
	default:
		html += fmt.Sprintf(`%s  <input type="text" id="%s" name="%s"%s%s>
`, indent, fieldName, fieldName, requiredAttr, placeholder)
	}

	// End form group
	html += fmt.Sprintf("%s</div>\n\n", indent)

	return html
}

func (s *JSONSchema) generateStringInput(fieldName string, prop Property, requiredAttr, placeholder, indent string) string {
	switch prop.Format {
	case "datetime-local":
		defaultValue := ""
		if prop.Default != nil {
			defaultValue = fmt.Sprintf(` value="%v"`, prop.Default)
		}
		return fmt.Sprintf(`%s  <input type="datetime-local" id="%s" name="%s"%s%s>
`, indent, fieldName, fieldName, requiredAttr, defaultValue)
	case "date":
		return fmt.Sprintf(`%s  <input type="date" id="%s" name="%s"%s>
`, indent, fieldName, fieldName, requiredAttr)
	case "email":
		return fmt.Sprintf(`%s  <input type="email" id="%s" name="%s"%s%s>
`, indent, fieldName, fieldName, requiredAttr, placeholder)
	case "uri", "url":
		return fmt.Sprintf(`%s  <input type="url" id="%s" name="%s"%s%s>
`, indent, fieldName, fieldName, requiredAttr, placeholder)
	default:
		if len(prop.Enum) > 0 {
			// Generate select dropdown for enum
			html := fmt.Sprintf(`%s  <select id="%s" name="%s"%s>
`, indent, fieldName, fieldName, requiredAttr)
			html += fmt.Sprintf(`%s    <option value="">-- Select --</option>
`, indent)
			for _, option := range prop.Enum {
				html += fmt.Sprintf(`%s    <option value="%s">%s</option>
`, indent, option, option)
			}
			html += fmt.Sprintf(`%s  </select>
`, indent)
			return html
		}

		maxlength := ""
		if prop.MaxLength > 0 {
			maxlength = fmt.Sprintf(` maxlength="%d"`, prop.MaxLength)
		}

		// Check if it should be textarea (long text)
		if prop.MaxLength > 200 {
			return fmt.Sprintf(`%s  <textarea id="%s" name="%s"%s%s%s></textarea>
`, indent, fieldName, fieldName, requiredAttr, maxlength, placeholder)
		}

		return fmt.Sprintf(`%s  <input type="text" id="%s" name="%s"%s%s%s>
`, indent, fieldName, fieldName, requiredAttr, maxlength, placeholder)
	}
}

func (s *JSONSchema) generateBooleanInput(fieldName string, prop Property, indent string) string {
	checked := ""
	if prop.Default != nil && prop.Default == true {
		checked = " checked"
	}
	return fmt.Sprintf(`%s  <input type="checkbox" id="%s" name="%s" value="true"%s>
`, indent, fieldName, fieldName, checked)
}

func (s *JSONSchema) generateNumberInput(fieldName string, prop Property, requiredAttr, indent string) string {
	min := ""
	max := ""
	if prop.Minimum > 0 {
		min = fmt.Sprintf(` min="%d"`, prop.Minimum)
	}
	if prop.Maximum > 0 {
		max = fmt.Sprintf(` max="%d"`, prop.Maximum)
	}

	return fmt.Sprintf(`%s  <input type="number" id="%s" name="%s"%s%s%s>
`, indent, fieldName, fieldName, requiredAttr, min, max)
}

func (s *JSONSchema) generateArrayInput(fieldName string, prop Property, indent string) string {
	// For now, show a note that array inputs are not yet implemented
	return fmt.Sprintf(`%s  <div class="array-input-placeholder">
%s    <p><em>Array input for "%s" - Coming soon</em></p>
%s  </div>
`, indent, indent, prop.Title, indent)
}

func (s *JSONSchema) generateObjectInput(fieldName string, prop Property, indent string) string {
	// For now, show a note that nested objects are not yet implemented
	return fmt.Sprintf(`%s  <div class="object-input-placeholder">
%s    <p><em>Nested object input for "%s" - Coming soon</em></p>
%s  </div>
`, indent, indent, prop.Title, indent)
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
