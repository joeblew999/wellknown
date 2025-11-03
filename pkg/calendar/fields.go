// Package calendar provides shared types and constants for calendar integrations.
//
// This package defines common field names, types, and interfaces used across
// all calendar implementations (Google Calendar, Apple Calendar, Microsoft Outlook, etc.).
//
// Platform-specific implementations should embed these shared types and add
// their platform-specific constants and logic.
package calendar

// Common Calendar Field Names
//
// These field names are used across ALL calendar platforms.
// They represent the JSON keys in validated form data.
//
// Basic Fields (required for all calendar events):
const (
	FieldTitle       = "title"        // Event title/summary
	FieldStart       = "start"        // Start time (datetime-local format: "2006-01-02T15:04")
	FieldEnd         = "end"          // End time (datetime-local format: "2006-01-02T15:04")
	FieldLocation    = "location"     // Event location
	FieldDescription = "description"  // Event description/details
)

// Advanced Fields (optional, platform-dependent):
const (
	FieldAllDay     = "allDay"      // Boolean: All-day event flag
	FieldAttendees  = "attendees"   // Array of attendee objects
	FieldRecurrence = "recurrence"  // Recurrence rule object
	FieldReminders  = "reminders"   // Array of reminder objects
	FieldOrganizer  = "organizer"   // Organizer object
	FieldAlarm      = "alarm"       // Alarm/notification settings
	FieldStatus     = "status"      // Event status (confirmed, tentative, cancelled)
	FieldPriority   = "priority"    // Event priority (low, medium, high)
	FieldURL        = "url"         // Associated URL
)

// BasicFields lists the minimum required fields for a calendar event.
// All platforms MUST support these fields.
var BasicFields = []string{
	FieldTitle,
	FieldStart,
	FieldEnd,
}

// OptionalBasicFields are commonly supported optional fields.
// Most platforms support these, but they're not strictly required.
var OptionalBasicFields = []string{
	FieldLocation,
	FieldDescription,
}

// AdvancedFields lists fields that indicate advanced calendar functionality.
// These are used for test categorization (basic vs advanced).
// Not all platforms support all advanced fields.
var AdvancedFields = []string{
	FieldRecurrence,
	FieldAttendees,
	FieldReminders,
	FieldOrganizer,
	FieldAlarm,
}

// Common Time Formats
//
// These are standard Go time format strings used across platforms.
const (
	// DateTimeLocalFormat is the HTML datetime-local input format
	DateTimeLocalFormat = "2006-01-02T15:04"

	// DateOnlyFormat is for all-day events
	DateOnlyFormat = "2006-01-02"

	// ISO8601Format is the standard ISO 8601 datetime format
	ISO8601Format = "2006-01-02T15:04:05Z07:00"

	// RFC3339Format is Go's RFC 3339 datetime format
	RFC3339Format = "2006-01-02T15:04:05Z07:00"
)

// FieldType represents the JSON Schema type of a field
type FieldType string

const (
	TypeString  FieldType = "string"
	TypeBoolean FieldType = "boolean"
	TypeNumber  FieldType = "number"
	TypeInteger FieldType = "integer"
	TypeObject  FieldType = "object"
	TypeArray   FieldType = "array"
)

// FieldMetadata describes a calendar field's properties
type FieldMetadata struct {
	Name        string    // Field name (e.g., "title")
	Type        FieldType // JSON Schema type
	Required    bool      // Is this field required?
	Description string    // Human-readable description
	Format      string    // Format hint (e.g., "date-time", "email")
	Advanced    bool      // Is this an advanced feature?
}

// CommonFieldMetadata provides metadata for all common calendar fields.
// Platform-specific implementations can extend this with their own fields.
var CommonFieldMetadata = map[string]FieldMetadata{
	FieldTitle: {
		Name:        FieldTitle,
		Type:        TypeString,
		Required:    true,
		Description: "Event title or summary",
		Advanced:    false,
	},
	FieldStart: {
		Name:        FieldStart,
		Type:        TypeString,
		Required:    true,
		Description: "Event start time",
		Format:      "datetime-local",
		Advanced:    false,
	},
	FieldEnd: {
		Name:        FieldEnd,
		Type:        TypeString,
		Required:    true,
		Description: "Event end time",
		Format:      "datetime-local",
		Advanced:    false,
	},
	FieldLocation: {
		Name:        FieldLocation,
		Type:        TypeString,
		Required:    false,
		Description: "Event location",
		Advanced:    false,
	},
	FieldDescription: {
		Name:        FieldDescription,
		Type:        TypeString,
		Required:    false,
		Description: "Event description or notes",
		Advanced:    false,
	},
	FieldAllDay: {
		Name:        FieldAllDay,
		Type:        TypeBoolean,
		Required:    false,
		Description: "All-day event flag",
		Advanced:    false,
	},
	FieldAttendees: {
		Name:        FieldAttendees,
		Type:        TypeArray,
		Required:    false,
		Description: "List of event attendees",
		Advanced:    true,
	},
	FieldRecurrence: {
		Name:        FieldRecurrence,
		Type:        TypeObject,
		Required:    false,
		Description: "Recurring event rule",
		Advanced:    true,
	},
	FieldReminders: {
		Name:        FieldReminders,
		Type:        TypeArray,
		Required:    false,
		Description: "Event reminders/notifications",
		Advanced:    true,
	},
}
