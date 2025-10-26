# wellknown Examples

This directory contains example programs demonstrating the wellknown library.

## Examples

### 1. Basic (`./basic`)
web demo that demonstrates link generation for

### 2. WebView (`./mcp`)
Demonstrates claude code calling out to the MCP.

### 3. WebView (`./webview`)
Demonstrates web gui running inside a webview.

### 4. Custom (`./custom`)
Demonstrates custom url schema.

## Workspace Setup

This project uses Go workspaces. The `go.work` file in the project root coordinates all modules:

## Development

To add a new example:
1. Create a new directory under `examples/`
2. Add `go.mod` with proper module name and replace directive
3. Create `main.go` with your example code
4. Add the directory to `go.work`

Example go.mod:
```go
module github.com/joeblew999/wellknown/examples/myexample

go 1.25

replace github.com/joeblew999/wellknown => ../..
```
