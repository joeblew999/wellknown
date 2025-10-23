# Testing wellknown with goup-util

**Purpose**: How to use goup-util and gio-plugins to test wellknown deep links

---

## Overview

Testing deep links requires opening URLs and verifying they trigger the correct apps. We use a 3-layer stack:

**Layer 1**: [Gio UI](https://gioui.org) - Pure Go cross-platform UI framework
**Layer 2**: [gio-plugins](https://github.com/gioui-plugins/gio-plugins) - Native OS integration
**Layer 3**: [goup-util](https://github.com/joeblew99/goup-util) - Build tool for cross-platform deployment

---

## goup-util

**Repository**: https://github.com/joeblew999/goup-util
**Local Path**: `/Users/apple/workspace/go/src/github.com/joeblew999/goup-util`
**What it is**: Cross-platform hybrid app build tool (Gio UI + native webviews)

### Capabilities for Testing

1. ✅ **Cross-platform builds** - macOS, iOS, Android, Windows, Linux
2. ✅ **SDK management** - Auto-installs Android NDK, iOS SDK, etc.
3. ✅ **Icon generation** - Generate platform-specific icons from one source
4. ✅ **Screenshot** - via robotgo integration
5. ✅ **Workspace management** - Handle multi-module projects

### Build Commands

```bash
# Build for macOS
goup-util build macos tests/testapp

# Build for iOS
goup-util build ios tests/testapp

# Build for Android (installs NDK if needed)
goup-util build android tests/testapp

# Build for all platforms
goup-util build macos tests/testapp
goup-util build ios tests/testapp
goup-util build android tests/testapp

# Take screenshot
goup-util screenshot test.png

# Screenshot with delay (for menus)
goup-util screenshot --delay 3000 menu.png
```

### Installation

```bash
# Quick install (macOS)
curl -fsSL https://raw.githubusercontent.com/joeblew99/goup-util/main/scripts/macos-bootstrap.sh | bash

# Or build from source
cd /Users/apple/workspace/go/src/github.com/joeblew999/goup-util
go build .
```

### Examples

The goup-util repository contains working examples:

- **gio-basic** - Simple Gio app
- **gio-plugin-hyperlink** - Opens URLs in system browser (key example!)
- **gio-plugin-webviewer** - Multi-tab browser with webview
- **hybrid-dashboard** - Embedded HTTP server + webview

**Location**: `/Users/apple/workspace/go/src/github.com/joeblew999/goup-util/examples/`

---

## gio-plugins

**Repository**: https://github.com/gioui-plugins/gio-plugins
**What it is**: Native platform integrations for Gio apps

### Available Plugins

| Plugin | Purpose | Relevant for Testing? |
|--------|---------|----------------------|
| **hyperlink** | Opens URLs in system browser/handler | ✅ **YES - This opens deep links!** |
| **webviewer** | Embed native webviews | ✅ YES - Display results |
| **share** | Native share dialog | ✅ YES - Shows URL handlers |
| **explorer** | Native file dialogs | ✅ YES - Save reports |
| **safedata** | Secure device storage | No |
| **auth** | OAuth (Google, Apple) | No |
| **pingpong** | System testing | No |

### hyperlink Plugin (Most Important!)

**This is how we open deep links.**

```go
import "github.com/gioui-plugins/gio-plugins/hyperlink/giohyperlink"
import "github.com/gioui-plugins/gio-plugins/plugin/gioplugins"

// In your Gio event loop:
if buttonClicked {
    url := wellknown.GoogleCalendar(event)
    gioplugins.Execute(gtx, giohyperlink.OpenCmd{URI: url})
}
```

**What it does**:
- Opens URL in system default handler
- Triggers registered URL schemes (e.g., `googlecalendar://`)
- Works on all platforms (iOS, Android, macOS, Windows, Linux)

### webviewer Plugin

Embed native webviews to display test results.

```go
import "github.com/gioui-plugins/gio-plugins/webviewer/giowebview"

// Navigate webview
gioplugins.Execute(gtx, giowebview.NavigateCmd{
    View: webviewTag,
    URL:  serverURL,
})
```

**Platform webviews**:
- iOS/macOS: WKWebView
- Android: Chromium WebView
- Windows: WebView2
- Linux: WebKitGTK

### share Plugin

Native share dialog - could indirectly show which apps can handle URLs.

```go
import "github.com/gioui-plugins/gio-plugins/share"

// Share a deep link
share.Share(deepLinkURL)
// OS shows which apps can handle the URL
```

---

## Test App Architecture

### Recommended Structure

```
tests/testapp/
├── main.go              # Gio UI + gio-plugins
├── web/                 # Embedded (go:embed)
│   ├── index.html       # Test case list
│   ├── results.html     # Results display
│   └── styles.css
├── icon-source.png      # Icon for all platforms
├── go.mod
└── README.md
```

### Test Flow

1. **UI Layer** (Gio)
   - Display test cases as buttons
   - "Test Google Calendar", "Test Apple Maps", etc.

2. **Generation** (wellknown library)
   - User clicks button
   - wellknown generates deep link: `googlecalendar://render?...`

3. **Opening** (gio-plugins/hyperlink)
   - `giohyperlink.OpenCmd{URI: url}` opens the URL
   - OS triggers registered app or falls back to browser

4. **Verification**
   - Screenshot (robotgo via goup-util)
   - Manual: user sees if correct app opened
   - Results logged

5. **Results Display** (gio-plugins/webview)
   - Embedded HTTP server serves results
   - WebView displays HTML/JS results
   - Go ↔ JS bridge for data

### Example main.go Pattern

```go
package main

import (
    "embed"
    "net/http"

    "gioui.org/app"
    "gioui.org/layout"
    "gioui.org/widget"
    "github.com/gioui-plugins/gio-plugins/hyperlink/giohyperlink"
    "github.com/gioui-plugins/gio-plugins/plugin/gioplugins"
    "github.com/gioui-plugins/gio-plugins/webviewer/giowebview"

    "github.com/joeblew999/wellknown/pkg/deeplink/google"
)

//go:embed web/*
var webContent embed.FS

func main() {
    // Start embedded HTTP server for results
    serverURL := startWebServer()

    // Launch Gio UI
    go runApp(serverURL)
    app.Main()
}

func runApp(serverURL string) {
    var (
        testGoogleCalBtn = &widget.Clickable{}
        testAppleMapsBtn = &widget.Clickable{}
    )

    window := &app.Window{}

    for {
        evt := gioplugins.Hijack(window)

        switch evt := evt.(type) {
        case app.FrameEvent:
            gtx := app.NewContext(&ops, evt)

            // Handle button clicks
            if testGoogleCalBtn.Clicked(gtx) {
                url := google.Calendar(/* event data */)
                gioplugins.Execute(gtx, giohyperlink.OpenCmd{URI: url})
            }

            // Render UI...
            evt.Frame(gtx.Ops)
        }
    }
}
```

---

## What We CAN Test

- ✅ **URL generation** - Unit tests in wellknown library
- ✅ **URL opening** - giohyperlink opens URLs
- ✅ **Cross-platform** - goup-util builds for all platforms
- ✅ **Visual verification** - Screenshots + manual confirmation
- ✅ **Results collection** - Hybrid UI with webview

---

## What We CANNOT Test Automatically

- ❌ **URL scheme detection** - Can't programmatically check if Google Calendar is installed
- ❌ **Automatic verification** - Can't automatically verify correct app opened
- ❌ **App interaction** - Can't automate clicks inside Calendar app

**Workaround**: Use share plugin to show which apps can handle a URL via OS share sheet.

---

## Building the Test App

### Local Development

```bash
cd tests/testapp

# Build for current platform (macOS)
goup-util build macos .

# Run the app
open .bin/testapp.app
```

### Cross-Platform Builds

```bash
# iOS (requires Xcode)
goup-util build ios tests/testapp

# Android (auto-installs NDK)
goup-util build android tests/testapp

# Deploy to device/simulator
# iOS: Xcode or goup-util deploy commands
# Android: adb install
```

### CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Test

on: [push]

jobs:
  test-app:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install goup-util
        run: go install github.com/joeblew999/goup-util@latest

      - name: Build test app (macOS)
        run: goup-util build macos tests/testapp

      - name: Build test app (iOS)
        run: goup-util build ios tests/testapp
```

---

## Common Patterns

### Opening a Deep Link

```go
// Generate URL with wellknown
url := google.Calendar(event)

// Open it (triggers native app or browser)
gioplugins.Execute(gtx, giohyperlink.OpenCmd{URI: url})
```

### Embedding Web Results

```go
//go:embed web/*
var webContent embed.FS

// Serve embedded content
webFS, _ := fs.Sub(webContent, "web")
http.Handle("/", http.FileServer(http.FS(webFS)))

// API endpoint for results
http.HandleFunc("/api/results", handleResults)
```

### Displaying Results in WebView

```go
// Navigate webview to local server
gioplugins.Execute(gtx, giowebview.NavigateCmd{
    View: webviewTag,
    URL:  "http://127.0.0.1:8080",
})
```

### Taking Screenshots

```bash
# From test app or CLI
goup-util screenshot before.png
# ... open deep link ...
sleep 2
goup-util screenshot after.png
```

---

## Platform-Specific Notes

### macOS
- WKWebView for webviews
- `open` command used by hyperlink plugin
- Screenshot requires Screen Recording permission
  - System Settings → Privacy & Security → Screen Recording

### iOS
- WKWebView
- Info.plist must declare queryable URL schemes
- Code signing required
- Can test in simulator or real device

### Android
- Chromium WebView
- Intent handlers for URL schemes
- NDK auto-installed by goup-util
- Can test in emulator or real device

### Windows
- WebView2
- Registry for URL schemes
- Cross-compilation may have issues (see goup-util docs)

### Linux
- WebKitGTK for webviews
- xdg-open for URLs
- Cross-compilation may have issues (see goup-util docs)

---

## Resources

- **goup-util README**: `/Users/apple/workspace/go/src/github.com/joeblew999/goup-util/README.md`
- **goup-util Examples**: `/Users/apple/workspace/go/src/github.com/joeblew999/goup-util/examples/`
- **gio-plugins Docs**: https://github.com/gioui-plugins/gio-plugins
- **Gio UI Docs**: https://gioui.org
- **hybrid-dashboard Example**: Best reference for embedded HTTP + webview pattern

---

## Troubleshooting

### "command not found: goup-util"

Install goup-util:
```bash
cd /Users/apple/workspace/go/src/github.com/joeblew999/goup-util
go build .
# Add to PATH or use ./goup-util
```

### Android NDK not found

goup-util auto-installs it:
```bash
goup-util install ndk-bundle
```

### Screenshot permissions (macOS)

Grant Screen Recording permission:
System Settings → Privacy & Security → Screen Recording → Enable for Terminal/IDE

### Deep link doesn't open

1. Check if app is installed
2. Use share plugin to see which apps can handle the URL
3. Try opening URL manually in browser to see redirect behavior

---

**Last Updated**: 2025-10-23
