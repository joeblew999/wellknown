# Deep Architecture Analysis: Will Schema-Only Approach Work?

**Date**: 2025-10-28
**Analysis**: Critical examination of proposed schema-driven architecture

---

## ❌ PROBLEM FOUND: Current Code STILL Uses event.go!

### Current Flow (Lines from calendar_generic.go):

```go
// Line 86: Convert form to map
formData := schema.FormDataToMap(r.Form)

// Line 89: Validate against schema
validationErrors := schema.ValidateAgainstSchema(formData, jsonSchema)

// Line 99: ❌ THIS IS THE PROBLEM!
event, err := cfg.BuildEvent(r)  // <-- Creates typed Event struct!

// Line 106: Generate URL from typed event
generatedURL, err := cfg.GenerateURL(event)
```

### What BuildEvent Actually Does (google_calendar.go:15-30):

```go
BuildEvent: func(r *http.Request) (interface{}, error) {
    // Manual parsing from form
    startTime, err := parseFormTime(r.FormValue("start"))
    endTime, err := parseFormTime(r.FormValue("end"))

    // ❌ Returns googlecalendar.Event struct!
    return googlecalendar.Event{
        Title:       r.FormValue("title"),
        StartTime:   startTime,
        EndTime:     endTime,
        Location:    r.FormValue("location"),
        Description: r.FormValue("description"),
    }, nil
}
```

**Result**: Even though you validate against schema, you STILL create the Event struct from event.go!

---

## The Fundamental Question

**Can we generate deep links from `map[string]interface{}` instead of Event struct?**

### Current URL Generation (Must Check):

Let me trace what `GenerateURL(event)` does:

```go
// google_calendar.go:33-38
GenerateURL: func(event interface{}) (string, error) {
    e := event.(googlecalendar.Event)  // Type assertion to Event struct!
    return e.GenerateURL(), nil
}
```

**This calls**: `googlecalendar.Event.GenerateURL()`

### So The Question Becomes:

**Can `GenerateURL()` work with a map instead of a struct?**

Let's analyze:

```go
// Current (event.go):
func (e *Event) GenerateURL() string {
    // Uses struct fields: e.Title, e.StartTime, etc.
    return fmt.Sprintf("https://calendar.google.com/calendar/render?action=TEMPLATE&text=%s...",
        url.QueryEscape(e.Title),
        formatTime(e.StartTime),
        ...
    )
}
```

**What we need**:
```go
// New approach - NO event.go:
func GenerateURL(data map[string]interface{}) (string, error) {
    // Extract from map instead
    title, _ := data["title"].(string)
    start, _ := data["start"].(string)
    // ... etc

    return fmt.Sprintf("https://calendar.google.com/calendar/render?action=TEMPLATE&text=%s...",
        url.QueryEscape(title),
        start,  // Already a string from form!
        ...
    ), nil
}
```

---

## Deep Analysis: What Changes Are Actually Needed?

### Option A: Keep Current Pattern, Just Use Maps

**Changes Required**:

1. **Delete event.go files** ✅ Confirmed we can do this

2. **Change BuildEvent** ❌ Actually, just REMOVE it entirely!
   ```go
   // OLD:
   event, err := cfg.BuildEvent(r)

   // NEW:
   // formData already has the validated data!
   // Just use it directly!
   ```

3. **Change GenerateURL signature**:
   ```go
   // OLD:
   type CalendarURLGenerator func(event interface{}) (string, error)

   // NEW:
   type CalendarURLGenerator func(data map[string]interface{}) (string, error)
   ```

4. **Rewrite URL generators to work with maps**:
   ```go
   // pkg/google/calendar/generator.go (NEW FILE)
   func GenerateURL(data map[string]interface{}) (string, error) {
       title, ok := data["title"].(string)
       if !ok {
           return "", errors.New("title is required")
       }
       // ... etc
   }
   ```

5. **Handle datetime conversion in generator**:
   ```go
   // The form gives us: "2025-10-28T10:00" (string)
   // Google Calendar needs: "20251028T100000Z"

   start, ok := data["start"].(string)
   startTime, err := time.Parse("2006-01-02T15:04", start)
   // Format for Google...
   ```

### Option B: Simplify Even More - Map All The Way

**Realization**: We're already converting form → map → validate.
Why convert back to struct?!

**Simplified Flow**:
```
1. Form POST → map[string]interface{}
2. Validate map against schema.json
3. If valid → GenerateURL(map) directly
4. Done!
```

**No structs anywhere!**

---

## The Real Problem: Time Parsing

Looking at the code, I see why BuildEvent exists:

```go
// It parses datetime-local format to time.Time:
startTime, err := parseFormTime(r.FormValue("start"))
// "2025-10-28T10:00" → time.Time object

// Then Event.GenerateURL() formats it for Google:
// time.Time → "20251028T100000Z"
```

**Question**: Can we do this WITHOUT the Event struct?

**Answer**: YES! Just do it in the URL generator:

```go
func GenerateURL(data map[string]interface{}) (string, error) {
    startStr, _ := data["start"].(string)

    // Parse the datetime-local format
    startTime, err := time.Parse("2006-01-02T15:04", startStr)
    if err != nil {
        return "", err
    }

    // Format for Google Calendar
    formattedStart := startTime.UTC().Format("20060102T150405Z")

    // Build URL
    return fmt.Sprintf("...&dates=%s/...", formattedStart), nil
}
```

---

## Conclusion: YES, It Will Work!

### What Needs To Change:

1. ✅ **Delete event.go files** - Can be done!

2. ✅ **Remove BuildEvent callback** - Not needed!
   Already have validated `map[string]interface{}`

3. ✅ **Rewrite URL generators** - Take map instead of struct
   Handle datetime parsing inside generator

4. ✅ **Update CalendarConfig** - Remove BuildEvent, change GenerateURL signature

### New Structure:

```
pkg/google/calendar/
├── schema.json         # Defines structure + validation
├── uischema.json       # Defines layout
└── generator.go        # NEW: GenerateURL(map[string]interface{})

pkg/apple/calendar/
├── schema.json
├── uischema.json
└── generator.go        # NEW: GenerateICS(map[string]interface{})
```

### New Flow:

```go
// calendar_generic.go
func handleCalendarPost(w http.ResponseWriter, r *http.Request, cfg CalendarConfig) {
    // 1. Parse form → map
    formData := schema.FormDataToMap(r.Form)

    // 2. Validate against schema
    errors := schema.ValidateAgainstSchema(formData, jsonSchema)
    if len(errors) > 0 {
        // Show errors
        return
    }

    // 3. Generate URL directly from map!
    url, err := cfg.GenerateURL(formData)  // <-- formData is map!

    // 4. Done!
}
```

---

## Testing Impact

With this architecture, reflection-based testing becomes PERFECT:

```go
// cmd/testgen/main.go
func GenerateTestCases(schema *JSONSchema) []TestCase {
    var tests []TestCase

    for fieldName, prop := range schema.Properties {
        // Generate test data as map (NOT struct!)
        validData := map[string]interface{}{
            "title": "Meeting",
            "start": "2025-10-28T10:00",
            // ...
        }

        invalidData := map[string]interface{}{
            "title": "", // Empty (violates minLength)
            // ...
        }

        tests = append(tests, TestCase{
            Field: fieldName,
            ValidInput: validData,
            InvalidInput: invalidData,
        })
    }

    return tests
}
```

**Benefits**:
- Test data format matches actual runtime data (both maps!)
- No type conversion needed
- Schema drives everything (single source of truth)

---

## Final Answer: Will It Hold Together?

### ✅ YES - With These Changes:

1. Delete `event.go` files
2. Create `generator.go` in each calendar package
3. Rewrite generators to accept `map[string]interface{}`
4. Remove `BuildEvent` from `CalendarConfig`
5. Update `GenerateURL` signature to take map

### The Architecture Is Sound Because:

- ✅ Already validating maps against schema
- ✅ URL generation is just string building (doesn't need structs)
- ✅ Datetime parsing can happen in generator
- ✅ Testing reflects actual runtime flow
- ✅ Schema is true single source of truth

### Estimated Effort:

- Delete event.go: 2 minutes
- Create generator.go (Google): 30 minutes
- Create generator.go (Apple): 30 minutes
- Update calendar_generic.go: 15 minutes
- Test everything: 30 minutes

**Total**: ~2 hours of work

---

**Verdict**: ✅ Architecture is SOLID and will work perfectly!

