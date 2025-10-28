package schema

// Standard JSON filenames used in platform directories (e.g., pkg/google/calendar/)
//
// These constants ensure consistent naming across the codebase and make it easy
// to find all places where these files are referenced.
const (
	// SchemaFilename is the JSON Schema file that defines validation rules
	// Example: pkg/google/calendar/schema.json
	SchemaFilename = "schema.json"

	// UISchemaFilename is the UI Schema file that defines form layout and rendering
	// Example: pkg/google/calendar/uischema.json
	UISchemaFilename = "uischema.json"

	// ExamplesFilename contains valid, user-facing examples used for:
	// - Showcase page (web UI)
	// - Valid test cases (Go tests + Playwright)
	// Example: pkg/google/calendar/data-examples.json
	ExamplesFilename = "data-examples.json"

	// FailuresFilename contains test data that should fail validation:
	// - Invalid inputs (missing fields, wrong formats)
	// - Edge cases (boundary tests)
	// Used only for testing (Go tests + Playwright)
	// Example: pkg/google/calendar/data-failures.json
	FailuresFilename = "data-failures.json"
)
