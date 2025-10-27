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

**What**: Universal Go library for generating and opening deep links across Google and Apple **Web, Desktop and Mobile** app ecosystems

**Key Principle**: Deterministic URL generation (same input → same output every time)

---

## Supported Deep Links

### ⚠️ CRITICAL: Google Calendar Deep Link Strategy

**Research Finding (2025-10-26, Verified 2025-10-27):**
- `googlecalendar://` exists and opens the app but does NOT support event parameters
- Google does not document any way to pass event data via native deep link
- **Solution**: Use web URLs (`https://calendar.google.com/calendar/render?action=TEMPLATE&...`)
- **How it works**:
  - Desktop: Opens in browser
  - Mobile: Opens in browser, then OS offers to open in Google Calendar app with event data
  - **Tested and working** on both iOS and Android mobile devices
- **Decision**: Our library uses web URLs as the universal approach for Google Calendar

### Google Ecosystem
- Calendar: ❌ No native deep link - **Use `https://calendar.google.com` instead**
- Maps: `comgooglemaps://?q=` ✅ Documented and working
- Drive: `googledrive://` ⚠️ May work but undocumented
- Mail: `mailto:` ✅ Universal standard

### Apple Ecosystem
- Calendar: `data:text/calendar` (ICS format) ✅ **Universal cross-platform method**
  - Uses ICS (iCalendar) format with base64-encoded data URI
  - Works on macOS, iOS, and cross-platform
  - Format: `data:text/calendar;base64,<encoded ICS>`
  - Note: `calshow:` exists but is undocumented and unreliable
- Maps: `maps://?q=` ✅ Universal on Apple devices
- Files: `shareddocuments://` ⚠️ iOS only
- Mail: `mailto:` ✅ Universal standard

---

## Architecture Decisions

### Folder Structure

```
wellknown/
├── pkg/                     # Core library (zero external dependencies)
│   ├── types/              # Shared data structures (CalendarEvent, etc.)
│   ├── google/             # Google service implementations
│   │   ├── calendar.go
│   │   └── calendar_examples.go
│   └── apple/              # Apple service implementations
│       ├── calendar.go
│       └── calendar_examples.go
├── examples/
│   ├── basic/              # Web demo (air hot-reload)
│   │   ├── main.go
│   │   ├── handlers/
│   │   │   ├── handlers.go         # Shared types (PageData)
│   │   │   ├── generic.go          # Service registry, factories
│   │   │   ├── google_calendar.go  # Google Calendar service registration
│   │   │   ├── apple_calendar.go   # Apple Calendar service registration
│   │   │   └── stub.go             # Stub handler for unimplemented
│   │   ├── templates/
│   │   │   ├── base.html           # Layout with nav
│   │   │   ├── custom.html         # Generic form template
│   │   │   ├── showcase.html       # Generic showcase template
│   │   │   └── stub.html           # Coming soon page
│   │   └── .air.toml
│   ├── mcp/                # MCP server example
│   └── webview/            # Webview example
├── docs/                   # User-facing documentation
├── CLAUDE.md              # This file (AI agent instructions)
├── STATUS.md              # Project status tracking
└── go.mod
```

**Key Decisions**:
- **No embedded templates**: Web demo uses HTML templates, not go:embed
- **Service registry pattern**: Services self-register with `RegisterService()`
- **Generic templates**: `custom.html` and `showcase.html` work for all services

### Web Demo Architecture Evolution (Phase 3 & 4 Refactoring)

**Phase 1-2: Template & Handler Generalization** (Completed)
- Created generic `custom.html` and `showcase.html` templates
- Built handler factories: `CalendarHandler(config)`, `ShowcaseHandler(config)`
- Reduced code duplication from 125 lines to 19 lines per service (84% reduction)

**Phase 3: Service Registry & Auto-Registration** (Completed)
- Created `ServiceRegistry` pattern in `handlers/generic.go`
- Services self-register with `RegisterService(ServiceConfig)`
- Routes auto-create via `RegisterRoutes()` - zero manual registration
- Example: Adding Apple Calendar auto-created `/apple/calendar` and `/apple/calendar/showcase`

**Phase 4: Type Safety Improvements** (Completed)
- Changed `PageData.Event` from `interface{}` to `*types.CalendarEvent`
- Added `ServiceExample` interface for examples
- Improved compile-time type checking
- Better IDE autocomplete and error detection

**Result**: Adding a new service now requires only ~15-19 lines of configuration code!

**Current Template Strategy:**
- **Generic HTML templates**: `custom.html` and `showcase.html` work for all services
- **Conditional rendering**: Templates use `{{if eq .AppType "calendar"}}` for service-specific sections
- **Go's html/template**: Standard library, zero external dependencies
- **Template loading**: All templates loaded via `template.ParseGlob("templates/*.html")`

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

### Test Pyramid

1. **Unit Tests** (Fast, Automated)
   - Test URL generation determinism
   - Test time formatting
   - Test validation logic
   - Run: `go test ./pkg/...`

2. **Web Demo Testing** (Interactive, Playwright MCP)
   - **Development mode**: `cd examples/basic && air` (hot-reload on code changes)
   - **Standard mode**: `go run ./examples/basic/main.go`
   - **URLs displayed on startup**:
     - Desktop: `http://localhost:8080`
     - Mobile: `http://192.168.1.84:8080` (auto-detected local network IP)
   - **🚨 CRITICAL - Browser Selection for Claude (AI Agent)**:
     - **When testing Apple services** (Calendar, Maps):
       - ❌ DO NOT use Playwright MCP if it's configured for Chrome
       - ✅ USE: `open -a Safari http://localhost:8080/apple/calendar`
       - **Reason**: Apple Calendar uses data URI ICS format best tested in Safari
     - **When testing Google services**: Playwright MCP with Chrome is fine
   - Use Playwright MCP to automate browser testing (Google services only)
   - Capture screenshots: Saved to `.playwright-mcp/`
   - **IMPORTANT**: Screenshots can be copied to `docs/` and referenced in README.md

3. **Deep Link Verification** (Manual, Real Device)
   - **Google Calendar**: Use web URL - mobile OS automatically offers to open in app
   - **Testing workflow**:
     1. Access mobile URL: `http://192.168.1.84:8080` from phone
     2. Click "Open in Calendar" button
     3. Browser opens URL, then OS prompts to open in Google Calendar app
     4. Verify event data (title, time, location, description) is correct
   - **Verified working** on iOS and Android mobile devices (2025-10-27)

4. **QR Code Testing** (Optional)
   - Generate QR code from deep link
   - Scan with mobile device
   - Verify app opens correctly

### Screenshot Workflow

When testing with Playwright MCP:
1. Playwright saves screenshots to `.playwright-mcp/`
2. Copy useful screenshots to `docs/screenshots/`
3. Reference in README.md to show working demos
4. Commit screenshots to prove functionality

Example:
```bash
# After Playwright test
cp .playwright-mcp/wellknown-demo-success.png docs/screenshots/
# Add to README: ![Demo](docs/screenshots/wellknown-demo-success.png)
```

### URL Format Validation

**Critical**: Always verify URLs against official documentation
- Google Calendar: Check `action=CREATE` vs `action=TEMPLATE`
- URL encoding: Use `url.PathEscape()` for `%20` or `url.QueryEscape()` for `+`
- Test both native app URL and web fallback URL

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

# Install Air for hot-reload (first time only)
go install github.com/air-verse/air@latest
```

### Web Demo Development

**Starting the web demo with hot-reload:**
```bash
cd examples/basic
air
```

**What Air does:**
- Watches `.go`, `.html`, `.tpl`, `.tmpl` files
- Automatically rebuilds and restarts server on changes
- No manual restarts needed during development

**Web Demo Architecture:**
```
examples/basic/
├── main.go                    # Entry point (calls RegisterRoutes())
├── handlers/
│   ├── handlers.go           # PageData, ServiceExample interface
│   ├── generic.go            # ServiceRegistry, factories, RegisterRoutes()
│   ├── google_calendar.go    # 19 lines - RegisterService(config)
│   ├── apple_calendar.go     # 19 lines - RegisterService(config)
│   └── stub.go               # Generic stub handler
├── templates/
│   ├── base.html             # Layout with nav, CSS, JS
│   ├── custom.html           # Generic form (all services)
│   ├── showcase.html         # Generic showcase (all services)
│   └── stub.html             # Coming soon page
├── .air.toml                 # Air hot-reload configuration
└── README.md                 # Usage instructions
```

**Key Patterns:**
- **Service Registry**: Services self-register, routes auto-created
- **Generic Templates**: One template works for all services (conditional rendering)
- **Handler Factories**: `CalendarHandler(config)`, `ShowcaseHandler(config)`
- **Type Safety**: PageData uses `*types.CalendarEvent` not `interface{}`
- **Adding Services**: Only ~15-19 lines of config code needed!
- **Mobile URLs**: Auto-detected local network IP for easy mobile testing
- **Responsive**: Hamburger menu on mobile (≤768px), fixed sidebar on desktop
- **Air Config**: Watches handlers/ and templates/ directories

### Adding a New Service (Step-by-Step)

**Example: Adding Apple Maps**

**Step 1:** Create core implementation in `pkg/apple/maps.go`
```go
package apple

import "github.com/joeblew999/wellknown/pkg/types"

func Maps(location types.Location) (string, error) {
    // Implementation
    return fmt.Sprintf("maps://?q=%s", url.QueryEscape(location.Address)), nil
}
```

**Step 2:** Create examples in `pkg/apple/maps_examples.go`
```go
package apple

type MapsExample struct {
    Name        string
    Description string
    Location    types.Location
}

func (e MapsExample) GetName() string { return e.Name }
func (e MapsExample) GetDescription() string { return e.Description }

var MapsExamples = []MapsExample{...}
```

**Step 3:** Register service in `examples/basic/handlers/apple_maps.go` (**only ~15-19 lines!**)
```go
package handlers

import "github.com/joeblew999/wellknown/pkg/apple"

var AppleMapsService = RegisterService(ServiceConfig{
    Platform:  "apple",
    AppType:   "maps",
    Examples:  apple.MapsExamples,
    Generator: apple.Maps,  // Adapt if different signature
})

var AppleMaps = AppleMapsService.CustomHandler
var AppleMapsShowcase = AppleMapsService.ShowcaseHandler
```

**Step 4:** Done! Routes auto-register
- `/apple/maps` - Custom form
- `/apple/maps/showcase` - Examples showcase
- Navigation sidebar automatically shows links (if already in base.html)

**That's it!** The service registry handles:
- ✅ Route registration
- ✅ Handler creation
- ✅ Template rendering
- ✅ QR code generation

**Optional:** Add unit tests in `pkg/apple/maps_test.go`

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

## Project Status and Roadmap

**For current status, completed work, and next tasks**: See [STATUS.md](STATUS.md)

**Recent major milestones:**
- ✅ Web demo with Air hot-reload
- ✅ Phase 3 & 4 refactoring (service registry, type safety)
- ✅ Google Calendar (web URL approach)
- ✅ Apple Calendar (ICS data URI approach)
- ✅ Generic template system (custom.html, showcase.html)
- ✅ Auto-registration pattern (~15 lines to add new service)

---

## Resources

- **MCP Go SDK**: https://github.com/modelcontextprotocol/go-sdk
- **Testing Guide**: `docs/testing-with-goup-util.md`
- **Go Templates**: https://pkg.go.dev/text/template

---

**Last Updated**: 2025-10-27 (Comprehensive cleanup after Phase 3 & 4 refactoring)
**Maintained by**: Claude (AI assistant)
