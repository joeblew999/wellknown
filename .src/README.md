# .src/ - External Reference Code

This directory contains external source code repositories for reference and learning purposes. **These are NOT part of our codebase** and are git-ignored.

## Purpose

- Study implementation patterns from other projects
- Extract useful techniques without direct dependencies
- Keep reference code locally accessible for offline work
- Avoid polluting our codebase with external code

## Current References

### goPocJsonSchemaForm
- **Repo**: https://github.com/warlockxins/goPocJsonSchemaForm
- **Why**: Demonstrates dynamic form generation from JSON Schema + HTMX patterns
- **What we're learning**:
  - JSON Schema → HTML form generation
  - HTMX for progressive enhancement
  - Go template rendering patterns
  - Contact form with validation example

**Clone command**:
```bash
cd .src
git clone https://github.com/warlockxins/goPocJsonSchemaForm.git
```

### Presentator
- **Repo**: https://github.com/presentator/presentator
- **Why**: Real-world Pocketbase extension as importable library
- **What we're learning**:
  - How to structure PB code as reusable library
  - `presentator.New()` pattern wrapping `pocketbase.New()`
  - `base/main.go` as demo/test server
  - Root package has all business logic
  - Hooks registration: `bindAppHooks(pr)`
  - Option management via PB Store

**Clone command**:
```bash
cd .src
git clone https://github.com/presentator/presentator.git
```

**Key Pattern**:
```go
// Root: presentator.go (library)
package presentator

type Presentator struct {
    *pocketbase.PocketBase
}

func New() *Presentator {
    pr := &Presentator{pocketbase.New()}
    bindAppHooks(pr)  // Register routes, hooks, etc
    return pr
}

// base/main.go (demo server)
package main

func main() {
    app := presentator.New()
    app.Start()
}
```

### pocketbase
- **Repo**: https://github.com/pocketbase/pocketbase
- **Why**: Official PocketBase source code - understand core record system and proxy patterns
- **What we're learning**:
  - How `core.Record` and `core.BaseRecordProxy` work
  - What methods exist that could cause shadowing issues
  - Official patterns for type-safe record handling
  - Auth collection special handling (email, password, etc.)
  - How record proxies are meant to be used
  - Alternative approaches to code generation

**Clone command**:
```bash
cd .src
git clone https://github.com/pocketbase/pocketbase.git
```

**Key files to study**:
```bash
# Core record system
.src/pocketbase/core/record.go
.src/pocketbase/core/record_proxy.go

# Daos and record operations
.src/pocketbase/daos/record.go

# Auth collection handling
.src/pocketbase/core/base_record_proxy.go
```

### pocketbase-gogen
- **Repo**: https://github.com/Snonky/pocketbase-gogen
- **Why**: Generate type-safe Go code from Pocketbase collections
- **What we're learning**:
  - Convert PB schemas to Go structs automatically
  - Type-safe DAOs/Proxies instead of raw records
  - Custom methods with typed data
  - Template generation from `pb_data/`
  - Proxy generation with getters/setters
  - Optional utils and hooks generation
  - **Limitation**: Generated code can shadow core.Record methods (email, etc.)

**Clone command**:
```bash
cd .src
git clone https://github.com/Snonky/pocketbase-gogen.git
```

**Install as tool**:
```bash
go install github.com/snonky/pocketbase-gogen@latest
```

**Usage**:
```bash
# Step 1: Generate template from pb_data
pocketbase-gogen template ./path/to/pb_data ./yourmodule/pbschema/template.go

# Step 2: Generate proxies from template
pocketbase-gogen generate ./yourmodule/pbschema/template.go ./yourmodule/generated/proxies.go --utils --hooks
```

**Known Issues**:
- Generated templates may contain fields that shadow `core.Record` methods
- Requires manual editing of template between generation steps
- See `pb/README.md` for workflow details

## Managing Reference Repositories

Use the included Makefile to manage all reference repositories:

```bash
# Clone all reference repos at once
make clone-all

# Or clone individual repos
make clone-pocketbase
make clone-gogen
make clone-presentator
make clone-jsonschema
make clone-goPocJson

# Update all repos (git pull)
make update-all

# List what's currently cloned
make list

# Show git status for all repos
make status

# Remove all cloned repos (WARNING: destructive!)
make clean
```

## Usage Pattern

1. **Clone**: Use `make clone-all` or clone individual repos
2. **Study**: Read code, run examples, understand patterns
3. **Extract**: Adapt useful patterns to our codebase
4. **Update**: Run `make update-all` periodically to stay current
5. **Never commit**: Cloned repos are gitignored - only README.md and Makefile are tracked

## Notes

- This is a **learning directory**, not a dependency directory
- Code here is from external authors with their own licenses
- Always check licenses before adapting patterns
- Keep our codebase clean - import ideas, not code directly
