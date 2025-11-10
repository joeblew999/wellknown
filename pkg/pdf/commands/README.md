# PDF Form Commands Package

**Single Source of Truth for All PDF Form Operations**

The `commands/` package is the central business logic layer that all interfaces (CLI, API, GUI) use. It provides event-driven commands with real-time notifications for building reactive UIs.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     User Interfaces                         │
│  ┌──────────┐    ┌──────────┐    ┌────────────────────┐   │
│  │   CLI    │    │  REST API │    │  htmx GUI (future) │   │
│  └────┬─────┘    └─────┬────┘    └──────┬─────────────┘   │
│       │                 │                 │                  │
│       └─────────────────┼─────────────────┘                 │
│                         │                                    │
│                         ▼                                    │
│           ┌─────────────────────────────┐                   │
│           │    commands/ Package        │                   │
│           │  (Single Source of Truth)   │                   │
│           └─────────────────────────────┘                   │
│                         │                                    │
│              ┌──────────┴──────────┐                        │
│              │                     │                         │
│              ▼                     ▼                         │
│     ┌────────────────┐    ┌────────────────┐               │
│     │ Business Logic  │    │  Event System  │               │
│     │   (Commands)    │    │  (Pub/Sub Bus) │               │
│     └────────────────┘    └────────────────┘               │
│              │                     │                         │
│              └──────────┬──────────┘                        │
│                         ▼                                    │
│              ┌─────────────────────┐                        │
│              │    pkg/pdf (Core)    │                       │
│              │  Low-level PDF ops   │                       │
│              └─────────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

## Why Commands + Events?

### 1. **Single Source of Truth**
All business logic lives in `commands/`. No duplication between CLI/API/GUI.

### 2. **Event-Driven Architecture**
Every command emits events:
- `browse.started`, `browse.completed`, `browse.error`
- `download.started`, `download.progress`, `download.completed`
- `inspect.started`, `inspect.completed`
- `fill.started`, `fill.completed`
- `case.created`, `case.loaded`, `case.updated`

### 3. **Real-Time Updates**
Perfect for htmx + datastar reactive UIs:
```go
// Subscribe to download events
ch := commands.Subscribe("download.*")
defer commands.Unsubscribe(ch)

// Start download in background
go commands.Download(opts)

// Stream SSE updates to browser
for event := range ch {
    fmt.Fprintf(w, "data: %s\n\n", event.ToJSON())
    w.(http.Flusher).Flush()
}
```

### 4. **Observable Operations**
Events enable:
- Real-time progress bars
- Live logs and monitoring
- Webhooks and integrations
- Audit trails
- Analytics

## Available Commands

### Browse Command
```go
result, err := commands.Browse(commands.BrowseOptions{
    CatalogPath: "/path/to/catalog.csv",
    State:       "QLD", // or "" for all states
})

// Events: browse.started, browse.completed, browse.error
```

### Download Command
```go
result, err := commands.Download(commands.DownloadOptions{
    CatalogPath: "/path/to/catalog.csv",
    FormCode:    "F3520",
    OutputDir:   "/path/to/downloads",
})

// Events: download.started, download.progress, download.completed, download.error
// Progress events include: stage, progress percentage
```

### Inspect Command
```go
result, err := commands.Inspect(commands.InspectOptions{
    PDFPath:   "/path/to/form.pdf",
    OutputDir: "/path/to/templates",
})

// Events: inspect.started, inspect.completed, inspect.error
```

### Fill Command
```go
result, err := commands.Fill(commands.FillOptions{
    DataPath:  "/path/to/data.json",
    OutputDir: "/path/to/outputs",
    Flatten:   true,
})

// Events: fill.started, fill.completed, fill.error
```

### Case Management Commands
```go
// Create case
c, casePath, err := commands.CreateCase(formCode, caseName, entityName, dataDir)
// Events: case.created, case.error

// List cases
cases, err := commands.ListCases(dataDir, entityName)
// No events (read-only)

// Load case
c, err := commands.LoadCase(casePath)
// Events: case.loaded, case.error

// Save case
err := commands.SaveCase(c, casePath)
// Events: case.updated, case.error

// Find case by ID
casePath, err := commands.FindCaseByID(caseID, dataDir)
// No events (helper function)
```

## Event System

### Event Types

```go
const (
    // Browse
    EventBrowseStarted   EventType = "browse.started"
    EventBrowseCompleted EventType = "browse.completed"
    EventBrowseError     EventType = "browse.error"

    // Download
    EventDownloadStarted   EventType = "download.started"
    EventDownloadProgress  EventType = "download.progress"
    EventDownloadCompleted EventType = "download.completed"
    EventDownloadError     EventType = "download.error"

    // Inspect
    EventInspectStarted   EventType = "inspect.started"
    EventInspectCompleted EventType = "inspect.completed"
    EventInspectError     EventType = "inspect.error"

    // Fill
    EventFillStarted   EventType = "fill.started"
    EventFillCompleted EventType = "fill.completed"
    EventFillError     EventType = "fill.error"

    // Cases
    EventCaseCreated EventType = "case.created"
    EventCaseLoaded  EventType = "case.loaded"
    EventCaseUpdated EventType = "case.updated"
    EventCaseError   EventType = "case.error"
)
```

### Event Structure

```go
type Event struct {
    Type      EventType              `json:"type"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
    Error     error                  `json:"error,omitempty"`
}
```

### Subscribe to Events

```go
// Subscribe to specific event
ch := commands.Subscribe("download.completed")

// Subscribe to all events of a type (wildcard)
ch := commands.Subscribe("download.*")

// Subscribe to all events
ch := commands.Subscribe("*")

// Always unsubscribe when done
defer commands.Unsubscribe(ch)

// Process events
for event := range ch {
    log.Printf("Event: %s at %s", event.Type, event.Timestamp)
    log.Printf("Data: %+v", event.Data)
}
```

### Emit Custom Events

```go
// Emit success event
commands.Emit(commands.EventDownloadCompleted, map[string]interface{}{
    "form_code": "F3520",
    "file_size": 1024000,
})

// Emit error event
commands.EmitError(commands.EventDownloadError, err, map[string]interface{}{
    "form_code": "F3520",
    "stage": "fetch_pdf",
})
```

## Integration Examples

### CLI Integration
```go
// cli/cli.go
import "github.com/joeblew999/wellknown/pkg/pdf/commands"

func downloadCommand(formCode string) error {
    result, err := commands.Download(commands.DownloadOptions{
        CatalogPath: config.CatalogFilePath(),
        FormCode:    formCode,
        OutputDir:   config.DownloadsPath(),
    })

    if err != nil {
        return err
    }

    fmt.Printf("Downloaded: %s\n", result.PDFPath)
    return nil
}
```

### API Integration
```go
// web/api/handlers.go
import "github.com/joeblew999/wellknown/pkg/pdf/commands"

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
    formCode := r.FormValue("form_code")

    result, err := commands.Download(commands.DownloadOptions{
        CatalogPath: h.config.CatalogFilePath(),
        FormCode:    formCode,
        OutputDir:   h.config.DownloadsPath(),
    })

    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    json.NewEncoder(w).Encode(result)
}
```

### Future htmx GUI Integration
```go
// web/gui/handlers.go with SSE
func (h *Handler) HandleDownloadStream(w http.ResponseWriter, r *http.Request) {
    // Set headers for SSE
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Subscribe to download events
    ch := commands.Subscribe("download.*")
    defer commands.Unsubscribe(ch)

    formCode := r.FormValue("form_code")

    // Start download in background
    go commands.Download(commands.DownloadOptions{
        CatalogPath: h.config.CatalogFilePath(),
        FormCode:    formCode,
        OutputDir:   h.config.DownloadsPath(),
    })

    // Stream events to browser
    for event := range ch {
        fmt.Fprintf(w, "data: %s\n\n", event.ToJSON())
        w.(http.Flusher).Flush()

        // Stop after completion or error
        if event.Type == commands.EventDownloadCompleted ||
           event.Type == commands.EventDownloadError {
            break
        }
    }
}
```

## Benefits

### ✅ For Developers
- **DRY**: Write business logic once, use everywhere
- **Testable**: Commands are pure functions, easy to test
- **Observable**: Events make debugging and monitoring easy
- **Type-safe**: Strong typing with Go structs

### ✅ For Users
- **Real-time feedback**: Progress bars, live updates
- **Responsive UI**: No page refreshes needed (with htmx)
- **Better UX**: Know what's happening at all times

### ✅ For Operations
- **Audit trails**: All events logged automatically
- **Monitoring**: Subscribe to events for metrics
- **Webhooks**: Trigger external systems on events
- **Analytics**: Track usage patterns

## Future Enhancements

- [ ] Add `test` command with events
- [ ] Add workflow command (multi-step operations)
- [ ] Add batch processing commands
- [ ] Event persistence (for audit trails)
- [ ] Event filtering and transformation
- [ ] Webhook support for external integrations
- [ ] Metrics and analytics from events

## Files

```
commands/
├── README.md          # This file
├── events.go          # Event system (pub/sub bus)
├── browse.go          # Browse forms catalog
├── download.go        # Download forms with progress
├── inspect.go         # Inspect PDF fields
├── fill.go            # Fill PDF forms
└── cases.go           # Case management
```

## Related Documentation

- [API Documentation](../web/API.md) - REST API endpoints
- [Web Package](../web/README.md) - Web server architecture
- [Core Package](../README.md) - Low-level PDF operations
