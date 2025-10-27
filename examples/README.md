# wellknown Examples

This directory contains example programs demonstrating the wellknown library.

**Note**: The main test server has been moved to `cmd/wellknown-server/` as it's essential infrastructure, not just an example.

## Examples

### 1. MCP Server (`./mcp`)
Demonstrates Model Context Protocol (MCP) integration for AI assistants.
Allows Claude Code to generate deep links via MCP tools.

### 2. WebView (`./webview`)
Demonstrates embedding the deep link test UI in a native webview.
Shows how to integrate wellknown server in desktop/mobile apps.

### 3. Custom Schemes (`./custom`)
Demonstrates custom URL schemes and service registration.
Example of extending wellknown for proprietary deep link formats.

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
