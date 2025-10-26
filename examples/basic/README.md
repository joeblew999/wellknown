# Basic Web Demo

Interactive web-based demo for the wellknown library.

## What It Does

- Provides a web form to create Google Calendar events
- Generates deep links using the wellknown library
- Allows you to test deep links directly from your browser

## Running the Demo

```bash
# Default port (8080)
go run main.go

# Custom port
go run main.go -port 3000
```

Then open http://localhost:8080 in your browser.

## How to Use

1. Fill in the event details (title, start time, end time, etc.)
2. Click "Generate Deep Link"
3. Click "Open in Google Calendar" to test the link on your device
4. Or copy the URL to share it

## Implementation Details

- Uses stdlib `net/http` for web server (zero dependencies)
- HTML template embedded with `go:embed`
- Configurable port via flag
- Demonstrates integration of the wellknown library

## Generated Deep Links

The deep links use the `googlecalendar://` URL scheme which:
- Opens Google Calendar app on mobile devices (iOS/Android)
- Pre-fills event details from the URL parameters
- Works if Google Calendar app is installed
