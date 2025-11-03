package calendar

import "time"

// Generator is the interface that all calendar platform implementations must satisfy.
//
// This enables platform-agnostic code and testing.
type Generator interface {
	// Platform returns the platform name (e.g., "google", "apple", "microsoft")
	Platform() string

	// Generate creates a calendar deep link/file from validated form data.
	// The input data is a map of field names to values, already validated by JSON Schema.
	// Returns platform-specific output (URL string, ICS bytes, etc.) and any error.
	Generate(data map[string]interface{}) (interface{}, error)

	// SupportsAdvancedFeatures returns true if this platform supports advanced
	// calendar features like recurrence, attendees, reminders, etc.
	SupportsAdvancedFeatures() bool

	// SupportedFields returns the list of field names this platform supports.
	// Must include at least: title, start, end
	SupportedFields() []string
}

// URLGenerator generates calendar URLs (e.g., Google Calendar, Microsoft Outlook)
type URLGenerator interface {
	Generator

	// GenerateURL creates a calendar deep link URL from validated form data.
	GenerateURL(data map[string]interface{}) (string, error)
}

// FileGenerator generates calendar files (e.g., Apple Calendar .ics files)
type FileGenerator interface {
	Generator

	// GenerateFile creates a calendar file from validated form data.
	// Returns file bytes and MIME type.
	GenerateFile(data map[string]interface{}) (content []byte, mimeType string, err error)
}

// EventData represents common calendar event data.
// Platform-specific implementations can embed this and add their own fields.
type EventData struct {
	// Basic fields (required)
	Title string
	Start time.Time
	End   time.Time

	// Optional basic fields
	Location    string
	Description string
	AllDay      bool

	// Advanced fields (optional, platform-dependent)
	Attendees  []Attendee
	Recurrence *RecurrenceRule
	Reminders  []Reminder
	Organizer  *Organizer
	Status     EventStatus
	Priority   EventPriority
	URL        string
}

// Attendee represents an event attendee
type Attendee struct {
	Email    string
	Name     string
	Required bool   // Required participant (vs optional)
	Role     string // e.g., "chair", "req-participant", "opt-participant"
	RSVP     bool   // Request RSVP
}

// RecurrenceRule represents an event recurrence pattern
type RecurrenceRule struct {
	Frequency string // DAILY, WEEKLY, MONTHLY, YEARLY
	Interval  int    // Repeat every N frequency units
	Count     int    // Number of occurrences (0 = infinite)
	Until     *time.Time
	ByDay     []string // e.g., "MO", "TU", "WE"
	ByMonth   []int    // 1-12
}

// Reminder represents an event reminder/alarm
type Reminder struct {
	MinutesBefore int    // Minutes before event to trigger
	Method        string // e.g., "email", "popup", "sms"
}

// Organizer represents the event organizer
type Organizer struct {
	Email string
	Name  string
}

// EventStatus represents the event's confirmation status
type EventStatus string

const (
	StatusConfirmed  EventStatus = "CONFIRMED"
	StatusTentative  EventStatus = "TENTATIVE"
	StatusCancelled  EventStatus = "CANCELLED"
)

// EventPriority represents the event's priority level
type EventPriority int

const (
	PriorityUndefined EventPriority = 0
	PriorityHigh      EventPriority = 1
	PriorityNormal    EventPriority = 5
	PriorityLow       EventPriority = 9
)

// ParseEventData extracts common event data from a validated form data map.
// This is a helper function that platform implementations can use.
func ParseEventData(data map[string]interface{}) (*EventData, error) {
	event := &EventData{}

	// Parse basic required fields
	if title, ok := data[FieldTitle].(string); ok {
		event.Title = title
	}

	if startStr, ok := data[FieldStart].(string); ok {
		if t, err := time.Parse(DateTimeLocalFormat, startStr); err == nil {
			event.Start = t
		}
	}

	if endStr, ok := data[FieldEnd].(string); ok {
		if t, err := time.Parse(DateTimeLocalFormat, endStr); err == nil {
			event.End = t
		}
	}

	// Parse optional basic fields
	if location, ok := data[FieldLocation].(string); ok {
		event.Location = location
	}

	if description, ok := data[FieldDescription].(string); ok {
		event.Description = description
	}

	if allDay, ok := data[FieldAllDay].(bool); ok {
		event.AllDay = allDay
	}

	// Parse advanced fields (attendees, recurrence, etc.)
	// Platform implementations can extend this

	return event, nil
}

// IsAdvancedEvent returns true if the event uses any advanced features.
// Used for test categorization (basic vs advanced).
func IsAdvancedEvent(data map[string]interface{}) bool {
	for _, field := range AdvancedFields {
		if _, exists := data[field]; exists {
			return true
		}
	}
	return false
}

// GetUsedFields returns a list of field names present in the data.
// Useful for test coverage analysis.
func GetUsedFields(data map[string]interface{}) []string {
	fields := make([]string, 0, len(data))
	for field := range data {
		fields = append(fields, field)
	}
	return fields
}
