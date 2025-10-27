# Basic Web Demo

Interactive web-based demo for the wellknown library.

## Running the Demo

### Development Mode (with hot-reload)

For development, use Air to automatically rebuild and restart when code changes:

```bash
# Install Air (first time only)
go install github.com/air-verse/air@latest

# Run with Air (default port 8080)
air
```

Air watches for changes to `.go`, `.html`, `.tpl`, and `.tmpl` files and automatically rebuilds/restarts the server.

### Standard Mode

```bash
# Default port (8080)
go run main.go

# Custom port
go run main.go -port 3000
```

Then open http://localhost:8080 in your browser.

## Mobile Testing

When the server starts, it displays both the local and network URLs:

```
💻 Local:  http://localhost:8080
📱 Mobile: http://192.168.1.84:8080
```

Use the mobile URL to test from your phone or tablet on the same network.

