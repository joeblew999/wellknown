# pocketbase-gogen - Type-Safe Go Code Generator

**Purpose**: Generate type-safe Go accessor structures from Pocketbase collections

**Repo**: https://github.com/Snonky/pocketbase-gogen

---

## What It Does

Converts Pocketbase database schemas into type-safe Go structs with getters/setters, eliminating the need to work with raw PocketBase records directly.

**Benefits**:
- ✅ Type safety - compiler catches field errors
- ✅ Better IDE autocomplete
- ✅ Custom methods that work with typed data
- ✅ Reduces boilerplate code

---

## Installation

```bash
go install github.com/snonky/pocketbase-gogen@latest
```

---

## Usage

### Step 1: Generate Template from pb_data

```bash
pocketbase-gogen template ./pb_data ./pb/generated/template.go
```

This reads your Pocketbase database (`pb_data/`) and generates a template file with all collection schemas.

### Step 2: Generate Typed Proxies

```bash
pocketbase-gogen generate ./pb/generated/template.go ./pb/generated/proxies.go
```

Optional flags:
- `--utils` - Generate helper functions and type constraints
- `--hooks` - Generate type-safe event hooks

**Full command with all features**:
```bash
pocketbase-gogen generate ./pb/generated/template.go ./pb/generated/proxies.go --utils --hooks
```

---

## Generated Files

### `template.go`
Contains the schema definitions extracted from `pb_data/`:

```go
// Auto-generated from your Pocketbase collections
var Collections = []Collection{
    {
        Name: "google_tokens",
        Fields: []Field{
            {Name: "user_id", Type: "text"},
            {Name: "access_token", Type: "text"},
            // ...
        },
    },
}
```

### `proxies.go`
Type-safe accessor structs:

```go
type GoogleToken struct {
    record *models.Record
}

func (gt *GoogleToken) UserID() string {
    return gt.record.GetString("user_id")
}

func (gt *GoogleToken) SetUserID(value string) {
    gt.record.Set("user_id", value)
}

// ... getters and setters for all fields
```

### `utils.go` (with `--utils`)
Generic helper functions for working with proxies.

### `proxy_hooks.go` + `proxy_events.go` (with `--hooks`)
Type-safe event hooks per collection:

```go
func (gt *GoogleToken) OnCreate(fn func(e *GoogleTokenEvent)) {
    // Type-safe hook registration
}
```

---

## Example: Using Generated Code

**Before** (raw records):
```go
// Unsafe - no compile-time checks
record := models.NewRecord(collection)
record.Set("user_id", "abc123")         // Typo: "user_di" would fail at runtime
accessToken := record.GetString("access_token")
```

**After** (generated proxies):
```go
// Type-safe - compiler catches errors
token := NewGoogleToken(record)
token.SetUserID("abc123")               // Typo: SetUserDI() → compile error ✅
accessToken := token.AccessToken()      // Autocomplete works ✅
```

---

## Workflow

```bash
# 1. Make schema changes in Pocketbase Admin UI
# 2. Regenerate code
pocketbase-gogen template ./pb_data ./pb/generated/template.go
pocketbase-gogen generate ./pb/generated/template.go ./pb/generated/proxies.go --utils --hooks

# 3. Use typed proxies in your code
```

---

## When to Use

✅ **Use pocketbase-gogen when**:
- You have many collections with complex schemas
- You want type safety and autocomplete
- You're building a large Pocketbase application

❌ **Skip if**:
- You have 1-2 simple collections
- You prefer working with raw records
- Your schema changes very frequently

---

## Reference

See `.src/pocketbase-gogen/` for full source code and examples.

**Related**: [pb/README.md](../../pb/README.md) - Our Pocketbase integration

---

**Last Updated**: 2025-10-27
