# PDF Form System Architecture

## Overview

The PDF form system provides a complete workflow for browsing, downloading, inspecting, filling, and testing PDF forms. It's organized as a clean, DRY codebase with centralized configuration.

## Directory Structure

```
pkg/pdf/
├── config.go              # Centralized path configuration
├── cli/                   # CLI package
│   └── cli.go             # All CLI commands
├── web/                   # Web server package
│   ├── handlers.go        # HTTP handlers
│   ├── server.go          # Server setup
│   ├── templates.go       # Template rendering
│   └── templates/         # Embedded HTML templates
│       ├── home.html
│       ├── browse.html
│       ├── download.html
│       ├── inspect.html
│       ├── fill.html
│       └── test.html
├── examples/pdfform/      # Entry point
│   └── main.go            # Sets up config and runs CLI
├── case.go                # Case management
├── commands.go            # Core PDF operations
├── forms_catalog.go       # Form catalog management
├── pdfform.go             # PDF filling logic
├── provenance.go          # Metadata tracking
└── workflows.go           # Multi-step workflows

data/                      # Data directory (not in pkg/)
├── catalog/               # Form catalogs
│   └── australian_transfer_forms.csv
├── downloads/             # Downloaded PDFs
├── templates/             # Field templates
├── outputs/               # Filled PDFs
├── cases/                 # Case files
│   ├── test_scenarios/    # Test cases
│   └── {entity}/          # Entity-specific cases
└── temp/                  # Temporary files
```

## Configuration System

### Central Config (`config.go`)

All file paths are managed through the `Config` struct:

```go
type Config struct {
    DataDir         string  // Base data directory
    CatalogDir      string  // catalog/
    DownloadsDir    string  // downloads/
    TemplatesDir    string  // templates/
    OutputsDir      string  // outputs/
    CasesDir        string  // cases/
    TempDir         string  // temp/
    CatalogFile     string  // australian_transfer_forms.csv
    SystemTempDir   string  // OS temp directory
}
```

### Path Helper Methods

The config provides helper methods for getting full paths:

```go
cfg.CatalogFilePath()       // data/catalog/australian_transfer_forms.csv
cfg.DownloadsPath()         // data/downloads
cfg.TemplatesPath()         // data/templates
cfg.OutputsPath()           // data/outputs
cfg.CasesPath()             // data/cases
cfg.TestScenariosPath()     // data/cases/test_scenarios
cfg.TempPath()              // data/temp
cfg.EntityCasesPath(name)   // data/cases/{name}
```

### Global Default Config

The package maintains a global default config that can be set at startup:

```go
// In main.go
config := pdfform.NewConfig("../../data")
pdfform.SetDefaultConfig(config)

// Now all package functions use this config
pdfform.Browse(...)  // Uses default config paths
```

### Benefits

1. **DRY** - All paths defined once in `config.go`
2. **Testable** - Easy to inject different configs for testing
3. **Flexible** - Can run with different data directories
4. **Garble-safe** - String constants in one place
5. **Thread-safe** - Config access protected by mutex

## Three-Layer Architecture

The codebase follows a strict three-layer architecture:

```
┌─────────────────────────────────────────┐
│  INTERFACE LAYER                        │
│  - CLI (cli/cli.go)                     │
│  - Web API (web/api/handlers.go)        │
│  - Web GUI (web/gui/handlers.go)        │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  COMMANDS LAYER (Single Source of Truth)│
│  - commands/browse.go                   │
│  - commands/download.go                 │
│  - commands/inspect.go                  │
│  - commands/fill.go                     │
│  - commands/cases.go                    │
│  - commands/events.go (event bus)       │
│  - commands/constants.go                │
│  - commands/helpers.go                  │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  CORE LAYER                             │
│  - pdfform.go (PDF operations)          │
│  - forms_catalog.go (catalog)           │
│  - case.go (case management)            │
│  - workflows.go (orchestration)         │
│  - config.go (configuration)            │
└─────────────────────────────────────────┘
```

### Layer Rules

1. **Interface Layer** → Only calls Commands Layer
2. **Commands Layer** → Single source of truth, emits events
3. **Core Layer** → Low-level operations, no events

**Critical**: Never bypass the Commands Layer. CLI, API, and GUI must all use commands.

## Package Structure

### Core Packages

#### `pkg/pdf` (Root Package)
- **config.go** - Configuration and path management
- **pdfform.go** - Core PDF filling functionality
- **forms_catalog.go** - Australian government forms catalog
- **case.go** - Case management (save/load form data)
- **provenance.go** - PDF metadata and provenance tracking
- **workflows.go** - Multi-step workflow orchestration
- **network.go** - Network utilities for LAN discovery
- **certs.go** - HTTPS certificate management via mkcert

#### `pkg/pdf/commands` (Commands Layer - Single Source of Truth)
- **browse.go** - Browse catalog command
- **download.go** - Download form command
- **inspect.go** - Inspect PDF fields command
- **fill.go** - Fill PDF command
- **cases.go** - Case management commands
- **events.go** - Event bus system for observability
- **constants.go** - All magic values (suffixes, stages, progress values)
- **helpers.go** - Shared utility functions (DRY)

#### `pkg/pdf/cli`
- **cli.go** - Complete CLI implementation using Cobra
- Accepts `*Config` and uses it for all paths
- Implements 5-step numbered workflow
- Calls commands package only

#### `pkg/pdf/web`
- **server.go** - HTTP server setup with HTTPS support
- Auto-discovery via FindDataDir()
- Zero-config deployment functions

#### `pkg/pdf/web/api`
- **handlers.go** - REST API handlers
- Returns JSON responses
- Uses commands package only

#### `pkg/pdf/web/gui`
- **handlers.go** - Web GUI handlers (future: htmx + datastar)
- **templates.go** - Embedded template rendering
- **templates/** - Embedded HTML files
- Uses commands package only

#### `pkg/pdf/web/httputil`
- **httputil.go** - HTTP helper functions for consistent handlers
- ValidateMethod, GetRequiredFormValue, RespondJSON, etc.

#### `examples/pdfform`
- **main.go** - Entry point
- Sets up config
- Ensures directories exist
- Runs CLI

## Three Ways to Use

### 1. CLI (Command Line)

```bash
pdfform 1-browse --state VIC
pdfform 2-download F3520
pdfform 3-inspect f3520.pdf
pdfform 4-fill template.json --flatten
pdfform 5-test vba_basic
```

### 2. Web GUI

```bash
pdfform serve --port 8080
# Open http://localhost:8080
```

### 3. Go API

```go
import pdfform "github.com/joeblew999/wellknown/pkg/pdf"

// Set up config
cfg := pdfform.NewConfig("./data")
pdfform.SetDefaultConfig(cfg)

// Use commands
result, err := pdfform.Browse(pdfform.BrowseOptions{
    CatalogPath: cfg.CatalogFilePath(),
    State: "VIC",
})
```

## Maintainability & Code Patterns

### DRY Principle - Use Helpers

**Never duplicate code.** All common patterns are in `commands/helpers.go`:

```go
// ✅ GOOD - Use helper
outputPath := DetermineOutputPath(dataPath, outputDir, FilledPDFSuffix)

// ❌ BAD - Duplicate logic
if outputDir == "" {
    base := filepath.Base(dataPath)
    name := base[:len(base)-len(filepath.Ext(base))]
    outputPath = name + "_filled.pdf"
}
```

**Available helpers:**
- `DetermineOutputPath(dataPath, outputDir, suffix)` - Smart output path resolution
- `EnsureOutputDir(filePath)` - Create parent directories
- `EmitStageError(eventType, stage, err, context)` - Standardized error events
- `BaseNameWithoutExt(path)` - Get filename without extension

### Use Constants, Not Magic Values

**Never use hardcoded strings.** All constants are in `commands/constants.go`:

```go
// ✅ GOOD - Use constant
outputPath := name + FilledPDFSuffix

// ❌ BAD - Magic string
outputPath := name + "_filled.pdf"
```

**Available constants:**
- File suffixes: `FilledPDFSuffix`, `FlatPDFSuffix`, `TemplateJSONSuffix`, `MetaJSONSuffix`
- Progress stages: `DownloadStageFoundForm`, `DownloadStageDownloading`, etc.
- Progress values: `ProgressFoundForm`, `ProgressDownloading`, `ProgressComplete`
- Permissions: `DefaultDirPerm`

### HTTP Handler Pattern

**Use httputil helpers for consistent error handling:**

```go
// ✅ GOOD - Use httputil
import "github.com/joeblew999/wellknown/pkg/pdf/web/httputil"

func (h *Handler) HandleFill(w http.ResponseWriter, r *http.Request) {
    if !httputil.ValidateMethod(w, r, "POST") {
        return
    }

    dataPath, ok := httputil.GetRequiredFormValue(w, r, "data_path")
    if !ok {
        return
    }

    result, err := commands.Fill(...)
    if err != nil {
        httputil.RespondInternalError(w, err)
        return
    }

    httputil.RespondJSONOK(w, result)
}

// ❌ BAD - Manual error handling
func (h *Handler) HandleFill(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    dataPath := r.FormValue("data_path")
    if dataPath == "" {
        http.Error(w, "data_path is required", http.StatusBadRequest)
        return
    }
    // ... duplicated error handling
}
```

### Event-Driven Pattern

**All commands must emit events:**

```go
func NewCommand(opts NewCommandOptions) (*NewCommandResult, error) {
    // 1. Emit started event
    Emit(EventNewStarted, map[string]interface{}{
        "input": opts.Input,
    })

    // 2. Do work
    result, err := doWork()
    if err != nil {
        // 3. Emit error event
        EmitError(EventNewError, err, map[string]interface{}{
            "stage": "do_work",
        })
        return nil, err
    }

    // 4. Emit completed event
    Emit(EventNewCompleted, map[string]interface{}{
        "output": result.Output,
    })

    return result, nil
}
```

**Event naming convention:**
- `{action}.started` - Command started
- `{action}.progress` - Progress update (include `progress` field 0.0-1.0)
- `{action}.completed` - Command completed successfully
- `{action}.error` - Command failed (include `stage` field)

### Error Handling with Stages

**Always include stage information in errors:**

```go
// ✅ GOOD - Use EmitStageError
if err := os.MkdirAll(outputDir, DefaultDirPerm); err != nil {
    EmitStageError(EventFillError, StageCreateDir, err, map[string]interface{}{
        "output_dir": outputDir,
    })
    return nil, fmt.Errorf("failed to create directory: %w", err)
}

// ❌ BAD - Generic error
if err := os.MkdirAll(outputDir, 0755); err != nil {
    EmitError(EventFillError, err, nil)
    return nil, err
}
```

## Data Flow & Workflow

### The 5-Step Workflow

**This is the core workflow - DO NOT change the order:**

```
1. BROWSE (commands/browse.go)
   └─> Load catalog CSV
   └─> Filter by state (optional)
   └─> Display available forms
   └─> Events: browse.started, browse.completed, browse.error

2. DOWNLOAD (commands/download.go)
   └─> Find form by code in catalog
   └─> Fetch PDF from DirectPDFURL
   └─> Save to downloads/
   └─> Create provenance metadata (.meta.json)
   └─> Events: download.started, download.progress (0.2, 0.4, 0.8), download.completed, download.error
   └─> Stages: found_form, downloading, saving_metadata

3. INSPECT (commands/inspect.go)
   └─> Extract PDF form fields
   └─> Generate template JSON with field metadata
   └─> Save to templates/
   └─> Events: inspect.started, inspect.completed, inspect.error

4. FILL (commands/fill.go)
   └─> Load template JSON with field values
   └─> Fill PDF with data
   └─> Optionally flatten (remove form fields)
   └─> Save to outputs/
   └─> Events: fill.started, fill.completed, fill.error
   └─> Stages: create_dir, fill_pdf, flatten

5. TEST (commands not yet created for this)
   └─> Load test case JSON from cases/test_scenarios/
   └─> Run fill operation
   └─> Verify success/failure
   └─> Report results
```

**Workflow automation:** `workflows.go` can run these steps automatically.

### Case Management

```
CREATE CASE
   └─> User provides: form_code, case_name, entity_name
   └─> Generate case_id
   └─> Create JSON in cases/{entity}/{case_id}.json
   └─> Store metadata + field values

FILL FROM CASE
   └─> Load case JSON
   └─> Get PDF from downloads/
   └─> Fill fields from case data
   └─> Save to outputs/
```

## Key Design Decisions

### 1. Centralized Configuration
**Why**: DRY principle, easier to maintain, garble-safe, testable
**How**: Single `Config` struct with path helper methods

### 2. Separate CLI and Web Packages
**Why**: Clean separation of concerns, reusable packages
**How**: Both accept `*Config`, implement different interfaces

### 3. Embedded Templates
**Why**: Single binary deployment, no external file dependencies
**How**: `//go:embed templates/*.html` in web package

### 4. Package-Level Default Config
**Why**: Convenient API, backward compatibility
**How**: Global config with mutex-protected access

### 5. Provenance Tracking
**Why**: Know where PDFs came from, track modifications
**How**: `.meta.json` files alongside PDFs

### 6. Case Management
**Why**: Reuse form data, organize by entity
**How**: JSON files in `cases/{entity}/` directory

## Future: Reactive UI with Datastar

### Datastar Integration (Planned)

**Package:** `github.com/starfederation/datastar-go`
**Version:** Check v1.0.0-RC.6+ release notes before implementing

The event bus system is designed to integrate with datastar for real-time UI updates:

```go
// Example: Stream events to browser via SSE
func (h *Handler) HandleEventStream(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")

    eventChan := commands.DefaultEventBus.Subscribe("*")
    defer close(eventChan)

    for event := range eventChan {
        // Send datastar fragment update
        datastar.PatchFragment(w, event.Data)
        w.(http.Flusher).Flush()
    }
}
```

**Key considerations:**
- Event bus already supports async operations
- NATS Jetstream will provide linearizability for async nature
- GUI handlers will use htmx + datastar for reactive updates
- Commands remain synchronous, events are async

### NATS Jetstream Integration (Planned)

**Purpose:** Ensure linearizability in the async event-driven architecture

As the system scales with real-time UI updates, NATS Jetstream will:
- Guarantee event ordering
- Provide event persistence
- Enable horizontal scaling
- Support replay and recovery

**Implementation notes:**
- Keep commands package synchronous
- Use NATS for event distribution only
- Maintain backward compatibility with in-memory event bus

## Adding New Features

### Adding a New Command

**Follow this pattern exactly:**

1. **Create command in `commands/new_command.go`:**
```go
package commands

type NewCommandOptions struct {
    Input string
    // Add context.Context in future
}

type NewCommandResult struct {
    Output string
}

func NewCommand(opts NewCommandOptions) (*NewCommandResult, error) {
    // 1. Emit started event
    Emit(EventNewStarted, map[string]interface{}{
        "input": opts.Input,
    })

    // 2. Use helpers and constants
    outputPath := DetermineOutputPath(opts.Input, "", "_new.pdf")
    if err := EnsureOutputDir(outputPath); err != nil {
        EmitStageError(EventNewError, StageCreate, err, nil)
        return nil, err
    }

    // 3. Do work using core layer
    result, err := doWork()
    if err != nil {
        EmitStageError(EventNewError, "do_work", err, nil)
        return nil, err
    }

    // 4. Emit completed event
    Emit(EventNewCompleted, map[string]interface{}{
        "output": result.Output,
    })

    return result, nil
}
```

2. **Add event types to `commands/events.go`:**
```go
const (
    EventNewStarted   EventType = "new.started"
    EventNewCompleted EventType = "new.completed"
    EventNewError     EventType = "new.error"
)
```

3. **Add CLI command in `cli/cli.go`:**
```go
newCmd := &cobra.Command{
    Use:   "new-command",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        result, err := commands.NewCommand(commands.NewCommandOptions{
            Input: args[0],
        })
        if err != nil {
            return err
        }
        fmt.Printf("Output: %s\n", result.Output)
        return nil
    },
}
```

4. **Add API handler in `web/api/handlers.go`:**
```go
func (h *Handler) HandleNew(w http.ResponseWriter, r *http.Request) {
    if !httputil.ValidateMethod(w, r, "POST") {
        return
    }

    input, ok := httputil.GetRequiredFormValue(w, r, "input")
    if !ok {
        return
    }

    result, err := commands.NewCommand(commands.NewCommandOptions{
        Input: input,
    })
    if err != nil {
        httputil.RespondInternalError(w, err)
        return
    }

    httputil.RespondJSONOK(w, result)
}
```

5. **Add GUI handler in `web/gui/handlers.go`:**
```go
func (h *Handler) HandleNewPage(w http.ResponseWriter, r *http.Request) {
    html, err := renderTemplate("new.html")
    if err != nil {
        httputil.RespondInternalError(w, err)
        return
    }
    w.Write([]byte(html))
}
```

### Adding a New Path Type

1. Add field to `Config` struct in `config.go`
2. Add helper method (e.g., `NewPath()`)
3. Update `EnsureDirectories()` if it's a directory

## Testing with Config

```go
func TestWithCustomConfig(t *testing.T) {
    // Create temp directory
    tempDir := t.TempDir()

    // Create custom config
    cfg := pdfform.NewConfig(tempDir)
    pdfform.SetDefaultConfig(cfg)
    cfg.EnsureDirectories()

    // Run tests - all operations use tempDir
    result, err := pdfform.Browse(...)
    // ...
}
```

## Garble Compatibility

All string constants (paths, filenames) are centralized in `config.go`. When using `garble`, the obfuscation will only affect this one file, making debugging easier.

The config values are set at runtime from `main.go`, so they remain configurable even when obfuscated.

## Security Considerations

1. **Path Traversal**: All paths validated in handlers
2. **Input Validation**: Form codes, case names sanitized
3. **File Permissions**: Directories created with 0755
4. **Temp Files**: Cleaned up after use
5. **No Secrets**: Config only contains paths, not credentials

## Performance

- **Embedded Templates**: No filesystem I/O for templates
- **Config Caching**: Global default config cached
- **Lazy Loading**: Catalog loaded only when needed
- **Streaming**: Large PDFs streamed, not loaded into memory

## Common Pitfalls & How to Avoid Them

### 1. Bypassing Commands Layer
**❌ WRONG:**
```go
// In API handler - calling core directly
catalog, err := pdfform.LoadFormsCatalog(path)
```

**✅ CORRECT:**
```go
// In API handler - using commands layer
result, err := commands.Browse(commands.BrowseOptions{...})
```

### 2. Using Magic Values
**❌ WRONG:**
```go
outputPath := name + "_filled.pdf"
if err := os.MkdirAll(dir, 0755); err != nil {
```

**✅ CORRECT:**
```go
outputPath := name + FilledPDFSuffix
if err := os.MkdirAll(dir, DefaultDirPerm); err != nil {
```

### 3. Duplicating Output Path Logic
**❌ WRONG:**
```go
if outputDir == "" {
    base := filepath.Base(dataPath)
    name := base[:len(base)-len(filepath.Ext(base))]
    outputPath = name + "_filled.pdf"
}
```

**✅ CORRECT:**
```go
outputPath := DetermineOutputPath(dataPath, outputDir, FilledPDFSuffix)
```

### 4. Forgetting Events
**❌ WRONG:**
```go
func NewCommand(opts NewCommandOptions) error {
    // Just do work without emitting events
    return doWork()
}
```

**✅ CORRECT:**
```go
func NewCommand(opts NewCommandOptions) error {
    Emit(EventNewStarted, ...)
    err := doWork()
    if err != nil {
        EmitError(EventNewError, err, ...)
        return err
    }
    Emit(EventNewCompleted, ...)
    return nil
}
```

### 5. Generic Error Messages
**❌ WRONG:**
```go
if err != nil {
    EmitError(EventFillError, err, nil)
    return err
}
```

**✅ CORRECT:**
```go
if err != nil {
    EmitStageError(EventFillError, StageFillPDF, err, map[string]interface{}{
        "pdf_path": pdfPath,
    })
    return fmt.Errorf("failed to fill PDF: %w", err)
}
```

### 6. Manual HTTP Error Handling
**❌ WRONG:**
```go
if r.Method != "POST" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
}
```

**✅ CORRECT:**
```go
if !httputil.ValidateMethod(w, r, "POST") {
    return
}
```

### 7. Hardcoded Paths
**❌ WRONG:**
```go
catalog, err := pdfform.LoadFormsCatalog("./data/catalog/forms.csv")
```

**✅ CORRECT:**
```go
catalog, err := pdfform.LoadFormsCatalog(h.config.CatalogFilePath())
```

## Maintainability Checklist

Before submitting code, verify:

- [ ] All business logic is in `commands/` package
- [ ] All magic values are in `constants.go`
- [ ] Duplicate code is in `helpers.go`
- [ ] HTTP handlers use `httputil` package
- [ ] Commands emit started/completed/error events
- [ ] Error events include stage information
- [ ] Progress events include progress value (0.0-1.0)
- [ ] Using config helper methods for paths
- [ ] No hardcoded file suffixes or permissions
- [ ] Interface layer only calls commands layer

## Future Enhancements

1. **NATS Jetstream**: Event persistence and linearizability
2. **Datastar Integration**: Real-time reactive UI (check v1.0.0-RC.6+)
3. **Context Support**: Add `context.Context` to all command options
4. **Database Backend**: Replace JSON files with SQLite
5. **Multi-User**: Add authentication and user isolation
6. **Cloud Storage**: Support S3/GCS for PDFs
7. **Form Validation**: Validate field values before filling
8. **Batch Operations**: Fill multiple forms at once
9. **PDF Preview**: Show PDF preview in web GUI
10. **Field Mapping**: Auto-map common field names
