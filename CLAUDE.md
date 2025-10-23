# Claude's Notes - wellknown Project

**Purpose of this file**: How to work on this codebase (for AI agents)

---

## IMPORTANT: File Responsibilities

**CLAUDE.md** (this file): Technical decisions, architecture, how things work
**STATUS.md**: Current state, what's done, what's next, milestones

**⚠️ ALWAYS UPDATE STATUS.md** when making progress or completing tasks
**🚫 DO NOT USE `docs/` FOLDER** for Claude tracking documents (user-facing docs only)

---

## Project Overview

**What**: Universal Go library for generating and opening deep links across Google and Apple app ecosystems
**Language**: Go (pure, zero external dependencies for core library)
**Module**: `github.com/joeblew999/wellknown`
**Path**: `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown`

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
│   ├── deeplink/
│   │   ├── google/          # calendar.go + calendar.tmpl (co-located!)
│   │   ├── apple/           # maps.go + maps.tmpl (co-located!)
│   │   └── web/             # Web fallbacks
│   ├── platform/            # Platform detection
│   └── opener/              # URL opener
├── internal/
│   ├── url/                 # URL building utilities
│   └── template/            # Template renderer/loader
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

**Functional Options**:
```go
url := google.Calendar(event,
    deeplink.WithTemplate("/custom/template.tmpl"),
    deeplink.WithPlatform("ios"))
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

1. Create package: `pkg/deeplink/[platform]/[app].go`
2. Create template: `pkg/deeplink/[platform]/[app].tmpl` (co-located!)
3. Implement interface with `go:embed` for template
4. Add unit tests: `[app]_test.go`
5. Add to test app UI
6. Update STATUS.md

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
