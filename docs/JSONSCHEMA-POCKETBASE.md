# JSON Schema + Pocketbase Integration

**Vision**: Use JSON Schema for frontend + Pocketbase for backend = Full-stack type safety

**Date**: 2025-10-27

---

## The Problem

Currently we have:
- **JSON Schema**: Great for forms, validation, UI generation (frontend)
- **Pocketbase**: Great for database, auth, CRUD APIs (backend)
- **Gap**: They don't talk to each other!

---

## The Vision

```
┌──────────────────────────────────────────────────────────────┐
│                  Single Source of Truth                       │
│                     (JSON Schema)                            │
└──────────────────────┬───────────────────────────────────────┘
                       │
         ┌─────────────┴──────────────┐
         ▼                            ▼
┌─────────────────┐          ┌─────────────────┐
│    Frontend     │          │     Backend     │
│  (pkg/schema)   │          │   (Pocketbase)  │
├─────────────────┤          ├─────────────────┤
│ ✅ Form HTML     │          │ ✅ DB Schema     │
│ ✅ Validation    │          │ ✅ Collections   │
│ ✅ UI rendering  │          │ ✅ API rules     │
│ ✅ Error display │          │ ✅ Validation    │
└─────────────────┘          └─────────────────┘
```

**Goal**: Write JSON Schema once, use it everywhere!

---

## Use Cases

### 1. Calendar Event Form

**JSON Schema** (already have!):
```json
{
  "type": "object",
  "properties": {
    "title": {"type": "string", "minLength": 1},
    "start": {"type": "string", "format": "date-time"},
    "end": {"type": "string", "format": "date-time"}
  },
  "required": ["title", "start", "end"]
}
```

**Frontend** (already working!):
```go
// pkg/schema/validator.go
formHTML := uiSchema.GenerateFormHTMLWithData(jsonSchema, formData, errors)
```

**Backend** (NEW - Pocketbase auto-creation!):
```go
// pb/schema_sync.go
collection := createCollectionFromSchema(jsonSchema)
// Creates PB collection with:
// - title: text field (required)
// - start: date field (required)
// - end: date field (required)
```

---

## Architecture

### Current State

```
pkg/
├── schema/                  # JSON Schema utilities
│   ├── schema.go           # Parse JSON Schema
│   ├── uischema.go         # Parse UI Schema
│   └── validator.go        # Validate data
│
└── google/calendar/
    ├── schema.json         # Event schema
    └── uischema.json       # Event UI layout
```

### Future State (With PB Integration)

```
pkg/
├── schema/                  # JSON Schema utilities (enhanced)
│   ├── schema.go           # Parse JSON Schema
│   ├── uischema.go         # Parse UI Schema
│   ├── validator.go        # Validate data
│   └── pb_adapter.go       # NEW: Convert to PB collections
│
pb/                          # Pocketbase extension
├── wellknown.go
├── schema_sync.go          # NEW: Sync schemas to PB
└── collections.go          # NEW: Auto-create collections
```

---

## Implementation Pattern

### Option A: Manual Sync (Simpler, Start Here)

**Developer workflow**:
1. Define JSON Schema (e.g., `calendar_event.schema.json`)
2. Run tool: `pb-schema-sync calendar_event.schema.json`
3. Tool creates PB collection matching schema
4. Use same schema for frontend forms

**Benefits**:
- ✅ Simple to understand
- ✅ Explicit control
- ✅ Easy to debug

**Drawbacks**:
- ⚠️ Manual sync step
- ⚠️ Schema drift possible

### Option B: Auto-Sync (Advanced, Future)

**Automatic on app start**:
```go
func bindAppHooks(wk *Wellknown) {
    wk.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        // Auto-sync all schemas in pkg/google/calendar/
        return syncSchemasToCollections(wk)
    })
}
```

**Benefits**:
- ✅ Always in sync
- ✅ Zero manual steps
- ✅ Single source of truth

**Drawbacks**:
- ⚠️ Complex
- ⚠️ Migration challenges
- ⚠️ Production risk

---

## Schema → Pocketbase Mapping

### JSON Schema Types → PB Field Types

| JSON Schema | Pocketbase Field |
|-------------|------------------|
| `string` | `text` |
| `string` (email) | `email` |
| `string` (url) | `url` |
| `string` (date-time) | `date` |
| `number` | `number` |
| `boolean` | `bool` |
| `array` | `json` or `relation` |
| `object` | `json` |
| `enum` | `select` (single/multiple) |

### Validation Rules → PB Rules

| JSON Schema | Pocketbase |
|-------------|-----------|
| `required` | Field required |
| `minLength` | Min length rule |
| `maxLength` | Max length rule |
| `minimum` | Min value rule |
| `maximum` | Max value rule |
| `pattern` | Regex validation |
| `enum` | Select options |

---

## Example: Calendar Event

### JSON Schema
```json
{
  "type": "object",
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1,
      "maxLength": 200
    },
    "start": {
      "type": "string",
      "format": "date-time"
    },
    "end": {
      "type": "string",
      "format": "date-time"
    },
    "location": {
      "type": "string",
      "maxLength": 500
    }
  },
  "required": ["title", "start", "end"]
}
```

### Auto-Generated PB Collection
```go
collection := &models.Collection{
    Name:   "calendar_events",
    Type:   models.CollectionTypeBase,
    Schema: []*schema.SchemaField{
        {
            Name:     "title",
            Type:     schema.FieldTypeText,
            Required: true,
            Options: &schema.TextOptions{
                Min: 1,
                Max: 200,
            },
        },
        {
            Name:     "start",
            Type:     schema.FieldTypeDate,
            Required: true,
        },
        {
            Name:     "end",
            Type:     schema.FieldTypeDate,
            Required: true,
        },
        {
            Name:     "location",
            Type:     schema.FieldTypeText,
            Options: &schema.TextOptions{
                Max: 500,
            },
        },
    },
}
```

---

## Benefits

### 1. Single Source of Truth
- ✅ Define schema once in JSON
- ✅ Frontend forms auto-generate
- ✅ Backend database auto-creates
- ✅ Validation rules consistent

### 2. Type Safety Across Stack
- ✅ Frontend validates before submit
- ✅ Backend validates on save
- ✅ Same rules, no drift

### 3. Rapid Development
- ✅ Change schema → both sides update
- ✅ No manual DB migration code
- ✅ No manual form code

### 4. Documentation
- ✅ Schema documents API
- ✅ Schema documents UI
- ✅ Self-documenting system

---

## Implementation Phases

### Phase 1: Schema Adapter (pkg/schema/)
```go
// pkg/schema/pb_adapter.go
func SchemaToPocketbaseCollection(jsonSchema *Schema) *pb.Collection {
    // Convert JSON Schema to PB collection definition
}
```

### Phase 2: CLI Tool (tools/pb-schema-sync/)
```bash
# Manual sync tool
go run tools/pb-schema-sync/main.go sync calendar_event.schema.json

# Output:
# ✅ Created collection: calendar_events
# ✅ Added field: title (text, required)
# ✅ Added field: start (date, required)
# ✅ Added field: end (date, required)
```

### Phase 3: PB Hook (pb/schema_sync.go)
```go
// Auto-sync on app start (optional)
func syncSchemasToCollections(wk *Wellknown) error {
    schemas := loadAllSchemas()
    for _, schema := range schemas {
        syncSchemaToCollection(wk, schema)
    }
}
```

### Phase 4: API Routes (pb/api.go)
```go
// CRUD routes use schema validation
e.Router.POST("/api/events", func(c echo.Context) error {
    // Parse JSON Schema
    schema := loadSchema("calendar_event.schema.json")

    // Validate request
    errors := ValidateAgainstSchema(c.Request().Body, schema)
    if len(errors) > 0 {
        return c.JSON(400, errors)
    }

    // Save to PB (already validated!)
    record := pb.SaveRecord("calendar_events", data)
})
```

---

## Example Workflow

### Developer Experience

1. **Define Schema**:
```bash
vi pkg/google/calendar/schema.json
# Edit event schema
```

2. **Sync to PB** (manual or auto):
```bash
make pb-schema-sync
# Or auto on app start
```

3. **Frontend gets form**:
```go
formHTML := uiSchema.GenerateFormHTMLWithData(jsonSchema, nil, nil)
// Renders form with validation
```

4. **User submits form**:
```go
// Frontend: Validates with JSON Schema
errors := ValidateAgainstSchema(formData, jsonSchema)

// Backend: Saves to PB (schema already matches!)
pb.SaveRecord("calendar_events", formData)
```

5. **Both sides in sync!** ✅

---

## Challenges

### 1. Schema Evolution
- **Problem**: Schema changes → DB migration?
- **Solution**: Tool generates migration code

### 2. Complex Types
- **Problem**: Nested objects, arrays
- **Solution**: Map to PB JSON fields or relations

### 3. Existing Data
- **Problem**: Schema change breaks existing records
- **Solution**: Migration + validation fallback

---

## Tools to Build

### `tools/pb-schema-sync/` - CLI Tool
```bash
# Sync single schema
pb-schema-sync sync calendar_event.schema.json

# Sync all schemas
pb-schema-sync sync-all pkg/**/schema.json

# Dry run (show what would change)
pb-schema-sync sync --dry-run calendar_event.schema.json
```

### `pkg/schema/pb_adapter.go` - Library
```go
// Convert JSON Schema → PB Collection
func SchemaToPocketbaseCollection(*Schema) *pb.Collection

// Validate data against schema (already have!)
func ValidateAgainstSchema(data, schema) ValidationErrors

// Compare schemas (detect drift)
func CompareSchemas(jsonSchema, pbCollection) Diff
```

---

## Success Criteria

✅ **Single Schema Definition**: Write once, use everywhere
✅ **Automatic Sync**: Schema changes propagate to DB
✅ **Consistent Validation**: Same rules frontend + backend
✅ **Self-Documenting**: Schema documents API + UI
✅ **Rapid Development**: Change schema, system updates

---

## Related Documents

- [POCKETBASE-ARCHITECTURE.md](POCKETBASE-ARCHITECTURE.md) - PB integration pattern
- [pkg/schema/README.md](../pkg/schema/README.md) - JSON Schema utilities

**Last Updated**: 2025-10-27
