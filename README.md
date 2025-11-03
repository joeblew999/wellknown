# wellknown

**Universal Go library for generating and opening deep links across the Google and Apple app ecosystems.**  
Pure Go · Zero deps · Deterministic URLs · Cross-platform.

---

## ✨ Overview

`wellknown` lets Go applications and CLIs create **native deep links** and **URL schemes** for common apps such as:

| Category | Google | Apple |
|-----------|---------|--------|
| Calendar | `googlecalendar://render?...` | `calshow:` |
| Maps | `comgooglemaps://?q=` | `maps://?q=` |
| Mail | `mailto:` | `mailto:` |
| Drive / Files | `googledrive://` | `shareddocuments://` |

The library also provides safe fallbacks to open the **web equivalents** when native apps aren’t available.

---

## 🧩 Features

- ✅ **Pure Go** — no external dependencies.  
- 🧠 **Deterministic**: same input → same output (great for reproducible infra / NATS messages).  
- ⚙️ **Cross-platform**: works on macOS, Windows, Linux, iOS, and Android.  
- 🕹 **Programmatic & CLI**: embed in binaries or call from shell scripts.  
- 🔗 **App-aware**: automatically chooses local URL scheme vs. browser fallback.  

---

## 🧱 Installation

```bash
go get github.com/joeblew999/wellknown
```


---

## 🧪 Testing Server

The wellknown library includes a web server for testing deep links on real devices. This is **essential infrastructure**, not just a demo, because deep links can only be properly tested in a browser on mobile devices.

### Development Setup

For hot-reload during development, install Air (optional but recommended):

```bash
go install github.com/air-verse/air@latest
```

### Running the Test Server

```bash
# Development mode with hot-reload (recommended)
make dev

# Or standard mode
go run ./cmd/server

# Or build and run
go build -o wellknown-server ./cmd/server
./wellknown-server
```

### Features

- **Live Testing**: Test deep links on real iOS/Android devices
- **Showcase Pages**: Pre-built examples for each service (Google Calendar, Apple Calendar, etc.)
- **Custom Forms**: Create your own deep links with custom parameters
- **QR Codes**: Generate QR codes for easy mobile testing (coming soon)
- **Hot Reload**: Air automatically rebuilds when code changes

### Mobile Testing

When the server starts, it displays both local and network URLs:

```
🚀 wellknown demo server starting...
💻 Local:  http://localhost:8080
📱 Mobile: http://192.168.1.84:8080
```

Scan the mobile URL from your phone to test deep links on actual devices.

---

## 📦 Package Structure

```
wellknown/
├── pkg/
│   ├── google/calendar/  # Google Calendar (URL-based, 5 fields)
│   ├── apple/calendar/   # Apple Calendar (ICS-based, full RFC 5545)
│   └── server/           # Web server for testing (embedded templates)
├── cmd/
│   └── wellknown-server/ # Test server binary (18 lines)
└── examples/             # Additional examples (MCP, WebView, Custom)


