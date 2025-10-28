package schema

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
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
	Placeholder string   `json:"placeholder,omitempty"`
	Multi       bool     `json:"multi,omitempty"`     // For multi-line text
	Format      string   `json:"format,omitempty"`    // Override format
	ShowLabel   *bool    `json:"showLabel,omitempty"` // Show/hide label
	Suggestions []string `json:"suggestions,omitempty"` // Autocomplete suggestions
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
func (u *UISchema) GenerateFormHTML(jsonSchema *jsonschema.Schema) template.HTML {
	return u.GenerateFormHTMLWithData(jsonSchema, nil, nil)
}

// GenerateFormHTMLWithData generates HTML with validation errors and pre-filled data
func (u *UISchema) GenerateFormHTMLWithData(jsonSchema *jsonschema.Schema, formData map[string]interface{}, validationErrors ValidationErrors) template.HTML {
	var html strings.Builder
	html.WriteString(`<div class="ui-schema-form">` + "\n")
	u.renderElementWithData(Element{Type: u.Type, Elements: u.Elements}, jsonSchema, formData, validationErrors, &html, 0)
	html.WriteString("</div>\n")
	return template.HTML(html.String())
}

// renderElementWithData renders a single UI element with validation errors and form data
func (u *UISchema) renderElementWithData(elem Element, jsonSchema *jsonschema.Schema, formData map[string]interface{}, validationErrors ValidationErrors, html *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)

	switch elem.Type {
	case "VerticalLayout":
		html.WriteString(indent + `<div class="vertical-layout">` + "\n")
		for _, child := range elem.Elements {
			u.renderElementWithData(child, jsonSchema, formData, validationErrors, html, depth+1)
		}
		html.WriteString(indent + "</div>\n")

	case "HorizontalLayout":
		html.WriteString(indent + `<div class="horizontal-layout">` + "\n")
		for _, child := range elem.Elements {
			html.WriteString(indent + `  <div class="horizontal-item">` + "\n")
			u.renderElementWithData(child, jsonSchema, formData, validationErrors, html, depth+2)
			html.WriteString(indent + `  </div>` + "\n")
		}
		html.WriteString(indent + "</div>\n")

	case "Group":
		html.WriteString(indent + `<fieldset class="form-group-section">` + "\n")
		if elem.Title != "" {
			html.WriteString(indent + `  <legend>` + elem.Title + `</legend>` + "\n")
		}
		for _, child := range elem.Elements {
			u.renderElementWithData(child, jsonSchema, formData, validationErrors, html, depth+1)
		}
		html.WriteString(indent + "</fieldset>\n")

	case "Control":
		u.renderControlWithData(elem, jsonSchema, formData, validationErrors, html, depth)

	case "Label":
		if elem.Text != "" {
			html.WriteString(indent + `<h4 class="ui-label">` + elem.Text + `</h4>` + "\n")
		}

	default:
		html.WriteString(indent + fmt.Sprintf("<!-- Unknown element type: %s -->\n", elem.Type))
	}
}

// renderControlWithData renders a form control with validation errors and form data
func (u *UISchema) renderControlWithData(elem Element, jsonSchema *jsonschema.Schema, formData map[string]interface{}, validationErrors ValidationErrors, html *strings.Builder, depth int) {
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

	// Get field value from formData (for pre-filling)
	var fieldValue string
	if formData != nil {
		if val, exists := formData[fieldName]; exists {
			fieldValue = fmt.Sprintf("%v", val)
		}
	}

	// Get validation error for this field
	var fieldError string
	if validationErrors != nil {
		if err, exists := validationErrors[fieldName]; exists {
			fieldError = err
		}
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
	u.renderInputWithData(elem, fieldName, prop, isRequired, fieldValue, html, indent)

	// Render validation error
	if fieldError != "" {
		html.WriteString(indent + `  <span class="field-error">` + fieldError + `</span>` + "\n")
	}

	// End form group
	html.WriteString(indent + "</div>\n")
}

// parseScopeToFieldName extracts field name from JSON pointer scope
func (u *UISchema) parseScopeToFieldName(scope string) string {
	// Scope format: "#/properties/fieldName"
	parts := strings.Split(scope, "/")
	if len(parts) >= 3 && parts[1] == "properties" {
		return parts[2]
	}
	return ""
}

// renderInputWithData renders an input field based on schema type with form data
func (u *UISchema) renderInputWithData(elem Element, fieldName string, prop *jsonschema.Schema, required bool, fieldValue string, html *strings.Builder, indent string) {
	requiredAttr := ""
	if required {
		requiredAttr = " required"
	}

	// Get placeholder from UI Schema options or schema examples
	placeholder := ""
	if elem.Options != nil && elem.Options.Placeholder != "" {
		placeholder = fmt.Sprintf(` placeholder="%s"`, elem.Options.Placeholder)
	} else if len(prop.Examples) > 0 {
		placeholder = fmt.Sprintf(` placeholder="%v"`, prop.Examples[0])
	}

	// Get format from UI Schema options or schema
	format := ""
	if elem.Options != nil && elem.Options.Format != "" {
		format = elem.Options.Format
	} else if prop.Format != nil {
		format = prop.Format.Name
	}

	// Determine type (from jsonschema Types field)
	propType := u.getSchemaType(prop)

	// Pre-filled value
	valueAttr := ""
	if fieldValue != "" {
		valueAttr = fmt.Sprintf(` value="%s"`, fieldValue)
	}

	switch propType {
	case "string":
		// Check for enum (dropdown)
		if prop.Enum != nil && len(prop.Enum.Values) > 0 {
			html.WriteString(indent + `  <select id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + `>` + "\n")
			html.WriteString(indent + `    <option value="">-- Select --</option>` + "\n")
			for _, option := range prop.Enum.Values {
				selected := ""
				if fmt.Sprintf("%v", option) == fieldValue {
					selected = " selected"
				}
				html.WriteString(indent + fmt.Sprintf(`    <option value="%v"%s>%v</option>`, option, selected, option) + "\n")
			}
			html.WriteString(indent + `  </select>` + "\n")
		} else if elem.Options != nil && elem.Options.Multi {
			// Multi-line text
			html.WriteString(indent + `  <textarea id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + placeholder + `>` + fieldValue + `</textarea>` + "\n")
		} else {
			// Single-line input with format-specific type
			inputType := "text"
			if format == "datetime-local" {
				inputType = "datetime-local"
			} else if format == "date" {
				inputType = "date"
			} else if format == "email" {
				inputType = "email"
			} else if format == "uri" || format == "url" {
				inputType = "url"
			}
			html.WriteString(indent + `  <input type="` + inputType + `" id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + placeholder + valueAttr + `>` + "\n")
		}

	case "boolean":
		checked := ""
		if fieldValue == "true" {
			checked = " checked"
		}
		html.WriteString(indent + `  <input type="checkbox" id="` + fieldName + `" name="` + fieldName + `" value="true"` + checked + `>` + "\n")

	case "integer", "number":
		min := ""
		max := ""
		if prop.Minimum != nil {
			min = fmt.Sprintf(` min="%v"`, *prop.Minimum)
		}
		if prop.Maximum != nil {
			max = fmt.Sprintf(` max="%v"`, *prop.Maximum)
		}
		html.WriteString(indent + `  <input type="number" id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + min + max + valueAttr + `>` + "\n")

	case "array":
		u.renderArrayInput(fieldName, prop, html, indent)

	case "object":
		u.renderObjectInput(fieldName, prop, html, indent)

	default:
		html.WriteString(indent + `  <input type="text" id="` + fieldName + `" name="` + fieldName + `"` + requiredAttr + placeholder + valueAttr + `>` + "\n")
	}
}

// getSchemaType determines the type from jsonschema.Schema
func (u *UISchema) getSchemaType(prop *jsonschema.Schema) string {
	if prop.Types == nil {
		return "string" // default
	}
	// Convert Types to string slice and take the first one
	types := prop.Types.ToStrings()
	if len(types) > 0 {
		return types[0]
	}
	return "string"
}

// renderArrayInput renders an array input with dynamic add/remove
func (u *UISchema) renderArrayInput(fieldName string, prop *jsonschema.Schema, html *strings.Builder, indent string) {
	title := prop.Title
	if title == "" {
		title = fieldName
	}

	html.WriteString(indent + `  <div class="array-input" data-field-name="` + fieldName + `">` + "\n")
	html.WriteString(indent + `    <div class="array-items" id="` + fieldName + `-items"></div>` + "\n")
	html.WriteString(indent + `    <button type="button" class="btn-add-array-item" onclick="addArrayItem('` + fieldName + `')">` + "\n")
	html.WriteString(indent + `      ➕ Add ` + title + "\n")
	html.WriteString(indent + `    </button>` + "\n")
	html.WriteString(indent + `    <template id="` + fieldName + `-template">` + "\n")

	// Render template for array items
	if prop.Items != nil {
		itemSchema, ok := prop.Items.(*jsonschema.Schema)
		if ok {
			u.renderArrayItemTemplate(fieldName, itemSchema, html, indent+"    ")
		}
	}

	html.WriteString(indent + `    </template>` + "\n")
	html.WriteString(indent + `  </div>` + "\n")
}

// renderArrayItemTemplate renders a template for array items
func (u *UISchema) renderArrayItemTemplate(fieldName string, itemSchema *jsonschema.Schema, html *strings.Builder, indent string) {
	html.WriteString(indent + `<div class="array-item">` + "\n")

	itemType := u.getSchemaType(itemSchema)

	if itemType == "object" && itemSchema.Properties != nil {
		// Array of objects - render nested fields
		for propName, propSchema := range itemSchema.Properties {
			isRequired := contains(itemSchema.Required, propName)
			label := propSchema.Title
			if label == "" {
				label = propName
			}
			html.WriteString(indent + `  <div class="form-group">` + "\n")
			html.WriteString(indent + `    <label>` + label)
			if isRequired {
				html.WriteString(" *")
			}
			html.WriteString(`</label>` + "\n")
			u.renderSimpleInput(fieldName+"[INDEX]."+propName, propSchema, isRequired, html, indent+"    ")
			html.WriteString(indent + `  </div>` + "\n")
		}
	} else {
		// Array of primitives
		u.renderSimpleInput(fieldName+"[INDEX]", itemSchema, false, html, indent+"  ")
	}

	html.WriteString(indent + `  <button type="button" class="btn-remove-array-item" onclick="removeArrayItem(this)">✖</button>` + "\n")
	html.WriteString(indent + `</div>` + "\n")
}

// renderSimpleInput renders a simple input (used in array templates)
func (u *UISchema) renderSimpleInput(fieldName string, prop *jsonschema.Schema, required bool, html *strings.Builder, indent string) {
	requiredAttr := ""
	if required {
		requiredAttr = " required"
	}

	propType := u.getSchemaType(prop)
	placeholder := prop.Title
	if placeholder == "" {
		placeholder = fieldName
	}

	switch propType {
	case "string":
		format := ""
		if prop.Format != nil {
			format = prop.Format.Name
		}
		if format == "email" {
			html.WriteString(indent + `<input type="email" name="` + fieldName + `"` + requiredAttr + ` placeholder="` + placeholder + `">` + "\n")
		} else {
			html.WriteString(indent + `<input type="text" name="` + fieldName + `"` + requiredAttr + ` placeholder="` + placeholder + `">` + "\n")
		}

	case "integer", "number":
		html.WriteString(indent + `<input type="number" name="` + fieldName + `"` + requiredAttr + ` placeholder="` + placeholder + `">` + "\n")
	default:
		html.WriteString(indent + `<input type="text" name="` + fieldName + `"` + requiredAttr + ` placeholder="` + placeholder + `">` + "\n")
	}
}

// renderObjectInput renders a nested object input
func (u *UISchema) renderObjectInput(fieldName string, prop *jsonschema.Schema, html *strings.Builder, indent string) {
	html.WriteString(indent + `  <div class="object-input">` + "\n")

	if prop.Properties != nil {
		for propName, propSchema := range prop.Properties {
			isRequired := contains(prop.Required, propName)
			label := propSchema.Title
			if label == "" {
				label = propName
			}
			html.WriteString(indent + `    <div class="form-group">` + "\n")
			html.WriteString(indent + `      <label>` + label)
			if isRequired {
				html.WriteString(" *")
			}
			html.WriteString(`</label>` + "\n")
			u.renderNestedInput(fieldName+"."+propName, propSchema, isRequired, html, indent+"      ")
			html.WriteString(indent + `    </div>` + "\n")
		}
	}

	html.WriteString(indent + `  </div>` + "\n")
}

// renderNestedInput renders an input for nested object properties
func (u *UISchema) renderNestedInput(fieldName string, prop *jsonschema.Schema, required bool, html *strings.Builder, indent string) {
	requiredAttr := ""
	if required {
		requiredAttr = " required"
	}

	propType := u.getSchemaType(prop)

	switch propType {
	case "string":
		format := ""
		if prop.Format != nil {
			format = prop.Format.Name
		}
		if format == "date" {
			html.WriteString(indent + `<input type="date" name="` + fieldName + `"` + requiredAttr + `>` + "\n")
		} else if prop.Enum != nil && len(prop.Enum.Values) > 0 {
			html.WriteString(indent + `<select name="` + fieldName + `"` + requiredAttr + `>` + "\n")
			html.WriteString(indent + `  <option value="">-- Select --</option>` + "\n")
			for _, option := range prop.Enum.Values {
				html.WriteString(indent + fmt.Sprintf(`  <option value="%v">%v</option>`, option, option) + "\n")
			}
			html.WriteString(indent + `</select>` + "\n")
		} else {
			html.WriteString(indent + `<input type="text" name="` + fieldName + `"` + requiredAttr + `>` + "\n")
		}
	case "integer", "number":
		html.WriteString(indent + `<input type="number" name="` + fieldName + `"` + requiredAttr + `>` + "\n")
	default:
		html.WriteString(indent + `<input type="text" name="` + fieldName + `"` + requiredAttr + `>` + "\n")
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
