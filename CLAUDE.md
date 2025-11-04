# Claude's Notes - wellknown Project

**Purpose**: Quick reference for AI agents working on this codebase

---

## 🚨 Critical Rules

1. **Module name**: `github.com/joeblew999/wellknown` (⚠️ `999` not `99`!)
2. **File Path**: `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown`
3. **Never use browser dialogs** (`alert()`, `confirm()`, `prompt()`) - use toast notifications
4. **Never commit without being asked** - user must explicitly request commits
- You MUST ONLY use the Makefile to run the code !!! This ensures the Makefile is the source of truth for how to run the system, and forced you to keep it valid with the code. You MUST adapt the Makefile if you need to and not just call into the code directly.
5. MUST use the .src folder is for reference code, because its much quicker than Web Searches. Maintain the Makefile in it too.


---

## What This Project Does

Universal Go library for generating deep links to Google/Apple calendar apps.

**Key Principle**: Deterministic (same input → same output every time)

---

## Architecture (DRY Pattern)

### Shared Packages
```
pkg/
├── calendar/         # Shared field constants & interfaces
├── testgen/          # Test generation library
├── schema/           # File name constants (schema.json, etc.)
├── google/calendar/  # Google Calendar (imports pkg/calendar)
└── apple/calendar/   # Apple Calendar (imports pkg/calendar)

cmd/
└── testdata-gen/     # Thin CLI wrapper for pkg/testgen
```

### Key Pattern
- **Single source of truth**: Field names in `pkg/calendar/fields.go`
- **Re-exports**: Platform packages import & re-export for backwards compat
- **Registry-based**: Easy to add new platforms

---

## Platform Details

### Google Calendar
- **Uses web URLs**: `https://calendar.google.com/calendar/render?action=TEMPLATE&...`
- **Why**: Native `googlecalendar://` doesn't support event parameters
- **Mobile behavior**: Browser opens, then OS offers to open in app

### Apple Calendar
- **Uses HTTP-served .ics files**: `/apple/calendar/download?event=<base64>`
- **Why**: Safari rejects `data:text/calendar` URIs as invalid
- **Format**: RFC 5545 iCalendar format
- **Supports**: Recurrence, attendees, reminders, etc.

---

## Development Workflow

### Hot Reload (Development)
```bash
make dev          # Starts Air with hot-reload
# Watches .go, .html files in pkg/ and cmd/
# Auto-rebuilds and restarts on changes
```

### Generate Test Data
```bash
go run ./cmd/testdata-gen
# Or: make gen-testdata
```

### Run Tests
```bash
go test ./...                    # All Go tests
bun test tests/e2e/             # Playwright tests
```

---

## Adding a New Platform

1. **Create directory**: `pkg/{platform}/{appType}/`
2. **Add files**:
   - `calendar.go` - Import `pkg/calendar`, implement generator
   - `schema.json` - JSON Schema validation
   - `uischema.json` - Form UI definition
   - `data-examples.json` - Test examples
3. **Register in testgen**: Add to registry in `pkg/testgen/generator.go`
4. **Register in server**: Add route in `pkg/server/routes.go`

That's it! Test generation and Playwright tests work automatically.

---

## UI/UX Rules

### ❌ NEVER Use
- `alert()` - Blocks UI, untestable
- `confirm()` - Slow tests, bad UX
- `prompt()` - Not mobile-friendly

### ✅ ALWAYS Use
- **Toast notifications** - Auto-dismiss, non-blocking
- **Inline errors** - Show below form fields
- **Modal dialogs** - For complex confirmations (rare)

---

## Testing Architecture

### Test Data Flow
```
pkg/{platform}/{appType}/data-examples.json
    ↓
cmd/testdata-gen (runs ACTUAL Go generators)
    ↓
tests/e2e/generated/{platform}-{apptype}-tests.json
    ↓
Playwright (platform-generic.spec.ts)
```

**Key**: Playwright validates against Go-generated expectations, not hardcoded values!

### Current Tests
- `tests/e2e/platform-generic.spec.ts` - Generic suite for ALL platforms
- `tests/e2e/wizard-core.spec.ts` - GCP OAuth setup wizard
- Go unit tests in each `pkg/` directory

---

## Key Files

### Must Know
- `pkg/calendar/fields.go` - Shared field constants
- `pkg/schema/const.go` - File name constants
- `pkg/testgen/generator.go` - Test generation logic
- `ARCHITECTURE.md` - Detailed architecture docs
- `REFACTORING_COMPLETE.md` - Recent refactoring summary

### Server
- `pkg/server/server.go` - Server struct (owns all deps, zero globals)
- `pkg/server/navigation.go` - Service registry
- `pkg/server/calendar_generic.go` - Schema-driven handler
- `pkg/server/templates/` - Embedded HTML templates

---

## Common Patterns

### Field Constants
```go
import cal "github.com/joeblew999/wellknown/pkg/calendar"

data[cal.FieldTitle]       // "title"
data[cal.FieldStart]       // "start"
data[cal.FieldAttendees]   // "attendees"
```

### Generator Signature
```go
func GenerateURL(data map[string]interface{}) (string, error)
func GenerateICS(data map[string]interface{}) ([]byte, error)
```

### Schema-Driven
- JSON Schema validates input
- UI Schema generates forms
- Generator functions assume valid input

---

## Git Workflow

**Only commit when explicitly asked!**

When creating commits:
1. Run `git status` and `git diff` in parallel
2. Draft commit message (focus on "why" not "what")
3. Add files and commit with this format:
```
Brief description (1-2 sentences)

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

**NEVER push** unless explicitly requested.

---

## Resources

- **Full architecture**: See `ARCHITECTURE.md`
- **Recent refactoring**: See `REFACTORING_COMPLETE.md`
- **Commit history**: Use `git log` for completed work
- **Claude Code docs**: https://docs.claude.com/en/docs/claude-code/

---

**Last Updated**: 2025-10-28
