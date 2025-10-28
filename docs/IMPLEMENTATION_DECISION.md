# Implementation Decision: Use santhosh-tekuri/jsonschema

**Date**: 2025-10-28
**Decision**: Use `github.com/santhosh-tekuri/jsonschema/v5` for all JSON Schema operations
**Status**: ✅ APPROVED

---

## Summary

**Replace custom stdlib-only validator with `santhosh-tekuri/jsonschema` for:**
- ✅ Runtime validation (production)
- ✅ Test case generation (development)
- ✅ Schema reflection (both)

---

## Why This Library?

### 1. Minimal Dependencies
```
go.mod:
require github.com/santhosh-tekuri/jsonschema/v5 v5.x

# Only 1 dependency (and it's only for testing):
require github.com/dlclark/regexp2 v1.10.0 // for testing
```

**Cost**: 1 tiny dependency (~100KB)
**Benefit**: Full JSON Schema spec compliance + perfect reflection

### 2. Perfect for Reflection

The `Schema` struct exposes ALL validation properties as public fields:

```go
type Schema struct {
    // Object validation
    Properties    map[string]*Schema  // ← Iterate all fields!
    Required      []string            // ← Required fields!

    // String validation
    MinLength     *int                // ← Generate test: too short
    MaxLength     *int                // ← Generate test: too long
    Pattern       Regexp              // ← Generate test: invalid pattern

    // Number validation
    Minimum       *big.Rat            // ← Generate test: below min
    Maximum       *big.Rat            // ← Generate test: above max

    // Enum validation
    Enum          *Enum               // ← Generate test: invalid enum

    // ... and many more!
}
```

**Result**: Test generation is TRIVIAL!

### 3. Full JSON Schema Spec Support

Supports all drafts:
- ✅ Draft 4
- ✅ Draft 6
- ✅ Draft 7
- ✅ Draft 2019-09
- ✅ Draft 2020-12

Your schemas use Draft 7 - fully supported!

### 4. Fast Performance

- Pre-compiles schemas for speed
- Uses `*big.Rat` for precise number validation
- Efficient validation algorithm

### 5. Battle-Tested

- 4.5k+ GitHub stars
- Used in production by many projects
- Active maintenance
- Pure Go (no C dependencies)

---

## Comparison with Current Custom Validator

### Current Approach (pkg/schema/validator.go)

```go
// ❌ Custom stdlib-only validator
// ~200 lines of hand-written code
// No external dependencies
// Basic validation only
```

**Problems:**
- You maintain all validation logic
- May miss edge cases
- Limited JSON Schema spec support
- Hard to extend

### New Approach (santhosh-tekuri/jsonschema)

```go
// ✅ Industry-standard library
// 1 tiny dependency (regexp2, testing only)
// Full JSON Schema spec
// Perfect for reflection
```

**Benefits:**
- Battle-tested validation
- Full spec compliance
- Easy to reflect for test generation
- Active maintenance

---

## Implementation Plan

### Phase 1: Replace Validator ✅ READY

**Current file**: `pkg/schema/validator.go` (~200 lines custom code)

**New implementation**:
```go
// pkg/schema/validator.go
package schema

import "github.com/santhosh-tekuri/jsonschema/v5"

type Validator struct {
    compiler *jsonschema.Compiler
    schemas  map[string]*jsonschema.Schema
}

func NewValidator() *Validator {
    return &Validator{
        compiler: jsonschema.NewCompiler(),
        schemas:  make(map[string]*jsonschema.Schema),
    }
}

func (v *Validator) LoadSchema(path string) (*jsonschema.Schema, error) {
    if cached, ok := v.schemas[path]; ok {
        return cached, nil
    }

    schema, err := v.compiler.Compile(path)
    if err != nil {
        return nil, err
    }

    v.schemas[path] = schema
    return schema, nil
}

func (v *Validator) Validate(data map[string]interface{}, schema *jsonschema.Schema) error {
    return schema.Validate(data)
}
```

**Estimated effort**: 1 hour

### Phase 2: Add Test Generator ✅ READY

**New file**: `pkg/schema/testgen.go`

```go
// pkg/schema/testgen.go
package schema

import "github.com/santhosh-tekuri/jsonschema/v5"

type TestCase struct {
    Description string                 `json:"description"`
    Field       string                 `json:"field"`
    Data        map[string]interface{} `json:"data"`
    ShouldPass  bool                   `json:"shouldPass"`
    Reason      string                 `json:"reason"`
}

func GenerateTestCases(schema *jsonschema.Schema) []TestCase {
    var cases []TestCase

    // Iterate over all properties
    for propName, propSchema := range schema.Properties {
        // Check if required
        isRequired := contains(schema.Required, propName)

        if isRequired {
            cases = append(cases, TestCase{
                Description: fmt.Sprintf("%s is required", propName),
                Field:       propName,
                Data:        map[string]interface{}{}, // Empty
                ShouldPass:  false,
                Reason:      "required field missing",
            })
        }

        // String validation
        if propSchema.MinLength != nil {
            cases = append(cases, TestCase{
                Description: fmt.Sprintf("rejects %s shorter than %d", propName, *propSchema.MinLength),
                Field:       propName,
                Data: map[string]interface{}{
                    propName: strings.Repeat("x", *propSchema.MinLength - 1),
                },
                ShouldPass: false,
                Reason:     fmt.Sprintf("minLength: %d", *propSchema.MinLength),
            })
        }

        if propSchema.MaxLength != nil {
            cases = append(cases, TestCase{
                Description: fmt.Sprintf("rejects %s longer than %d", propName, *propSchema.MaxLength),
                Field:       propName,
                Data: map[string]interface{}{
                    propName: strings.Repeat("x", *propSchema.MaxLength + 1),
                },
                ShouldPass: false,
                Reason:     fmt.Sprintf("maxLength: %d", *propSchema.MaxLength),
            })
        }

        // Pattern validation
        if propSchema.Pattern != "" {
            // Generate test data that violates pattern
            // (pattern-specific logic here)
        }

        // Number validation
        if propSchema.Minimum != nil {
            // Generate test: value below minimum
        }

        if propSchema.Maximum != nil {
            // Generate test: value above maximum
        }

        // Enum validation
        if propSchema.Enum != nil {
            // Generate test: invalid enum value
        }
    }

    return cases
}
```

**Estimated effort**: 2 hours

### Phase 3: CLI Tool ✅ READY

**New file**: `cmd/testgen/main.go`

```go
package main

import (
    "encoding/json"
    "os"
    "github.com/joeblew999/wellknown/pkg/schema"
    "github.com/santhosh-tekuri/jsonschema/v5"
)

func main() {
    compiler := jsonschema.NewCompiler()

    // Load Google Calendar schema
    googleSchema, _ := compiler.Compile("../../pkg/google/calendar/schema.json")
    googleTests := schema.GenerateTestCases(googleSchema)

    // Load Apple Calendar schema
    appleSchema, _ := compiler.Compile("../../pkg/apple/calendar/schema.json")
    appleTests := schema.GenerateTestCases(appleSchema)

    // Combine all test cases
    allTests := map[string][]schema.TestCase{
        "google-calendar": googleTests,
        "apple-calendar":  appleTests,
    }

    // Write to JSON
    data, _ := json.MarshalIndent(allTests, "", "  ")
    os.WriteFile("../../tests/.cache/test-cases.json", data, 0644)

    fmt.Printf("✅ Generated %d test cases\n", len(googleTests) + len(appleTests))
}
```

**Estimated effort**: 30 minutes

### Phase 4: TypeScript Test Runner ✅ READY

**Update file**: `tests/e2e/schema-reflection.spec.ts`

```typescript
// Load generated test cases
import testCases from '../.cache/test-cases.json';

// Run tests for Google Calendar
testCases['google-calendar'].forEach(testCase => {
    test(testCase.description, async ({ page }) => {
        await page.goto('/google/calendar');

        // Fill form with test data
        for (const [field, value] of Object.entries(testCase.data)) {
            await page.fill(`[name="${field}"]`, String(value));
        }

        // Submit form
        await page.click('button[type="submit"]');

        if (testCase.shouldPass) {
            await expect(page.locator('.error-message')).toHaveCount(0);
        } else {
            await expect(page.locator('.error-message')).toBeVisible();
        }
    });
});

// Same for Apple Calendar...
```

**Estimated effort**: 30 minutes

---

## Total Implementation Effort

1. Replace validator: **1 hour**
2. Add test generator: **2 hours**
3. Create CLI tool: **30 minutes**
4. Update TypeScript runner: **30 minutes**

**Total**: ~4 hours of work

---

## Migration Steps

### Step 1: Add Dependency
```bash
go get github.com/santhosh-tekuri/jsonschema/v5
```

### Step 2: Replace Validator (Keep Existing Tests Passing!)
- Rewrite `pkg/schema/validator.go` to use jsonschema library
- Update server handlers to use new validator
- **CRITICAL: Ensure existing functionality works:**
  - ✅ Google Calendar form loads
  - ✅ Google Calendar validation works
  - ✅ Google Calendar URL generation works
  - ✅ Apple Calendar form loads
  - ✅ Apple Calendar validation works
  - ✅ Apple Calendar ICS generation works

**Testing checklist**:
```bash
# 1. Start server
go run ./cmd/server

# 2. Test Google Calendar manually:
#    - Go to http://localhost:8080/google/calendar
#    - Fill form with valid data
#    - Verify URL generates correctly
#    - Test with invalid data (minLength violation)
#    - Verify validation error shows

# 3. Test Apple Calendar manually:
#    - Go to http://localhost:8080/apple/calendar
#    - Fill form with valid data
#    - Verify ICS download works
#    - Test with invalid data
#    - Verify validation error shows

# 4. Run existing Playwright tests:
cd tests
bun run playwright test wizard-core --grep-invert "Delete"
# Should still pass 13/13 tests!
```

### Step 3: Create URL/ICS Generators (No event.go!)

**CRITICAL CHANGE**: Rewrite generators to work with `map[string]interface{}`

**Create**: `pkg/google/calendar/generator.go`
```go
package googlecalendar

import (
    "fmt"
    "net/url"
    "time"
)

// GenerateURL creates a Google Calendar URL from form data (NO struct needed!)
func GenerateURL(data map[string]interface{}) (string, error) {
    // Extract fields from map
    title, _ := data["title"].(string)
    startStr, _ := data["start"].(string)
    endStr, _ := data["end"].(string)
    location, _ := data["location"].(string)
    description, _ := data["description"].(string)

    // Parse datetime-local format
    startTime, err := time.Parse("2006-01-02T15:04", startStr)
    if err != nil {
        return "", err
    }
    endTime, err := time.Parse("2006-01-02T15:04", endStr)
    if err != nil {
        return "", err
    }

    // Format for Google Calendar (UTC)
    formattedStart := startTime.UTC().Format("20060102T150405Z")
    formattedEnd := endTime.UTC().Format("20060102T150405Z")

    // Build URL
    baseURL := "https://calendar.google.com/calendar/render"
    params := url.Values{}
    params.Add("action", "TEMPLATE")
    params.Add("text", title)
    params.Add("dates", formattedStart+"/"+formattedEnd)
    if location != "" {
        params.Add("location", location)
    }
    if description != "" {
        params.Add("details", description)
    }

    return fmt.Sprintf("%s?%s", baseURL, params.Encode()), nil
}
```

**Create**: `pkg/apple/calendar/generator.go`
```go
package applecalendar

import (
    "fmt"
    "time"
)

// GenerateICS creates ICS file content from form data (NO struct needed!)
func GenerateICS(data map[string]interface{}) (string, error) {
    // Extract fields from map
    title, _ := data["title"].(string)
    startStr, _ := data["start"].(string)
    endStr, _ := data["end"].(string)
    location, _ := data["location"].(string)
    description, _ := data["description"].(string)

    // Parse datetime-local format
    startTime, err := time.Parse("2006-01-02T15:04", startStr)
    if err != nil {
        return "", err
    }
    endTime, err := time.Parse("2006-01-02T15:04", endStr)
    if err != nil {
        return "", err
    }

    // Format for ICS (UTC)
    formattedStart := startTime.UTC().Format("20060102T150405Z")
    formattedEnd := endTime.UTC().Format("20060102T150405Z")

    // Build ICS content
    ics := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//wellknown//EN
BEGIN:VEVENT
DTSTART:%s
DTEND:%s
SUMMARY:%s
LOCATION:%s
DESCRIPTION:%s
END:VEVENT
END:VCALENDAR`, formattedStart, formattedEnd, title, location, description)

    return ics, nil
}
```

### Step 4: Update CalendarConfig (Remove BuildEvent!)

**Modify**: `pkg/server/calendar_generic.go`

```go
// OLD CalendarConfig:
type CalendarConfig struct {
    Platform     string
    AppType      string
    BuildEvent   CalendarEventBuilder  // ❌ DELETE THIS!
    GenerateURL  CalendarURLGenerator  // ❌ Change signature!
    SuccessLabel string
}

// NEW CalendarConfig:
type CalendarConfig struct {
    Platform     string
    AppType      string
    // BuildEvent removed! ✅
    GenerateURL  func(data map[string]interface{}) (string, error)  // ✅ Takes map!
    SuccessLabel string
}
```

**Update handler**:
```go
func handleCalendarPost(w http.ResponseWriter, r *http.Request, cfg CalendarConfig) {
    // ... (validation code stays same) ...

    // ❌ OLD: Build typed event struct
    // event, err := cfg.BuildEvent(r)

    // ✅ NEW: Use validated map directly!
    generatedURL, err := cfg.GenerateURL(formData)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // ... (render response) ...
}
```

### Step 5: Delete event.go Files! 🎉
```bash
rm pkg/google/calendar/event.go
rm pkg/apple/calendar/event.go
```

**Verify everything still works**:
```bash
# Run server
go run ./cmd/server

# Manual testing (same as Step 2)

# Run Playwright tests
cd tests && bun run playwright test wizard-core
# Should STILL pass 13/13!
```

### Step 6: Add Test Generator
- Create `pkg/schema/testgen.go`
- Implement `GenerateTestCases()`
- Test with Google Calendar schema

### Step 7: Create CLI Tool
- Create `cmd/testgen/main.go`
- Generate test-cases.json
- Verify JSON output format

### Step 8: Update TypeScript
- Modify `tests/e2e/schema-reflection.spec.ts`
- Read test-cases.json
- Run generated tests

### Step 9: Integrate Workflow
```json
// tests/package.json
{
  "scripts": {
    "testgen": "go run ../../cmd/testgen",
    "test": "bun run testgen && bun run playwright test"
  }
}
```

---

## Benefits Summary

### For Development
- ✅ Change schema.json → Tests auto-update
- ✅ No manual test maintenance
- ✅ Always in sync

### For Production
- ✅ Full JSON Schema validation
- ✅ Battle-tested library
- ✅ Fast performance
- ✅ Minimal dependency (1 package, testing only)

### For Reflection-Based Testing
- ✅ Perfect struct for reflection
- ✅ All properties public and accessible
- ✅ Easy to iterate and generate tests
- ✅ Comprehensive coverage

---

## Risks & Mitigations

### Risk 1: Dependency on External Package
**Mitigation**: Library is widely used, actively maintained, and has only 1 dependency

### Risk 2: Breaking Changes in Library
**Mitigation**: Use versioned import (`/v5`), pin to specific version in go.mod

### Risk 3: Performance Impact
**Mitigation**: Library is designed for performance (pre-compiles schemas), likely faster than custom validator

### Risk 4: Learning Curve
**Mitigation**: Well-documented API, many examples, similar to what we already have

---

## Decision

✅ **APPROVED**: Use `santhosh-tekuri/jsonschema/v5` for all JSON Schema operations

**Rationale**:
- Minimal cost (1 dependency)
- Maximum benefit (full spec + perfect reflection)
- Simplifies codebase (don't maintain custom validator)
- Enables reflection-based testing (the original goal!)

**Next Steps**:
1. Add dependency to go.mod
2. Implement Phase 1-5 (estimated 4 hours)
3. Test thoroughly
4. Document new workflow

---

**Approved by**: [Your approval]
**Date**: 2025-10-28
