# Calendar Package - Shared Calendar Integration Types

**Package**: `github.com/joeblew999/wellknown/pkg/calendar`

## Purpose

This package provides **shared types, constants, and interfaces** for all calendar platform integrations (Google Calendar, Apple Calendar, Microsoft Outlook, etc.).

By extracting common patterns into a shared package, we achieve:
- ✅ **Zero duplication** across platform implementations
- ✅ **Type safety** with shared interfaces
- ✅ **Consistent field names** across all platforms
- ✅ **Reusable validation logic**
- ✅ **Platform-agnostic testing**

---

## Architecture

```
pkg/calendar/              # SHARED calendar types
├── fields.go              # Common field names & metadata
└── generator.go           # Interfaces & shared logic

pkg/google/calendar/       # Google Calendar implementation
├── calendar.go            # Imports pkg/calendar, adds Google-specific logic
└── ...

pkg/apple/calendar/        # Apple Calendar implementation
├── calendar.go            # Imports pkg/calendar, adds Apple-specific logic
└── ...

pkg/microsoft/outlook/     # Future: Microsoft Outlook implementation
├── calendar.go            # Imports pkg/calendar, adds MS-specific logic
└── ...
```

---

## Shared Constants

### Field Names (`fields.go`)

All platforms use the same field names for common calendar event properties:

```go
import "github.com/joeblew999/wellknown/pkg/calendar"

// Basic fields (required)
calendar.FieldTitle       // "title"
calendar.FieldStart       // "start"
calendar.FieldEnd         // "end"

// Optional basic fields
calendar.FieldLocation    // "location"
calendar.FieldDescription // "description"
calendar.FieldAllDay      // "allDay"

// Advanced fields (platform-dependent)
calendar.FieldAttendees   // "attendees"
calendar.FieldRecurrence  // "recurrence"
calendar.FieldReminders   // "reminders"
calendar.FieldOrganizer   // "organizer"
```

**Benefits**:
- Change field name once, affects all platforms
- Type-safe field access with IDE autocomplete
- Consistent naming across Google, Apple, Microsoft, etc.

### Field Collections

Pre-defined field groups for common use cases:

```go
calendar.BasicFields          // [title, start, end]
calendar.OptionalBasicFields  // [location, description]
calendar.AdvancedFields       // [recurrence, attendees, reminders, ...]
```

### Time Formats

Standard Go time format strings:

```go
calendar.DateTimeLocalFormat  // "2006-01-02T15:04" (HTML datetime-local)
calendar.DateOnlyFormat       // "2006-01-02" (all-day events)
calendar.ISO8601Format        // ISO 8601 datetime
calendar.RFC3339Format        // RFC 3339 datetime
```

---

## Shared Interfaces

### `Generator` Interface

All calendar implementations must satisfy this interface:

```go
type Generator interface {
    Platform() string                                         // "google", "apple", etc.
    Generate(data map[string]interface{}) (interface{}, error)
    SupportsAdvancedFeatures() bool
    SupportedFields() []string
}
```

**Enables**:
- Platform-agnostic code (accept `Generator` instead of concrete type)
- Registry pattern (store `[]Generator`)
- Generic testing

### `URLGenerator` Interface

For platforms that generate URLs (Google Calendar, Microsoft Outlook):

```go
type URLGenerator interface {
    Generator
    GenerateURL(data map[string]interface{}) (string, error)
}
```

### `FileGenerator` Interface

For platforms that generate files (Apple Calendar .ics):

```go
type FileGenerator interface {
    Generator
    GenerateFile(data map[string]interface{}) (content []byte, mimeType string, err error)
}
```

---

## Shared Types

### `EventData`

Common calendar event data structure:

```go
type EventData struct {
    // Basic fields
    Title       string
    Start       time.Time
    End         time.Time
    Location    string
    Description string
    AllDay      bool

    // Advanced fields
    Attendees  []Attendee
    Recurrence *RecurrenceRule
    Reminders  []Reminder
    Organizer  *Organizer
    Status     EventStatus
    Priority   EventPriority
}
```

### `Attendee`, `RecurrenceRule`, `Reminder`, etc.

Standardized types for advanced calendar features. Platform implementations can use these directly or extend them.

---

## Shared Helper Functions

### `ParseEventData()`

Extracts common event data from validated form data:

```go
func ParseEventData(data map[string]interface{}) (*EventData, error)
```

**Usage**:
```go
// In platform implementation
func GenerateURL(data map[string]interface{}) (string, error) {
    event, err := calendar.ParseEventData(data)
    if err != nil {
        return "", err
    }

    // Use event.Title, event.Start, event.End, etc.
    // Add platform-specific logic here
}
```

### `IsAdvancedEvent()`

Checks if event uses advanced features:

```go
func IsAdvancedEvent(data map[string]interface{}) bool
```

Used for test categorization (basic vs advanced).

### `GetUsedFields()`

Returns list of fields present in event data:

```go
func GetUsedFields(data map[string]interface{}) []string
```

Useful for test coverage analysis.

---

## Usage in Platform Implementations

### Google Calendar Example

```go
package calendar

import (
    "fmt"
    "net/url"

    "github.com/joeblew999/wellknown/pkg/calendar"
)

// Platform-specific constants
const (
    BaseURL     = "https://calendar.google.com/calendar/render"
    ActionParam = "TEMPLATE"
    // ...
)

// Use shared field names
func GenerateURL(data map[string]interface{}) (string, error) {
    // Use calendar.FieldTitle instead of hardcoded "title"
    title, ok := data[calendar.FieldTitle].(string)
    if !ok {
        return "", fmt.Errorf("missing title")
    }

    // Use calendar.FieldStart instead of hardcoded "start"
    startStr, ok := data[calendar.FieldStart].(string)
    // ...

    // Build Google-specific URL
    params := url.Values{}
    params.Set("action", ActionParam)
    params.Set("text", title)
    // ...

    return BaseURL + "?" + params.Encode(), nil
}

// Implement Generator interface
func (g *GoogleCalendar) Platform() string {
    return "google"
}

func (g *GoogleCalendar) SupportsAdvancedFeatures() bool {
    return false // Google Calendar web URLs don't support attendees, recurrence, etc.
}

func (g *GoogleCalendar) SupportedFields() []string {
    return []string{
        calendar.FieldTitle,
        calendar.FieldStart,
        calendar.FieldEnd,
        calendar.FieldLocation,
        calendar.FieldDescription,
    }
}
```

### Apple Calendar Example

```go
package calendar

import (
    "github.com/joeblew999/wellknown/pkg/calendar"
)

// Platform-specific ICS constants
const (
    ICSBeginCalendar = "BEGIN:VCALENDAR"
    ICSVersion       = "2.0"
    // ...
)

// Use shared field names
func GenerateICS(data map[string]interface{}) ([]byte, error) {
    // Use shared field constants
    title, _ := data[calendar.FieldTitle].(string)
    startStr, _ := data[calendar.FieldStart].(string)

    // Generate ICS file using Apple-specific format
    var buf bytes.Buffer
    buf.WriteString(ICSBeginCalendar + "\r\n")
    buf.WriteString("SUMMARY:" + title + "\r\n")
    // ...

    return buf.Bytes(), nil
}

// Implement Generator interface
func (a *AppleCalendar) Platform() string {
    return "apple"
}

func (a *AppleCalendar) SupportsAdvancedFeatures() bool {
    return true // Apple Calendar ICS supports everything
}

func (a *AppleCalendar) SupportedFields() []string {
    return []string{
        calendar.FieldTitle,
        calendar.FieldStart,
        calendar.FieldEnd,
        calendar.FieldLocation,
        calendar.FieldDescription,
        calendar.FieldAllDay,
        calendar.FieldAttendees,
        calendar.FieldRecurrence,
        calendar.FieldReminders,
    }
}
```

---

## Benefits of Shared Package

### 1. Zero Duplication

**Before** (without shared package):
```go
// pkg/google/calendar/calendar.go
const FieldTitle = "title"
const FieldStart = "start"

// pkg/apple/calendar/calendar.go
const FieldTitle = "title"  // DUPLICATE!
const FieldStart = "start"  // DUPLICATE!

// pkg/microsoft/outlook/calendar.go
const FieldTitle = "title"  // DUPLICATE!
const FieldStart = "start"  // DUPLICATE!
```

**After** (with shared package):
```go
// pkg/calendar/fields.go
const FieldTitle = "title"  // ONCE!
const FieldStart = "start"  // ONCE!

// All platforms import and use
import "github.com/joeblew999/wellknown/pkg/calendar"
```

### 2. Type-Safe Platform Registry

```go
// cmd/testdata-gen/main.go or pkg/server/registry.go
type PlatformRegistry struct {
    generators map[string]calendar.Generator
}

func (r *PlatformRegistry) Register(gen calendar.Generator) {
    r.generators[gen.Platform()] = gen
}

// Register all platforms
registry.Register(google.NewGenerator())
registry.Register(apple.NewGenerator())
registry.Register(microsoft.NewGenerator())

// Use platform-agnostically
for name, gen := range registry.generators {
    output, err := gen.Generate(data)
    if gen.SupportsAdvancedFeatures() {
        // Enable advanced UI features
    }
}
```

### 3. Generic Testing

```go
// tests/generator_test.go
func TestAllGenerators(t *testing.T) {
    generators := []calendar.Generator{
        google.NewGenerator(),
        apple.NewGenerator(),
        microsoft.NewGenerator(),
    }

    for _, gen := range generators {
        t.Run(gen.Platform(), func(t *testing.T) {
            // Test with basic event
            basicEvent := map[string]interface{}{
                calendar.FieldTitle: "Test Event",
                calendar.FieldStart: "2025-11-01T10:00",
                calendar.FieldEnd:   "2025-11-01T11:00",
            }

            output, err := gen.Generate(basicEvent)
            assert.NoError(t, err)
            assert.NotNil(t, output)

            // Test advanced features only if supported
            if gen.SupportsAdvancedFeatures() {
                advancedEvent := basicEvent
                advancedEvent[calendar.FieldAttendees] = []map[string]interface{}{
                    {"email": "user@example.com", "required": true},
                }

                output, err := gen.Generate(advancedEvent)
                assert.NoError(t, err)
            }
        })
    }
}
```

### 4. Field Metadata

```go
// Auto-generate documentation from metadata
for fieldName, meta := range calendar.CommonFieldMetadata {
    fmt.Printf("Field: %s\n", fieldName)
    fmt.Printf("  Type: %s\n", meta.Type)
    fmt.Printf("  Required: %v\n", meta.Required)
    fmt.Printf("  Description: %s\n", meta.Description)
    if meta.Advanced {
        fmt.Println("  [ADVANCED FEATURE]")
    }
}
```

---

## Migration Guide

### Updating Existing Platforms

**Step 1**: Import shared calendar package
```go
import "github.com/joeblew999/wellknown/pkg/calendar"
```

**Step 2**: Replace local field constants with shared ones
```go
// Before
const FieldTitle = "title"
data["title"]

// After
calendar.FieldTitle
data[calendar.FieldTitle]
```

**Step 3**: Implement `Generator` interface (optional but recommended)
```go
type GoogleCalendar struct{}

func (g *GoogleCalendar) Platform() string { return "google" }
func (g *GoogleCalendar) Generate(data map[string]interface{}) (interface{}, error) {
    return GenerateURL(data)
}
func (g *GoogleCalendar) SupportsAdvancedFeatures() bool { return false }
func (g *GoogleCalendar) SupportedFields() []string { return []string{...} }
```

**Step 4**: Use shared helper functions
```go
// Instead of manual parsing
title := data["title"].(string)

// Use helper
event, _ := calendar.ParseEventData(data)
title := event.Title
```

---

## Future Enhancements

- [ ] Auto-generate TypeScript types from Go types
- [ ] JSON Schema generation from field metadata
- [ ] Validation functions based on field metadata
- [ ] Platform capability matrix (which platforms support which features)
- [ ] Calendar format converters (URL ↔ ICS ↔ JSON)

---

## Summary

The `pkg/calendar` package provides a **professional, reusable foundation** for all calendar integrations:

✅ **Shared field constants** - Define once, use everywhere
✅ **Standard interfaces** - Platform-agnostic code
✅ **Common types** - Consistent event data structures
✅ **Helper functions** - Reusable parsing and validation
✅ **Zero duplication** - DRY architecture
✅ **Type safety** - Compile-time correctness
✅ **Extensibility** - Easy to add new platforms

**Adding a new platform** (e.g., Microsoft Outlook) now requires:
1. Import `pkg/calendar`
2. Use shared field constants
3. Implement platform-specific generation logic
4. Optionally implement `Generator` interface

**That's it!** No need to redefine field names, types, or shared logic.
