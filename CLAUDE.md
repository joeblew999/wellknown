# Claude's Notes - wellknown Project

**Purpose of this file**: How to work on this codebase (for AI agents)

---

## 🚨 CRITICAL INSTRUCTIONS - READ FIRST 🚨

### ✅ MUST DO:
1. **ALWAYS verify module name is `joeblew999`** (NOT `joeblew99`!) - see Project Overview below
**Module**: `github.com/joeblew999/wellknown` ⚠️ NOTE: `999` not `99`!
**Path**: `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown`
2. **UPDATE STATUS.md** whenever making progress or completing tasks

### ❌ NEVER DO:
1. **NEVER use wrong module name** (`joeblew99` is WRONG, must be `joeblew999`)
3. **NEVER use `docs/` folder** for Claude tracking documents (user-facing docs only)
4. **NEVER add external dependencies** to core library (stdlib only)

### 📁 File Responsibilities:
- **CLAUDE.md** (this file): Technical decisions, architecture, how things work
- **STATUS.md**: Current state, what's done, what's next, milestones
- **docs/**: User-facing documentation ONLY (never for Claude tracking)

---

## Project Overview

**What**: Universal Go library for generating and opening deep links across Google and Apple **Web, Deckstop and Mobile** app ecosystems

**Key Principle**: Deterministic URL generation (same input → same output every time)

---

## Supported Deep Links

### Google Ecosystem
- Calendar: `googlecalendar://render?...`
- Maps: `comgooglemaps://?q=`
- Drive: `googledrive://`
- Mail: `mailto:`

### Apple Ecosystem
- Calendar: `calshow:`
- Maps: `maps://?q=`
- Files: `shareddocuments://`
- Mail: `mailto:`

---

## Architecture Decisions

### Folder Structure

```
wellknown/
├── cmd/
│   ├── wellknown/           # CLI tool
│   └── wellknown-mcp/       # MCP server
├── pkg/
│   ├── types/               # Shared data structures (CalendarEvent, etc.)
│   ├── google/              # calendar.go + calendar.tmpl (co-located!)
│   ├── apple/               # calendar.go + calendar.tmpl (co-located!)
│   ├── web/                 # Web fallbacks
│   ├── platform/            # Platform detection
│   └── opener/              # URL opener
├── internal/
│   ├── url/                 # URL building utilities (if needed)
│   └── template/            # Template renderer/loader (if needed)
├── examples/                # With go.work support
├── tests/
│   ├── testapp/            # Gio-based test application
│   └── integration/
├── docs/                    # User-facing documentation
│   └── testing-with-goup-util.md  # How to test with goup-util
├── CLAUDE.md               # This file
├── STATUS.md               # Project status tracking
└── go.mod
```

**Key Decision**: Templates co-located with implementations (e.g., `calendar.go` + `calendar.tmpl` in same directory)

### Template Strategy

- **Default templates**: Embedded via `go:embed` for zero-config usage
- **Custom templates**: Users can override with their own template files
- **Template engine**: Go's `text/template` from stdlib (zero deps)

Example:
```go
//go:embed calendar.tmpl
var calendarTemplate string
```

### API Design Patterns

**Simple Function Calls**:
```go
import "github.com/joeblew999/wellknown/pkg/google"
import "github.com/joeblew999/wellknown/pkg/types"

event := types.CalendarEvent{...}
url := google.Calendar(event)
```

**Builder Pattern**:
```go
event := calendar.NewEvent().
    Title("Meeting").
    StartTime(time.Now()).
    Build()
```

**Principles**:
- Simple defaults (zero config for common cases)
- Type-safe structs
- Return errors, don't panic

---

## MCP (Model Context Protocol) Integration

**Official Go SDK**: `github.com/modelcontextprotocol/go-sdk`
**Maintained by**: Anthropic + Google
**Implementation**: `cmd/wellknown-mcp/main.go`

**MCP Tools to Expose**:
- `create_calendar_event` - Generate Google/Apple calendar deep links
- `create_maps_link` - Generate navigation deep links
- `create_drive_link` - Generate file/folder deep links

**MCP Resources**:
- Templates (allow inspection of URL templates)

**Transport**: STDIO (standard MCP)

**Use Cases**:
- AI assistants creating calendar events
- LLMs generating navigation links
- Automated workflows

---

## Testing Strategy

**See**: `docs/testing-with-goup-util.md` for complete testing guide

**Summary**:
- **Unit tests**: Standard Go tests for URL generation
- **Integration tests**: Gio-based test app using goup-util
- **Manual verification**: Open deep links and verify they work

---

## Development Workflow

### Initial Setup

```bash
cd /Users/apple/workspace/go/src/github.com/joeblew999/wellknown

# Initialize go.work if using examples
go work init
go work use . examples/some-example

# Run tests
go test ./...
```

### Creating New Deep Link Support

1. Create package: `pkg/[platform]/[service].go`
2. Create template: `pkg/[platform]/[service].tmpl` (co-located!)
3. Implement with `go:embed` for template
4. Add unit tests: `[service]_test.go`
5. Add to test app UI
6. Update STATUS.md

Example: Adding Apple Maps support
- Create: `pkg/apple/maps.go`
- Create: `pkg/apple/maps.tmpl`
- Test: `pkg/apple/maps_test.go`

---

## Important Constraints

### Zero Dependencies (Core Library)

The core wellknown library must have **zero external dependencies**.
- Use only Go stdlib
- `text/template` is allowed (stdlib)
- Platform-specific code via build tags is OK

### Test Infrastructure Can Have Dependencies

- goup-util (build tool)
- gio-plugins (testing)
- Gio UI (test app)
- MCP SDK (MCP server)

### Deterministic Output

**Critical**: Same input must ALWAYS produce same output URL.
- No random IDs
- No timestamps (unless provided as input)
- Consistent URL encoding
- Template order matters

---

## Common Patterns

### Embedding Templates

```go
package google

import _ "embed"

//go:embed calendar.tmpl
var calendarTemplate string
```

### URL Building

```go
// Use Go's text/template
tmpl, err := template.New("calendar").Parse(calendarTemplate)
data := struct {
    Title string
    Start string
}{
    Title: "Meeting",
    Start: "20251023T100000Z",
}
var buf bytes.Buffer
tmpl.Execute(&buf, data)
url := buf.String()
```

### Error Handling

```go
// Always return errors, never panic
func Calendar(event Event, opts ...Option) (string, error) {
    if event.Title == "" {
        return "", fmt.Errorf("event title is required")
    }
    // ...
}
```

---

## Next Steps for Implementation

See STATUS.md for current tasks and milestones.

Key phases:
1. Core library structure (pkg/deeplink/)
2. Template system (internal/template/)
3. Unit tests
4. Test app (tests/testapp/)
5. MCP server (cmd/wellknown-mcp/)
6. CLI tool (cmd/wellknown/)

---

## Resources

- **MCP Go SDK**: https://github.com/modelcontextprotocol/go-sdk
- **Testing Guide**: `docs/testing-with-goup-util.md`
- **Go Templates**: https://pkg.go.dev/text/template

---

**Last Updated**: 2025-10-23
**Maintained by**: Claude (AI assistant)
