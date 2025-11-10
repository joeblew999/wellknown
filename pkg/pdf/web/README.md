# PDF Form Web Server

Easy-to-mount web server for PDF form filling with auto-discovery, HTTPS support, and separated API/GUI concerns.

## Quick Start - Easy Mounting

The web package is designed to be easily mounted in external projects with minimal code:

### Option 1: Zero-Config Start (HTTPS, Auto-Discovery)
```go
import "github.com/joeblew999/wellknown/pkg/pdf/web"

func main() {
    web.Start(8080)
}
```

This will:
- Auto-discover the `.data` directory
- Generate HTTPS certificates automatically (using mkcert)
- Start server on port 8080 with full GUI and API
- Display all accessible URLs (localhost + LAN IPs)

### Option 2: HTTP Mode (for development)
```go
import "github.com/joeblew999/wellknown/pkg/pdf/web"

func main() {
    web.StartHTTP(8080)
}
```

### Option 3: Custom Data Directory
```go
import "github.com/joeblew999/wellknown/pkg/pdf/web"

func main() {
    web.StartWithConfig(8080, "/custom/data/path")
}
```

### Option 4: Full Control
```go
import "github.com/joeblew999/wellknown/pkg/pdf/web"

func main() {
    // port, dataDir, https
    web.StartWithOptions(8080, "/custom/path", true)
}
```

## Architecture - Separated Concerns

The web package is organized for clean separation and future extensibility:

```
web/
├── server.go          # Main server with easy mounting functions
├── templates.go       # (Legacy - to be removed)
├── api/
│   └── handlers.go    # REST API handlers (/api/*)
├── gui/
│   ├── handlers.go    # GUI handlers (/, /1-browse, etc.)
│   └── templates/     # HTML templates (will be replaced with htmx)
└── README.md          # This file
```

### API Package (`web/api`)
- Handles all `/api/*` endpoints
- JSON responses
- Stateless REST API
- No HTML rendering
- See [API.md](API.md) for full API documentation

### GUI Package (`web/gui`)
- Handles all GUI routes (`/`, `/1-browse`, `/2-download`, etc.)
- HTML rendering
- Currently uses basic templates
- **Future**: Will be replaced with htmx for reactive UI

## Advanced Usage

### Accessing Individual Components

If you need more control, you can use the lower-level API:

```go
import (
    pdfform "github.com/joeblew999/wellknown/pkg/pdf"
    "github.com/joeblew999/wellknown/pkg/pdf/web"
)

// Create custom config
dataDir := pdfform.FindDataDir()  // Or use custom path
config := pdfform.NewConfig(dataDir)

// Create and start server
server := web.NewServer(8080, config, true)
server.Start()
```

### Mounting API-Only or GUI-Only

```go
import (
    "net/http"
    pdfform "github.com/joeblew999/wellknown/pkg/pdf"
    "github.com/joeblew999/wellknown/pkg/pdf/web/api"
    "github.com/joeblew999/wellknown/pkg/pdf/web/gui"
)

func main() {
    config := pdfform.GetDefaultConfig()
    mux := http.NewServeMux()

    // Option 1: API only
    apiHandler := api.NewHandler(config)
    apiHandler.RegisterRoutes(mux)

    // Option 2: GUI only
    gui.InitTemplates()
    guiHandler := gui.NewHandler(config)
    guiHandler.RegisterRoutes(mux)

    http.ListenAndServe(":8080", mux)
}
```

## HTTPS & Certificate Management

When using HTTPS (default), the server will:

1. Check if mkcert is installed
2. Install mkcert via `go install` if needed
3. Generate certificates for localhost + all LAN IPs
4. Store certificates in `.data/certs/`
5. Auto-trust certificates on desktop browsers
6. Provide simple iOS trust workflow (just visit root URL in Safari)

### Certificate Commands

If using the CLI directly:

```bash
pdfform certs info       # Show certificate information
pdfform certs generate   # Generate new certificates
pdfform certs regenerate # Regenerate existing certificates
```

## Configuration Discovery

The `FindDataDir()` function searches in this order:

1. `PDFFORM_DATA_DIR` environment variable
2. `/app/.data` (Docker environment)
3. `.data` in current directory
4. `.data` in parent directories (up to 3 levels)
5. `.data` (default fallback)

This makes the web server work seamlessly in:
- Development environments
- Docker containers
- Production deployments
- External projects

## Network Accessibility

The server automatically:
- Detects all local network interfaces
- Displays localhost URLs
- Displays LAN URLs for mobile access
- Shows iOS-specific setup instructions when HTTPS is enabled

Example output:
```
🌐 HTTPS Server started on port 8080

📱 Access from your devices:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🖥️  Local (this computer)
     https://localhost:8080

  🔗 Local (IP address)
     https://127.0.0.1:8080

  📡 LAN (from other devices)
     https://192.168.1.100:8080

🔒 HTTPS is enabled - certificates in .data/certs/

📱 iOS Setup:
   1. Open Safari on your iPhone/iPad
   2. Visit any URL above
   3. Tap 'Show Details' > 'visit this website'
   4. Done! Certificate is now trusted
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Migration Guide

### Old Code (Before Refactoring)
```go
import (
    pdfform "github.com/joeblew999/wellknown/pkg/pdf"
    "github.com/joeblew999/wellknown/pkg/pdf/web"
)

dataDir := pdfform.FindDataDir()
config := pdfform.NewConfig(dataDir)
web.StartServer(8080, config, true)
```

### New Code (After Refactoring)
```go
import "github.com/joeblew999/wellknown/pkg/pdf/web"

web.Start(8080)
```

The old `StartServer()` function is still available but deprecated in favor of the simpler mounting functions.

## Future Plans

- **GUI Package**: Replace basic templates with htmx for reactive UI
- **API Package**: Remains stable, JSON-only REST API
- **Web Package**: Focus on easy mounting and zero-config deployment

This separation allows you to:
1. Use API-only for headless deployments
2. Replace GUI with your own frontend
3. Extend either package independently
