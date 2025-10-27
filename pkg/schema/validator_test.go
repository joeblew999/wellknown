package schema

import (
	"testing"
)

// TestValidateAgainstSchema_RequiredFields tests required field validation
func TestValidateAgainstSchema_RequiredFields(t *testing.T) {
	schema := &JSONSchema{
		Required: []string{"title", "start"},
		Properties: map[string]Property{
			"title": {Type: "string"},
			"start": {Type: "string"},
		},
	}

	tests := []struct {
		name          string
		data          map[string]interface{}
		wantErrors    int
		wantFieldErrs []string
	}{
		{
			name:       "All required fields present",
			data:       map[string]interface{}{"title": "Meeting", "start": "2025-10-28T10:00"},
			wantErrors: 0,
		},
		{
			name:          "Missing title",
			data:          map[string]interface{}{"start": "2025-10-28T10:00"},
			wantErrors:    1,
			wantFieldErrs: []string{"title"},
		},
		{
			name:          "Empty title",
			data:          map[string]interface{}{"title": "  ", "start": "2025-10-28T10:00"},
			wantErrors:    1,
			wantFieldErrs: []string{"title"},
		},
		{
			name:          "Missing all required fields",
			data:          map[string]interface{}{},
			wantErrors:    2,
			wantFieldErrs: []string{"title", "start"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateAgainstSchema(tt.data, schema)
			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateAgainstSchema() error count = %d, want %d. Errors: %v", len(errors), tt.wantErrors, errors)
			}
			for _, field := range tt.wantFieldErrs {
				if _, exists := errors[field]; !exists {
					t.Errorf("Expected error for field %s, but got none", field)
				}
			}
		})
	}
}

// TestValidateString tests string validation
func TestValidateString(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		prop      Property
		wantError bool
	}{
		{
			name:      "Valid string",
			value:     "Hello",
			prop:      Property{Type: "string"},
			wantError: false,
		},
		{
			name:      "Not a string",
			value:     123,
			prop:      Property{Type: "string"},
			wantError: true,
		},
		{
			name:      "String too short",
			value:     "Hi",
			prop:      Property{Type: "string", MinLength: 5},
			wantError: true,
		},
		{
			name:      "String too long",
			value:     "This is a very long string",
			prop:      Property{Type: "string", MaxLength: 10},
			wantError: true,
		},
		{
			name:      "String meets minLength",
			value:     "Hello",
			prop:      Property{Type: "string", MinLength: 5},
			wantError: false,
		},
		{
			name:      "String meets maxLength",
			value:     "Hello",
			prop:      Property{Type: "string", MaxLength: 10},
			wantError: false,
		},
		{
			name:      "Enum valid value",
			value:     "DAILY",
			prop:      Property{Type: "string", Enum: []string{"DAILY", "WEEKLY", "MONTHLY"}},
			wantError: false,
		},
		{
			name:      "Enum invalid value",
			value:     "YEARLY",
			prop:      Property{Type: "string", Enum: []string{"DAILY", "WEEKLY", "MONTHLY"}},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateString("testField", tt.value, tt.prop)
			if (err != nil) != tt.wantError {
				t.Errorf("validateString() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestValidateFormat tests format validation
func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		format    string
		wantError bool
	}{
		// Email validation
		{
			name:      "Valid email",
			value:     "test@example.com",
			format:    "email",
			wantError: false,
		},
		{
			name:      "Invalid email",
			value:     "not-an-email",
			format:    "email",
			wantError: true,
		},
		// URL validation
		{
			name:      "Valid URL",
			value:     "https://example.com",
			format:    "uri",
			wantError: false,
		},
		{
			name:      "Invalid URL",
			value:     "not a url",
			format:    "uri",
			wantError: true,
		},
		// Date validation
		{
			name:      "Valid date",
			value:     "2025-10-28",
			format:    "date",
			wantError: false,
		},
		{
			name:      "Invalid date format",
			value:     "28-10-2025",
			format:    "date",
			wantError: true,
		},
		{
			name:      "Invalid date",
			value:     "2025-13-45",
			format:    "date",
			wantError: true,
		},
		// Datetime-local validation
		{
			name:      "Valid datetime-local",
			value:     "2025-10-28T14:30",
			format:    "datetime-local",
			wantError: false,
		},
		{
			name:      "Invalid datetime-local",
			value:     "2025-10-28 14:30",
			format:    "datetime-local",
			wantError: true,
		},
		// Time validation
		{
			name:      "Valid time",
			value:     "14:30",
			format:    "time",
			wantError: false,
		},
		{
			name:      "Invalid time",
			value:     "25:70",
			format:    "time",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFormat(tt.value, tt.format)
			if (err != nil) != tt.wantError {
				t.Errorf("validateFormat() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestValidateNumber tests number and integer validation
func TestValidateNumber(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		prop      Property
		wantError bool
	}{
		// String to number conversion
		{
			name:      "Valid number string",
			value:     "42",
			prop:      Property{Type: "number"},
			wantError: false,
		},
		{
			name:      "Invalid number string",
			value:     "not-a-number",
			prop:      Property{Type: "number"},
			wantError: true,
		},
		// Integer validation
		{
			name:      "Valid integer",
			value:     "42",
			prop:      Property{Type: "integer"},
			wantError: false,
		},
		{
			name:      "Invalid integer (float)",
			value:     "42.5",
			prop:      Property{Type: "integer"},
			wantError: true,
		},
		// Minimum validation
		{
			name:      "Number meets minimum",
			value:     "10",
			prop:      Property{Type: "number", Minimum: 5},
			wantError: false,
		},
		{
			name:      "Number below minimum",
			value:     "3",
			prop:      Property{Type: "number", Minimum: 5},
			wantError: true,
		},
		// Maximum validation
		{
			name:      "Number meets maximum",
			value:     "10",
			prop:      Property{Type: "number", Maximum: 20},
			wantError: false,
		},
		{
			name:      "Number above maximum",
			value:     "25",
			prop:      Property{Type: "number", Maximum: 20},
			wantError: true,
		},
		// Native number types
		{
			name:      "Native float64",
			value:     float64(42.5),
			prop:      Property{Type: "number"},
			wantError: false,
		},
		{
			name:      "Native int",
			value:     int(42),
			prop:      Property{Type: "integer"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNumber("testField", tt.value, tt.prop)
			if (err != nil) != tt.wantError {
				t.Errorf("validateNumber() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestValidateBoolean tests boolean validation
func TestValidateBoolean(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{name: "Native true", value: true, wantError: false},
		{name: "Native false", value: false, wantError: false},
		{name: "String 'true'", value: "true", wantError: false},
		{name: "String 'false'", value: "false", wantError: false},
		{name: "String 'on' (checkbox)", value: "on", wantError: false},
		{name: "Empty string (unchecked)", value: "", wantError: false},
		{name: "Invalid string", value: "yes", wantError: true},
		{name: "Number", value: 1, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBoolean("testField", tt.value, Property{Type: "boolean"})
			if (err != nil) != tt.wantError {
				t.Errorf("validateBoolean() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestFormDataToMap tests form data conversion
func TestFormDataToMap(t *testing.T) {
	tests := []struct {
		name     string
		formData map[string][]string
		want     map[string]interface{}
	}{
		{
			name: "Simple fields",
			formData: map[string][]string{
				"title": {"Meeting"},
				"start": {"2025-10-28T10:00"},
			},
			want: map[string]interface{}{
				"title": "Meeting",
				"start": "2025-10-28T10:00",
			},
		},
		{
			name: "Nested object (dot notation)",
			formData: map[string][]string{
				"title":          {"Meeting"},
				"organizer.name": {"John"},
			},
			want: map[string]interface{}{
				"title": "Meeting",
				"organizer": map[string]interface{}{
					"name": "John",
				},
			},
		},
		{
			name: "Deep nesting",
			formData: map[string][]string{
				"event.organizer.name":  {"John"},
				"event.organizer.email": {"john@example.com"},
			},
			want: map[string]interface{}{
				"event": map[string]interface{}{
					"organizer": map[string]interface{}{
						"name":  "John",
						"email": "john@example.com",
					},
				},
			},
		},
		{
			name: "Empty values ignored",
			formData: map[string][]string{
				"title": {"Meeting"},
				"empty": {},
			},
			want: map[string]interface{}{
				"title": "Meeting",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormDataToMap(tt.formData)
			if !equalMaps(got, tt.want) {
				t.Errorf("FormDataToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetNestedValue tests nested value retrieval
func TestGetNestedValue(t *testing.T) {
	data := map[string]interface{}{
		"title": "Meeting",
		"organizer": map[string]interface{}{
			"name":  "John",
			"email": "john@example.com",
		},
	}

	tests := []struct {
		name       string
		path       string
		wantValue  interface{}
		wantExists bool
	}{
		{
			name:       "Simple field",
			path:       "title",
			wantValue:  "Meeting",
			wantExists: true,
		},
		{
			name:       "Nested field",
			path:       "organizer.name",
			wantValue:  "John",
			wantExists: true,
		},
		{
			name:       "Deep nested field",
			path:       "organizer.email",
			wantValue:  "john@example.com",
			wantExists: true,
		},
		{
			name:       "Non-existent field",
			path:       "location",
			wantValue:  nil,
			wantExists: false,
		},
		{
			name:       "Non-existent nested field",
			path:       "organizer.phone",
			wantValue:  nil,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := GetNestedValue(data, tt.path)
			if exists != tt.wantExists {
				t.Errorf("GetNestedValue() exists = %v, want %v", exists, tt.wantExists)
			}
			if exists && got != tt.wantValue {
				t.Errorf("GetNestedValue() value = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

// Helper function to compare maps
func equalMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok {
			return false
		} else {
			switch av := v.(type) {
			case map[string]interface{}:
				if bvMap, ok := bv.(map[string]interface{}); !ok {
					return false
				} else if !equalMaps(av, bvMap) {
					return false
				}
			default:
				if av != bv {
					return false
				}
			}
		}
	}
	return true
}

// TestValidateCrossFieldRule tests custom cross-field validation (x-validations)
func TestValidateCrossFieldRule(t *testing.T) {
	tests := []struct {
		name      string
		schema    *JSONSchema
		data      map[string]interface{}
		wantError bool
		wantField string
		wantMsg   string
	}{
		{
			name: "Valid: end after start",
			schema: &JSONSchema{
				Properties: map[string]Property{
					"start": {Type: "string", Format: "datetime-local"},
					"end":   {Type: "string", Format: "datetime-local"},
				},
				Required: []string{"start", "end"},
				XValidations: map[string]CrossFieldRule{
					"endAfterStart": {
						Fields:  []string{"end", "start"},
						Message: "End time must be after start time",
					},
				},
			},
			data: map[string]interface{}{
				"start": "2025-10-28T10:00",
				"end":   "2025-10-28T11:00",
			},
			wantError: false,
		},
		{
			name: "Invalid: end before start",
			schema: &JSONSchema{
				Properties: map[string]Property{
					"start": {Type: "string", Format: "datetime-local"},
					"end":   {Type: "string", Format: "datetime-local"},
				},
				Required: []string{"start", "end"},
				XValidations: map[string]CrossFieldRule{
					"endAfterStart": {
						Fields:  []string{"end", "start"},
						Message: "End time must be after start time",
					},
				},
			},
			data: map[string]interface{}{
				"start": "2025-10-28T11:00",
				"end":   "2025-10-28T10:00",
			},
			wantError: true,
			wantField: "end",
			wantMsg:   "End time must be after start time",
		},
		{
			name: "Invalid: end equals start",
			schema: &JSONSchema{
				Properties: map[string]Property{
					"start": {Type: "string", Format: "datetime-local"},
					"end":   {Type: "string", Format: "datetime-local"},
				},
				Required: []string{"start", "end"},
				XValidations: map[string]CrossFieldRule{
					"endAfterStart": {
						Fields:  []string{"end", "start"},
						Message: "End time must be after start time",
					},
				},
			},
			data: map[string]interface{}{
				"start": "2025-10-28T10:00",
				"end":   "2025-10-28T10:00",
			},
			wantError: true,
			wantField: "end",
			wantMsg:   "End time must be after start time",
		},
		{
			name: "Skip: missing start field",
			schema: &JSONSchema{
				Properties: map[string]Property{
					"start": {Type: "string", Format: "datetime-local"},
					"end":   {Type: "string", Format: "datetime-local"},
				},
				Required: []string{"start", "end"},
				XValidations: map[string]CrossFieldRule{
					"endAfterStart": {
						Fields:  []string{"end", "start"},
						Message: "End time must be after start time",
					},
				},
			},
			data: map[string]interface{}{
				"end": "2025-10-28T11:00",
			},
			wantError: true, // Required field error, not cross-field error
			wantField: "start",
			wantMsg:   "This field is required",
		},
		{
			name: "Skip: empty optional fields",
			schema: &JSONSchema{
				Properties: map[string]Property{
					"start": {Type: "string", Format: "datetime-local"},
					"end":   {Type: "string", Format: "datetime-local"},
				},
				XValidations: map[string]CrossFieldRule{
					"endAfterStart": {
						Fields:  []string{"end", "start"},
						Message: "End time must be after start time",
					},
				},
			},
			data: map[string]interface{}{
				"start": "",
				"end":   "",
			},
			wantError: false, // Empty optional fields, skip validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateAgainstSchema(tt.data, tt.schema)

			if tt.wantError {
				if len(errors) == 0 {
					t.Errorf("Expected validation error, got none")
					return
				}
				if tt.wantField != "" {
					if msg, exists := errors[tt.wantField]; !exists {
						t.Errorf("Expected error on field %q, got errors: %v", tt.wantField, errors)
					} else if tt.wantMsg != "" && msg != tt.wantMsg {
						t.Errorf("Expected error message %q, got %q", tt.wantMsg, msg)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("Expected no validation errors, got: %v", errors)
				}
			}
		})
	}
}
