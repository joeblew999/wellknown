# JSON Schema Implementation Plan

**Based on**: `.src/goPocJsonSchemaForm` reference code patterns

## Current State (Completed)

### Phase 7: JSON Schema (DONE ✅)
- JSON Schema parser (stdlib only)
- Auto-form generation from JSON Schema
- Route: `/google/calendar-schema`

### Phase 8: UI Schema (DONE ✅)
- UI Schema for layout control
- VerticalLayout, HorizontalLayout, Control, Label, Group
- Route: `/google/calendar-uischema`

## Patterns from Reference Code

### Key Concepts Extracted:

1. **Nested Object Properties** ✅ Partially Implemented
   ```json
   "#/properties/personalData/properties/age"
   ```
   - Allows accessing nested fields in schemas
   - Our UI Schema parser handles this

2. **Validation** ⏳ TODO
   - Reference code uses `qri-io/jsonschema` library for validation
   - We need stdlib-only validation
   - Validate on POST, return errors to form

3. **Form Data Binding** ⏳ TODO
   - `BindToSchema()` - binds form values + errors to controls
   - `FormValuesToGroupedMap()` - converts flat form to nested structure
   - Needed for nested objects (attendees, recurrence, etc.)

4. **Suggestions/Autocomplete** ⏳ TODO
   ```json
   {
     "type": "Control",
     "scope": "#/properties/occupation",
     "suggestion": ["Engineer", "Teacher", "Student"]
   }
   ```

5. **Multi-Screen Support** ❌ Not Needed
   - Reference code loads schemas from `./screens/` directory
   - We inline schemas in code - simpler!

## Next Steps (Phase 9)

### 9a: Add Validation Support

**Goal**: Validate form submissions against JSON Schema

**Implementation**:
```go
// pkg/server/validator.go
func ValidateAgainstSchema(data map[string]interface{}, schemaJSON string) map[string]string {
    // Parse schema
    // Check required fields
    // Check types (string, number, boolean)
    // Check constraints (minLength, maxLength, minimum, maximum)
    // Check formats (email, datetime-local, uri)
    // Return map of field -> error message
}
```

**stdlib only** - no external validation libraries!

### 9b: Support Nested Object Form Binding

**Goal**: Handle nested objects in forms (attendees, recurrence, organizer)

**Challenge**: HTML forms are flat, but we need nested structure
```html
<!-- Flat form submission -->
<input name="organizer.name" value="John Doe">
<input name="organizer.email" value="john@example.com">
```

**Solution**: Parse dot-notation or array notation
```go
// organizer.name -> map["organizer"]["name"]
// attendees[0].email -> map["attendees"][0]["email"]
```

### 9c: Add Suggestions/Autocomplete

**Goal**: Support `datalist` for autocomplete from UI Schema

**Example**:
```json
{
  "type": "Control",
  "scope": "#/properties/location",
  "options": {
    "suggestions": ["Conference Room A", "Conference Room B", "Zoom"]
  }
}
```

Renders as:
```html
<input list="location-suggestions" ...>
<datalist id="location-suggestions">
  <option value="Conference Room A">
  <option value="Conference Room B">
  <option value="Zoom">
</datalist>
```

### 9d: Array Input Support

**Goal**: Add/remove dynamic form fields for arrays (attendees, reminders)

**UI Pattern**:
```
Attendees:
┌─────────────────────────────────────┐
│ Email: john@example.com             │
│ Name: John Doe                      │
│ ☑ Required                          │
│ [Remove]                            │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│ Email: jane@example.com             │
│ Name: Jane Smith                    │
│ ☑ Required                          │
│ [Remove]                            │
└─────────────────────────────────────┘
[+ Add Attendee]
```

**Implementation**: Client-side JavaScript for add/remove (progressive enhancement)

## Routes Strategy

### Keep All 3 Approaches Working

**Google Calendar**:
1. `/google/calendar` - Hardcoded HTML ✅
2. `/google/calendar-schema` - JSON Schema only ✅
3. `/google/calendar-uischema` - UI Schema + JSON Schema ✅

**Apple Calendar**:
1. `/apple/calendar` - Hardcoded HTML ✅
2. `/apple/calendar-schema` - JSON Schema only ⏳ TODO
3. `/apple/calendar-uischema` - UI Schema + JSON Schema ⏳ TODO

**Future Services** (Maps, Slack, Zoom, etc.):
- Start with JSON Schema + UI Schema approach
- Skip hardcoded HTML (use schema pattern from the start!)

## Benefits of JSON Schema Approach

1. **Declarative**: Define schema once, form generates automatically
2. **Validation**: Built-in type checking and constraints
3. **Self-documenting**: Schema describes the API
4. **Flexible**: UI Schema separates layout from data
5. **Maintainable**: Add new services with just JSON
6. **Type-safe**: JSON Schema ensures correct data types
7. **stdlib only**: Zero external dependencies in core library

## Migration Path

### For New Services:
```
1. Define JSON Schema (validation + types)
2. Define UI Schema (layout + presentation)
3. Done! Form + validation work automatically
```

### For Existing Services:
```
1. Keep hardcoded version working
2. Add JSON Schema version
3. Add UI Schema version
4. All 3 approaches coexist
```

## Reference Code Patterns Not Used

**Why we skip some patterns**:

1. **qri-io/jsonschema library** - We use stdlib only
2. **Echo framework** - We use stdlib `net/http`
3. **Multi-screen directory loading** - We inline schemas in code
4. **HTMX** - We start with server-side rendering, add HTMX later if needed

## File Organization

```
pkg/
├── google/calendar/
│   ├── schema.go       # JSON Schema definition
│   ├── uischema.go     # UI Schema layout
│   ├── event.go        # Event struct + GenerateURL()
│   └── examples.go     # Example events
├── apple/calendar/
│   ├── schema.go       # JSON Schema (with nested objects)
│   ├── uischema.go     # UI Schema (TODO)
│   ├── types.go        # ICS types (Attendee, Recurrence, etc.)
│   ├── event.go        # Event struct + GenerateICS()
│   └── examples.go     # Example events
└── server/
    ├── schema.go       # JSON Schema parser
    ├── uischema.go     # UI Schema renderer
    ├── validator.go    # Validation (TODO)
    └── formbinder.go   # Nested object binding (TODO)
```

---

**Last Updated**: 2025-10-27
**Status**: Phase 8 complete, Phase 9 planned
