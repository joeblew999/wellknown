# Schema-Driven Architecture for Wellknown

**Date**: 2025-10-28
**Purpose**: Document the architectural decision to use JSON Schema as the single source of truth for forms, validation, and testing.

---

## Executive Summary

**Decision**: Eliminate Go struct definitions (`event.go` files) and use **JSON Schema + UI Schema** as the single source of truth for:
- Form generation
- Data validation
- Test generation
- Type definitions

**Inspired by**: [goPocJsonSchemaForm](../.src/goPocJsonSchemaForm) - A Go library that demonstrates pure schema-driven form handling.

---

## Current State (Before)

### Dual Source of Truth Problem

```
pkg/google/calendar/
├── event.go              # Go struct definition ❌ (duplicate!)
├── schema.json           # JSON Schema ✅
└── uischema.json         # UI Schema ✅

pkg/apple/calendar/
├── event.go              # Go struct definition ❌ (duplicate!)
├── schema.json           # JSON Schema ✅
└── uischema.json         # UI Schema ✅
```

**Problems:**
- ❌ Two sources of truth (Go struct + JSON Schema)
- ❌ Can get out of sync
- ❌ Changes require updating both
- ❌ Testing harder (which is correct?)

---

## Target State (After)

### Single Source of Truth: JSON Schema

```
pkg/google/calendar/
├── schema.json           # ✅ SINGLE SOURCE OF TRUTH
├── uischema.json         # ✅ Layout definition
└── (no event.go!)

pkg/apple/calendar/
├── schema.json           # ✅ SINGLE SOURCE OF TRUTH
├── uischema.json         # ✅ Layout definition
└── (no event.go!)

pkg/schema/
├── schema.go             # Schema loading/parsing
├── uischema.go           # UI Schema handling
├── validator.go          # Runtime validation
└── validator_test.go     # Validation tests
```

**Benefits:**
- ✅ Single source of truth
- ✅ Cannot get out of sync
- ✅ Schema changes = instant update everywhere
- ✅ Perfect for reflection-based testing

---

## Architecture: Schema-Driven Everything

### 1. Data Flow

```
JSON Schema (validation rules)
    +
UI Schema (layout, labels, placeholders)
    ↓
Go Server (pkg/schema/)
    ↓
Form Generation (HTML)
    ↓
Browser (user fills form)
    ↓
Validation (client + server)
    ↓
Deep Link Generation
```

### 2. No Go Structs - Use Generic Maps

**Instead of:**
```go
// ❌ OLD WAY - Requires event.go
type Event struct {
    Title       string
    StartTime   time.Time
    Location    string
}

event := Event{
    Title: "Meeting",
    StartTime: time.Now(),
}
```

**Use:**
```go
// ✅ NEW WAY - Schema-driven
data := map[string]interface{}{
    "title": "Meeting",
    "start": "2025-10-28T10:00",
    "location": "Conference Room A",
}

// Validate against schema.json
validator := schema.NewValidator("google/calendar/schema.json")
if err := validator.Validate(data); err != nil {
    // Handle validation errors
}

// Generate deep link from validated data
url := generator.GenerateURL(data, "google/calendar")
```

### 3. Form Generation

Forms are generated from JSON Schema + UI Schema:

```go
// Load schemas
jsonSchema := schema.LoadJSONSchema("google/calendar/schema.json")
uiSchema := schema.LoadUISchema("google/calendar/uischema.json")

// Generate HTML form
formHTML := schema.GenerateFormHTML(jsonSchema, uiSchema)

// Render in template
tmpl.Execute(w, PageData{
    SchemaFormHTML: formHTML,
})
```

**Result**: Change `schema.json` → Form automatically updates!

---

## How This Enables Reflection-Based Testing

### The Vision: Go Tool Reflects Over Schemas

**Step 1**: Go tool reads schemas
```go
// cmd/testgen/main.go
func main() {
    // Load schema.json + uischema.json
    schema := LoadSchema("google/calendar/schema.json")

    // Reflect over properties
    for _, property := range schema.Properties {
        // Extract validation rules
        if property.MinLength > 0 {
            // Generate test case: "Rejects value < minLength"
        }
        if property.MaxLength > 0 {
            // Generate test case: "Rejects value > maxLength"
        }
        if property.Pattern != "" {
            // Generate test case: "Rejects invalid pattern"
        }
    }

    // Output test-cases.json
    WriteTestCases("tests/.cache/test-cases.json")
}
```

**Step 2**: Minimal TypeScript test runner
```typescript
// tests/e2e/schema-runner.spec.ts (~50 lines)
import testCases from '../.cache/test-cases.json';

testCases.forEach(testCase => {
    test(testCase.description, async ({ page }) => {
        await page.fill(`[name="${testCase.field}"]`, testCase.input);

        if (testCase.shouldFail) {
            await expect(page.locator('.error-message')).toBeVisible();
        } else {
            await expect(page.locator('.error-message')).toHaveCount(0);
        }
    });
});
```

**Step 3**: Workflow
```bash
# 1. Edit schema.json
# 2. Regenerate test cases
go run cmd/testgen

# 3. Run tests (reads generated test-cases.json)
bun test

# Result: Tests automatically updated!
```

---

## What We Learned from goPocJsonSchemaForm

### Key Files Analyzed

1. **handler/schemaHandler.go**:
   - Uses `github.com/qri-io/jsonschema` for validation
   - `Validate(data *map[string]interface{})` - No Go structs!
   - Returns `map[string]string` of field errors

2. **handler/model.go**:
   - `FormData` with `Values` and `Errors`
   - `Control` binds to schema via "scope" (e.g., `#/properties/title`)
   - `Layout` types: VerticalLayout, HorizontalLayout, Control, Label

3. **Architecture**:
   ```go
   // Load schema once
   schemaHandler.LoadSchema("schema.json")

   // Validate any data
   data := map[string]interface{}{...}
   errors := schemaHandler.Validate(&data)

   // Bind UI controls to schema properties
   control.Scope = "#/properties/title"
   schemaHandler.AssignControlParameters(control)
   ```

### What We Adopted

Our `pkg/schema/` package is already inspired by goPocJsonSchemaForm:

```go
// pkg/schema/uischema.go (line 11-12)
// UISchema represents the UI layout configuration for a form
// Inspired by JSON Forms (jsonforms.io) and goPocJsonSchemaForm
```

**We already have:**
- ✅ `pkg/schema/schema.go` - JSON Schema handling
- ✅ `pkg/schema/uischema.go` - UI Schema handling
- ✅ `pkg/schema/validator.go` - Validation logic
- ✅ `pkg/schema/validator_test.go` - Tests

**We still have (TO REMOVE):**
- ❌ `pkg/google/calendar/event.go` - Delete this!
- ❌ `pkg/apple/calendar/event.go` - Delete this!

---

## Migration Plan

### Phase 1: Remove Event Structs ✅
- Delete `pkg/google/calendar/event.go`
- Delete `pkg/apple/calendar/event.go`
- Update server handlers to use `map[string]interface{}`

### Phase 2: Create Go Test Generator 🔄
```
cmd/testgen/
└── main.go              # Reflects over schemas, outputs test-cases.json
```

### Phase 3: Minimal TypeScript Test Runner 🔄
```
tests/e2e/
└── schema-runner.spec.ts  # Reads test-cases.json, runs via Playwright
```

### Phase 4: Workflow Integration 🔄
```
tests/package.json:
{
  "scripts": {
    "testgen": "go run ../../cmd/testgen",
    "test": "bun run testgen && bun run playwright test"
  }
}
```

---

## Benefits of This Approach

### 1. Single Source of Truth
- ✅ Schema.json defines everything
- ✅ Forms generated from schema
- ✅ Validation from schema
- ✅ Tests generated from schema

### 2. Always in Sync
- ✅ Change schema → Everything updates
- ✅ Cannot get out of sync
- ✅ No manual maintenance

### 3. Reflection-Based Testing
- ✅ Go reads schemas at build time
- ✅ Generates test cases automatically
- ✅ TypeScript is just a thin runner
- ✅ Minimal TypeScript code (~50 lines)

### 4. Flexibility
- ✅ Add field → Schema change only
- ✅ Change validation → Schema change only
- ✅ Update UI layout → UI Schema change only
- ✅ No Go recompilation needed

### 5. Type Safety (Where It Matters)
- ✅ Runtime validation against schema (safe!)
- ✅ Schema validator catches errors
- ✅ Tests auto-generated from schema
- ❌ No compile-time type safety (trade-off)

---

## Trade-offs

### What We Gain
- ✅ Single source of truth (schema.json)
- ✅ Always in sync
- ✅ Reflection-based testing
- ✅ Schema changes = instant updates everywhere

### What We Lose
- ❌ No compile-time type safety in Go
- ❌ No IDE autocomplete for event fields
- ❌ Runtime errors instead of compile errors
- ❌ Less "idiomatic Go" (Go loves structs!)

### Why It's Worth It
For a **web-form-driven application** where:
- All data comes from JSON forms
- Schema defines validation rules
- Testing needs to stay in sync
- Flexibility > compile-time safety

**Schema-driven is the right choice!**

---

## References

- **goPocJsonSchemaForm**: `.src/goPocJsonSchemaForm/`
- **JSON Schema**: https://json-schema.org/
- **JSON Forms**: https://jsonforms.io/
- **qri-io/jsonschema**: https://github.com/qri-io/jsonschema

---

## Questions & Answers

### Q: "Why do we even need event.go?"
**A**: We don't! goPocJsonSchemaForm doesn't use it. Schema is enough.

### Q: "What does map[string]interface{} + schema validation mean?"
**A**: Generic data container validated at runtime against JSON Schema. No Go structs needed.

### Q: "Should we do what goPocJsonSchemaForm does?"
**A**: We already are! Our `pkg/schema/` is inspired by it. Now we just need to remove the duplicate event.go files.

### Q: "How does this help testing?"
**A**: Go tool reflects over schema.json, generates test cases, minimal TypeScript runs them. Always in sync!

---

**Status**: Architecture documented, awaiting implementation approval.
