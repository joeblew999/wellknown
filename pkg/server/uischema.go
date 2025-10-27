package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

// UISchema represents the UI layout configuration for a form
// Inspired by JSON Forms (jsonforms.io) and goPocJsonSchemaForm
type UISchema struct {
	Type     string    `json:"type"`
	Elements []Element `json:"elements,omitempty"`
}

// Element is a discriminated union of different UI element types
type Element struct {
	Type string `json:"type"`

	// For Layout types (VerticalLayout, HorizontalLayout)
	Elements []Element `json:"elements,omitempty"`

	// For Control type
	Scope       string   `json:"scope,omitempty"`       // JSON pointer to schema property (e.g., "#/properties/title")
	Label       string   `json:"label,omitempty"`       // Override label from schema
	Description string   `json:"description,omitempty"` // Override description from schema
	Options     *Options `json:"options,omitempty"`     // Control options

	// For Label type
	Text string `json:"text,omitempty"`

	// For Group type
	Title string `json:"title,omitempty"`
}

// Options for control rendering
type Options struct {
	Placeholder string `json:"placeholder,omitempty"`
	Multi       bool   `json:"multi,omitempty"`     // For multi-line text
	Format      string `json:"format,omitempty"`    // Override format
	ShowLabel   *bool  `json:"showLabel,omitempty"` // Show/hide label
}

// ParseUISchema parses a UI Schema JSON string
func ParseUISchema(uiSchemaJSON string) (*UISchema, error) {
	var uiSchema UISchema
	if err := json.Unmarshal([]byte(uiSchemaJSON), &uiSchema); err != nil {
		return nil, fmt.Errorf("failed to parse UI schema: %w", err)
	}
	return &uiSchema, nil
}

// GenerateFormHTML generates HTML from UI Schema and JSON Schema
func (u *UISchema) GenerateFormHTML(jsonSchema *JSONSchema) template.HTML {
	var html strings.Builder
	html.WriteString(`<div class="ui-schema-form">` + "\n")
	u.renderElement(Element{Type: u.Type, Elements: u.Elements}, jsonSchema, &html, 0)
	html.WriteString("</div>\n")
	return template.HTML(html.String())
}

// renderElement renders a single UI element
func (u *UISchema) renderElement(elem Element, jsonSchema *JSONSchema, html *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)

	switch elem.Type {
	case "VerticalLayout":
		html.WriteString(indent + `<div class="vertical-layout">` + "\n")
		for _, child := range elem.Elements {
			u.renderElement(child, jsonSchema, html, depth+1)
		}
		html.WriteString(indent + "</div>\n")

	case "HorizontalLayout":
		html.WriteString(indent + `<div class="horizontal-layout">` + "\n")
		for _, child := range elem.Elements {
			html.WriteString(indent + `  <div class="horizontal-item">` + "\n")
			u.renderElement(child, jsonSchema, html, depth+2)
			html.WriteString(indent + `  </div>` + "\n")
		}
		html.WriteString(indent + "</div>\n")

	case "Group":
		html.WriteString(indent + `<fieldset class="form-group-section">` + "\n")
		if elem.Title != "" {
			html.WriteString(indent + `  <legend>` + elem.Title + `</legend>` + "\n")
		}
		for _, child := range elem.Elements {
			u.renderElement(child, jsonSchema, html, depth+1)
		}
		html.WriteString(indent + "</fieldset>\n")

	case "Control":
		u.renderControl(elem, jsonSchema, html, depth)

	case "Label":
		if elem.Text != "" {
			html.WriteString(indent + `<h4 class="ui-label">` + elem.Text + `</h4>` + "\n")
		}

	default:
		html.WriteString(indent + fmt.Sprintf("<!-- Unknown element type: %s -->\n", elem.Type))
	}
}

// renderControl renders a form control based on schema
func (u *UISchema) renderControl(elem Element, jsonSchema *JSONSchema, html *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)

	// Parse scope to get field name (e.g., "#/properties/title" -> "title")
	fieldName := u.parseScopeToFieldName(elem.Scope)
	if fieldName == "" {
		html.WriteString(indent + fmt.Sprintf("<!-- Invalid scope: %s -->\n", elem.Scope))
		return
	}

	// Get property from JSON Schema
	prop, exists := jsonSchema.Properties[fieldName]
	if !exists {
		html.WriteString(indent + fmt.Sprintf("<!-- Property not found in schema: %s -->\n", fieldName))
		return
	}

	// Check if required
	isRequired := contains(jsonSchema.Required, fieldName)

	// Use label from UI Schema or fall back to JSON Schema
	label := elem.Label
	if label == "" {
		label = prop.Title
		if label == "" {
			label = fieldName
		}
	}

	// Use description from UI Schema or fall back to JSON Schema
	description := elem.Description
	if description == "" {
		description = prop.Description
	}

	// Start form group
	html.WriteString(indent + `<div class="form-group">` + "\n")

	// Render label (unless explicitly hidden)
	showLabel := true
	if elem.Options != nil && elem.Options.ShowLabel != nil {
		showLabel = *elem.Options.ShowLabel
	}
	if showLabel {
		html.WriteString(indent + `  <label for="` + fieldName + `">` + label)
		if isRequired {
			html.WriteString(" *")
		}
		html.WriteString("</label>\n")
	}

	// Render description
	if description != "" {
		html.WriteString(indent + `  <p class="field-description">` + description + `</p>` + "\n")
	}

	// Render input based on type
	u.renderInput(elem, fieldName, prop, isRequired, html, indent)

	// End form group
	html.WriteString(indent + "</div>\n")
}

// renderInput renders the actual input element
func (u *UISchema) renderInput(elem Element, fieldName string, prop Property, required bool, html *strings.Builder, indent string) {
	requiredAttr := ""
	if required {
		requiredAttr = " required"
	}

	placeholder := ""
	if elem.Options != nil && elem.Options.Placeholder != "" {
		placeholder = fmt.Sprintf(` placeholder="%s"`, elem.Options.Placeholder)
	} else if len(prop.Examples) > 0 {
		placeholder = fmt.Sprintf(` placeholder="%s"`, prop.Examples[0])
	}

	// Determine format (UI Schema overrides JSON Schema)
	format := prop.Format
	if elem.Options != nil && elem.Options.Format != "" {
		format = elem.Options.Format
	}

	switch prop.Type {
	case "string":
		switch format {
		case "datetime-local":
			html.WriteString(indent + `  <input type="datetime-local" id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + `>` + "\n")
		case "date":
			html.WriteString(indent + `  <input type="date" id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + `>` + "\n")
		case "email":
			html.WriteString(indent + `  <input type="email" id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + placeholder + `>` + "\n")
		default:
			// Check for enum (select dropdown)
			if len(prop.Enum) > 0 {
				html.WriteString(indent + `  <select id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + `>` + "\n")
				html.WriteString(indent + `    <option value="">-- Select --</option>` + "\n")
				for _, option := range prop.Enum {
					html.WriteString(indent + `    <option value="` + option + `">` + option + `</option>` + "\n")
				}
				html.WriteString(indent + `  </select>` + "\n")
			} else if elem.Options != nil && elem.Options.Multi || prop.MaxLength > 200 {
				// Textarea for multi-line or long text
				maxlength := ""
				if prop.MaxLength > 0 {
					maxlength = fmt.Sprintf(` maxlength="%d"`, prop.MaxLength)
				}
				html.WriteString(indent + `  <textarea id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + maxlength + placeholder + `></textarea>` + "\n")
			} else {
				// Regular text input
				maxlength := ""
				if prop.MaxLength > 0 {
					maxlength = fmt.Sprintf(` maxlength="%d"`, prop.MaxLength)
				}
				html.WriteString(indent + `  <input type="text" id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + maxlength + placeholder + `>` + "\n")
			}
		}

	case "boolean":
		checked := ""
		if prop.Default != nil && prop.Default == true {
			checked = " checked"
		}
		html.WriteString(indent + `  <input type="checkbox" id="` + fieldName + `" name="` + fieldName + `" value="true"` + checked + `>` + "\n")

	case "integer", "number":
		min := ""
		max := ""
		if prop.Minimum > 0 {
			min = fmt.Sprintf(` min="%d"`, prop.Minimum)
		}
		if prop.Maximum > 0 {
			max = fmt.Sprintf(` max="%d"`, prop.Maximum)
		}
		html.WriteString(indent + `  <input type="number" id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + min + max + `>` + "\n")

	case "array":
		html.WriteString(indent + `  <div class="array-input-placeholder">` + "\n")
		html.WriteString(indent + `    <p><em>Array input - Coming soon</em></p>` + "\n")
		html.WriteString(indent + `  </div>` + "\n")

	case "object":
		html.WriteString(indent + `  <div class="object-input-placeholder">` + "\n")
		html.WriteString(indent + `    <p><em>Nested object - Coming soon</em></p>` + "\n")
		html.WriteString(indent + `  </div>` + "\n")
	}
}

// parseScopeToFieldName extracts field name from JSON pointer scope
// e.g., "#/properties/title" -> "title"
func (u *UISchema) parseScopeToFieldName(scope string) string {
	// Remove leading "#/"
	scope = strings.TrimPrefix(scope, "#/")

	// Split by "/" and get last part
	parts := strings.Split(scope, "/")
	if len(parts) == 0 {
		return ""
	}

	// For "#/properties/fieldname", return "fieldname"
	// For "#/fieldname", return "fieldname"
	if len(parts) >= 2 && parts[0] == "properties" {
		return parts[1]
	}

	return parts[len(parts)-1]
}
