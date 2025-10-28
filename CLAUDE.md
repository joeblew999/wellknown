# Claude's Notes - wellknown Project

**Purpose of this file**: How to work on this codebase (for AI agents)

---

## 🚨 CRITICAL INSTRUCTIONS - READ FIRST 🚨

### ✅ MUST DO:
1. **ALWAYS verify module name is `joeblew999`** (NOT `joeblew99`!) - see Project Overview below
**Module**: `github.com/joeblew999/wellknown` ⚠️ NOTE: `999` not `99`!
**Path**: `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown`

### ❌ NEVER DO:
1. **NEVER use wrong module name** (`joeblew99` is WRONG, must be `joeblew999`)
2. **NEVER use `alert()`, `confirm()`, or `prompt()` dialogs** - use toast notifications or inline UI instead
3. **NEVER use `docs/` folder** for Claude tracking documents (user-facing docs only)
4. **NEVER add external dependencies** to core library (stdlib only)

### 📁 File Responsibilities:
- **CLAUDE.md** (this file): Technical decisions, architecture, how things work
- **docs/**: User-facing documentation ONLY (never for Claude tracking)
- **git log**: See commit history for what's been completed

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
- Calendar: **HTTP-served .ics files** ✅ **CORRECT and WORKING method**
  - ❌ **WRONG**: `data:text/calendar` URIs do NOT work (Safari rejects as "invalid address")
  - ✅ **CORRECT**: Serve .ics file via HTTP endpoint with proper headers
  - **Implementation**:
    - Generate ICS content (RFC 5545 iCalendar format)
    - Serve via HTTP endpoint: `/apple/calendar/download?event=<base64_ics>`
    - Headers: `Content-Type: text/calendar; charset=utf-8`, `Content-Disposition: attachment; filename="event.ics"`
    - Safari downloads .ics → macOS/iOS automatically offers "Add to Calendar"
  - **Tested and working** on macOS Safari (2025-10-27)
  - Note: `calshow:` exists but is undocumented and unreliable
  - **Research verified**: All working "Add to Calendar" tools use HTTP-served .ics files, NOT data URIs
- Maps: `maps://?q=` ✅ Universal on Apple devices
- Files: `shareddocuments://` ⚠️ iOS only
- Mail: `mailto:` ✅ Universal standard

---

## Architecture Decisions

### Folder Structure

```
wellknown/
├── pkg/                     # Core library
│   ├── types/              # Shared low-level types only (errors, etc.)
│   ├── schema/             # JSON Schema validation and UI generation
│   │   ├── validator.go           # ValidatorV6 - single source of truth for validation
│   │   ├── ui_schema.go           # UI Schema parser and HTML form generator
│   │   └── loader.go              # Schema loading utilities
│   ├── google/calendar/    # Google Calendar (platform-specific)
│   │   ├── schema.json            # JSON Schema for Google Calendar events
│   │   ├── ui_schema.json         # UI Schema for form generation
│   │   ├── generator.go           # Schema-driven URL generator
│   │   └── generator_test.go      # 4 passing tests
│   ├── apple/calendar/     # Apple Calendar (platform-specific)
│   │   ├── schema.json            # JSON Schema for Apple Calendar ICS events
│   │   ├── ui_schema.json         # UI Schema for form generation
│   │   ├── generator.go           # Schema-driven ICS generator
│   │   └── generator_test.go      # 9 passing tests
│   └── server/             # Web server for testing deep links
│       ├── server.go              # Server type - OWNS all dependencies (no globals!)
│       ├── handlers.go            # PageData, handler types
│       ├── navigation.go          # ServiceRegistry (no global state)
│       ├── calendar_generic.go    # Generic calendar handler (schema-driven)
│       ├── handlers_impl.go       # Home, stub handlers (use s.render())
│       ├── routes.go              # Route registration
│       ├── gcp_setup.go           # GCP OAuth setup wizard
│       ├── env_manager.go         # .env file manager
│       └── templates/             # Embedded HTML templates
│           ├── base.html          # Layout with navigation
│           ├── schema_form.html   # Schema-driven form template
│           ├── showcase.html      # Generic showcase template
│           ├── gcp_tool.html      # GCP setup wizard
│           └── stub.html          # Coming soon page
├── cmd/
│   ├── server/             # Test server binary
│   │   └── main.go                # Imports pkg/server, starts server
│   └── wellknown-mcp/      # MCP server (future)
├── docs/                   # User-facing documentation
├── CLAUDE.md              # This file (AI agent instructions)
└── go.mod
```

**Key Decisions**:
- **Schema-Driven Architecture**: JSON Schema is the single source of truth
  - ValidatorV6 validates all inputs (no custom validation code)
  - UI Schema generates HTML forms automatically
  - Schemas live in platform directories (e.g., `pkg/google/calendar/schema.json`)
- **Zero Global State**: Server struct owns ALL dependencies
  - No package-level variables (Templates, LocalURL, MobileURL, etc.)
  - ServiceRegistry is owned by Server (not global)
  - Perfect for testing and concurrency
- **Single Render Method**: `Server.render()` is THE ONLY place that calls ExecuteTemplate
  - Auto-populates LocalURL, MobileURL, Navigation
  - All handlers use `s.render()` - zero duplication
  - Perfect DRY compliance
- **Embedded templates**: Server uses `//go:embed` to embed all HTML templates (zero external files needed)
- **Full platform separation**: Each platform has completely separate schemas and generators
  - **Google**: 5 basic fields (Title, StartTime, EndTime, Location, Description) - URL-based
  - **Apple**: Full ICS spec with recurring events, attendees, reminders, priority, status, etc.
  - **Benefit**: Each platform uses its native capabilities without compromise

### Current Server Architecture

**Schema-Driven Design:**
- JSON Schema (`schema.json`) defines data structure and validation rules
- UI Schema (`ui_schema.json`) defines form generation (labels, placeholders, field types)
- ValidatorV6 (`pkg/schema/validator.go`) validates all inputs
- Forms auto-generate from schemas - zero manual HTML

**Zero Global State:**
- Server struct owns ALL dependencies (templates, mux, registry, gcpSetupStatus)
- ServiceRegistry is owned by Server (not package-level global)
- Thread-safe, testable, clear dependency graph

**Single Render Method:**
- `Server.render()` is THE ONLY place that calls ExecuteTemplate
- Auto-populates LocalURL, MobileURL, Navigation
- All handlers use `s.render()` - perfect DRY compliance

**Test Coverage:** 17/17 tests passing (9 Apple Calendar + 4 Google Calendar + 4 Server)

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

## UI/UX Patterns & Best Practices

### 🚫 NEVER Use Browser Dialogs

**CRITICAL RULE**: Never use `alert()`, `confirm()`, or `prompt()` in web interfaces.

**Why dialogs are BAD**:
- ❌ Block the entire page (bad UX)
- ❌ Impossible to test with Playwright without complex event handlers
- ❌ Slow down automated tests (need to handle async dialogs)
- ❌ Not customizable (ugly default browser styles)
- ❌ Not accessible (screen readers struggle with them)
- ❌ Mobile unfriendly (often render poorly)

**✅ USE INSTEAD**:
1. **Toast Notifications** - For success/error/warning/info messages
   - Auto-dismiss after 3 seconds
   - Non-blocking
   - Fully customizable
   - Easily testable
   - Example: `showToast('Success!', 'success')`

2. **Inline Error Messages** - For form validation
   - Show errors directly below input fields
   - Use `<div class="error-message">` with visibility toggle
   - Red border on invalid inputs
   - Example in [gcp_tool.html:183-192](pkg/server/templates/gcp_tool.html:183-192)

3. **Modal Dialogs** - For complex confirmations (when necessary)
   - Custom HTML/CSS modal overlays
   - Cancel/Confirm buttons
   - Fully testable with standard Playwright selectors

**Real-World Example (2025-10-27)**:
- **Before**: GCP Setup Wizard had 13 `alert()`/`confirm()` calls
- **After**: Replaced ALL with toast notifications
- **Result**: Tests went from 2+ minutes to 8.7 seconds (14x faster!)
- **Test pass rate**: 40% → 100%

### Toast Notification System

**Implementation** (see [gcp_tool.html:194-250](pkg/server/templates/gcp_tool.html:194-250)):
```javascript
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;

    const icons = {
        success: '✅',
        error: '❌',
        warning: '⚠️',
        info: 'ℹ️'
    };

    toast.innerHTML = `
        <div class="toast-icon">${icons[type] || icons.info}</div>
        <div class="toast-message">${message}</div>
    `;

    container.appendChild(toast);

    // Auto-dismiss after 3 seconds
    setTimeout(() => {
        toast.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}
```

**Usage**:
```javascript
// Success
showToast('✅ Reset complete!', 'success');

// Error
showToast('❌ Failed to save', 'error');

// Warning
showToast('⚠️ Please enter a project ID', 'warning');

// Info
showToast('ℹ️ Loading projects...', 'info');
```

**Testing with Playwright**:
```typescript
// NO dialog handlers needed!
await page.click('button:has-text("Save")');
await expect(page.locator('.toast.success')).toBeVisible();
await expect(page.locator('.toast-message')).toContainText('Saved');
```

---

## Testing Strategy

### Test Pyramid

1. **Unit Tests** (Fast, Automated)
   - Test URL generation determinism
   - Test time formatting
   - Test validation logic
   - Run: `go test ./pkg/...`

2. **E2E Tests with Playwright** (Fast, Automated)
   - **Run tests**: `make test-e2e` (runs 13 core tests in ~8 seconds)
   - **Location**: `tests/e2e/wizard-core.spec.ts`
   - **Browser**: WebKit (Safari) via Playwright
   - **Test suite**: 13/13 passing (100%)
   - **Speed**: 8.7 seconds for full suite
   - **Coverage**: Form validation, URL generation, state persistence, reset functionality
   - **Key achievement (2025-10-27)**: Removed ALL alert/confirm dialogs → 14x faster tests!

3. **Web Demo Testing** (Playwright MCP with Safari/WebKit)
   - **Development mode**: `cd examples/basic && air` (hot-reload on code changes)
   - **Standard mode**: `go run ./examples/basic/main.go`
   - **URLs displayed on startup**:
     - Desktop: `http://localhost:8080`
     - Mobile: `http://192.168.1.84:8080` (auto-detected local network IP)
   - **🚨 CRITICAL - Browser for Claude (AI Agent)**:
     - **USE PLAYWRIGHT MCP** - Configured in `.claude.json` to use WebKit (Safari)
     - The `.claude.json` file sets `PLAYWRIGHT_BROWSER=webkit` for this project
     - Playwright MCP will use Safari's WebKit engine on macOS
     - Use all mcp__playwright__* tools for automated testing
     - **Reason**: WebKit IS Safari on macOS - perfect for Apple integrations
   - Capture screenshots with Playwright: Saved to `.playwright-mcp/`
   - Screenshots can be copied to `docs/` and referenced in README.md

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

**🚨 CRITICAL: ALWAYS use Air for development (NEVER use `go run` directly)**

**Starting the web demo with hot-reload:**
```bash
make dev
# This runs Air from project root with proper configuration
```

**What Air does:**
- Watches `.go`, `.html`, `.tpl`, `.tmpl` files in `pkg/` and `cmd/`
- Automatically rebuilds and restarts server on changes
- No manual restarts needed during development
- Configured via `.air.toml` at project root

**Why Air is required:**
- Hot-reload on file changes (edit code → see changes immediately)
- Watches both `pkg/` and `cmd/` directories
- Prevents multiple server instances from running
- Consistent development experience

**Server Architecture:**
```
pkg/server/
├── server.go              # Server struct - owns ALL dependencies (no globals!)
├── navigation.go          # ServiceRegistry (owned by Server)
├── calendar_generic.go    # Schema-driven calendar handler
├── handlers_impl.go       # Home, stub handlers (all use s.render())
├── handlers.go            # PageData, CalendarConfig types
├── routes.go              # Route registration
├── gcp_setup.go           # GCP OAuth setup wizard
├── env_manager.go         # .env file manager
└── templates/
    ├── base.html          # Layout with navigation
    ├── schema_form.html   # Schema-driven form template
    ├── showcase.html      # Generic showcase template
    ├── gcp_tool.html      # GCP setup wizard
    └── stub.html          # Coming soon page
```

**Key Patterns:**
- **Zero Global State**: Server owns templates, mux, registry, gcpSetupStatus
- **Single Render Method**: `Server.render()` is THE ONLY ExecuteTemplate call
- **Schema-Driven**: Forms auto-generate from JSON Schema + UI Schema
- **DRY Compliance**: All handlers use `s.render()` - zero duplication
- **Mobile URLs**: Auto-detected local network IP for easy mobile testing
- **Responsive**: Hamburger menu on mobile (≤768px), fixed sidebar on desktop

### Adding a New Service (Schema-Driven Approach)

**Example: Adding Apple Maps**

**Step 1:** Create JSON Schema in `pkg/apple/maps/schema.json`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "minLength": 1,
      "description": "Search query or address"
    },
    "latitude": {"type": "number"},
    "longitude": {"type": "number"}
  },
  "required": ["query"]
}
```

**Step 2:** Create UI Schema in `pkg/apple/maps/ui_schema.json`
```json
{
  "fields": [
    {
      "key": "query",
      "type": "text",
      "label": "Search Query",
      "placeholder": "Enter address or place name"
    },
    {
      "key": "latitude",
      "type": "number",
      "label": "Latitude (optional)"
    },
    {
      "key": "longitude",
      "type": "number",
      "label": "Longitude (optional)"
    }
  ]
}
```

**Step 3:** Create generator in `pkg/apple/maps/generator.go`
```go
package maps

func GenerateURL(data map[string]interface{}) (string, error) {
    query, _ := data["query"].(string)
    return fmt.Sprintf("maps://?q=%s", url.QueryEscape(query)), nil
}
```

**Step 4:** Register route in `pkg/server/routes.go`
```go
// Add to registerMapsRoutes()
handler := s.makeGenericCalendarHandler(CalendarConfig{
    Platform:     "apple",
    AppType:      "maps",
    SuccessLabel: "Deep Link",
    GenerateURL:  maps.GenerateURL,
})
s.mux.HandleFunc("/apple/maps", handler)
```

**That's it!** The schema-driven system handles:
- ✅ Form generation from UI Schema
- ✅ Validation from JSON Schema
- ✅ Type coercion (strings to numbers, etc.)
- ✅ Error messages
- ✅ Zero custom validation code needed

**Benefits:**
- Schema changes don't require Go code changes
- Forms auto-update when schema changes
- Single source of truth for validation rules

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

## Current Status

**Architecture:**
- ✅ **Schema-Driven**: JSON Schema validates, UI Schema generates forms
- ✅ **Zero Global State**: Server owns all dependencies
- ✅ **Single Render Method**: Perfect DRY compliance
- ✅ **Test Coverage**: 17/17 Go tests passing

**Implemented Services:**
- ✅ **Google Calendar**: 5-field event generator (web URL approach)
- ✅ **Apple Calendar**: Full ICS spec with advanced features (HTTP-served .ics files)
- ✅ **GCP Setup Wizard**: OAuth credential management with toast notifications

**Infrastructure:**
- ✅ **Air hot-reload**: Development server with instant updates
- ✅ **Playwright E2E tests**: 13/13 passing in 8.7 seconds
- ✅ **Mobile testing**: Auto-detected network IP for device testing

**For detailed history**: See `git log` for commit history

---

## Resources

- **MCP Go SDK**: https://github.com/modelcontextprotocol/go-sdk
- **Playwright Documentation**: https://playwright.dev/
- **E2E Testing with Bun**: https://bun.sh/docs/cli/test
- **Go Templates**: https://pkg.go.dev/text/template

---

**Last Updated**: 2025-10-28 (Completed schema-driven architecture migration)
**Maintained by**: Claude (AI assistant)
