# wellknown Examples

This directory contains example programs demonstrating the wellknown library.

## Examples

### 1. Basic (`./basic`)
Demonstrates native deep link generation for:
- Google Calendar events
- Apple Calendar events
- Google Maps navigation
- Apple Maps navigation
- Google Drive files
- Apple Files/iCloud
- Email (universal)

**Run:**
```bash
go run ./basic/main.go
```

### 2. WebView (`./mcp`)
Demonstrates claude code calling out MCP.

**Run:**
```bash
go run ./mcp/main.go
```

### 3. WebView (`./webview`)
Demonstrates web fallback URLs for scenarios where native apps aren't available:
- Google Calendar web interface
- Google Maps web interface
- Google Drive web interface
- iCloud web interface

**Run:**
```bash
go run ./webview/main.go
```

## Workspace Setup

This project uses Go workspaces. The `go.work` file in the project root coordinates all modules:

```
go.work
├── . (main module)
├── ./examples/basic
└── ./examples/webview
```

## Development

To add a new example:
1. Create a new directory under `examples/`
2. Add `go.mod` with proper module name and replace directive
3. Create `main.go` with your example code
4. Add the directory to `go.work`

Example go.mod:
```go
module github.com/joeblew999/wellknown/examples/myexample

go 1.23

replace github.com/joeblew999/wellknown => ../..
```
