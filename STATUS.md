# Project Status

## Current State: PLANNING COMPLETE

**Last Updated**: 2025-10-23

---

## Overview

The `wellknown` project is a Universal Go library for generating and opening deep links across the Google and Apple app ecosystems.

**Repository**: `github.com/joeblew999/wellknown`

**Design & Architecture**: See [CLAUDE.md](CLAUDE.md) for technical decisions, folder structure, API patterns, and MCP integration details.

---

## Completion Status

### ✅ Completed
- [x] Repository initialization
- [x] README.md documentation
- [x] LICENSE file
- [x] .gitignore configuration
- [x] Go module initialization (`go.mod`)
- [x] Initial git commits
- [x] Folder structure design and approval
- [x] Template strategy decision (go:embed + custom support)
- [x] Go workspace decision (go.work for examples)

### 🚧 In Progress
- [ ] Implementation ready to begin

### 📋 Planned - Infrastructure
- [ ] Create folder structure
- [ ] Set up go.work and go.work.example
- [ ] Update .gitignore for go.work files
- [ ] Create internal/template package (renderer, loader)
- [ ] Create pkg/deeplink core interfaces

### 📋 Planned - Google Platform
- [ ] Google Calendar (pkg/deeplink/google/calendar.go + calendar.tmpl)
- [ ] Google Maps (pkg/deeplink/google/maps.go + maps.tmpl)
- [ ] Google Drive (pkg/deeplink/google/drive.go + drive.tmpl)
- [ ] Google web fallbacks (pkg/deeplink/web/google.go + templates)

### 📋 Planned - Apple Platform
- [ ] Apple Calendar (pkg/deeplink/apple/calendar.go + calendar.tmpl)
- [ ] Apple Maps (pkg/deeplink/apple/maps.go + maps.tmpl)
- [ ] Apple Files (pkg/deeplink/apple/files.go + files.tmpl)
- [ ] Apple web fallbacks (pkg/deeplink/web/apple.go + templates)

### 📋 Planned - Additional Features
- [ ] Platform detection interface (pkg/platform/detect.go)
- [ ] Platform detection implementations (real + mock)
- [ ] URL opener interface (pkg/opener/opener.go)
- [ ] URL opener implementations (real + spy)
- [ ] CLI tool (cmd/wellknown/main.go)
- [ ] Examples with go.work setup
- [ ] Custom template validation
- [ ] Documentation (docs/)

### 📋 Planned - Testing (Phase 1: Unit Tests)
- [x] Document testing approach (docs/testing-with-goup-util.md)
- [ ] Write unit tests for URL generation
- [ ] Write unit tests for template rendering
- [ ] Write unit tests for deterministic output
- [ ] Build mock opener/detector for testing
- [ ] Set up test coverage reporting

### 📋 Planned - Testing (Phase 2: Test App with goup-util)
- [ ] Create test app structure (tests/testapp/)
- [ ] Set up Gio UI with test case buttons
- [ ] Integrate wellknown library
- [ ] Use gio-plugins/hyperlink to open deep links
- [ ] Embed webview for results display (gio-plugins/webviewer)
- [ ] Build with goup-util for macOS/iOS/Android

### 📋 Planned - Testing (Phase 3: CI/CD)
- [ ] Set up GitHub Actions workflow
- [ ] Run unit tests on all platforms
- [ ] Build test app with goup-util
- [ ] Add test coverage reporting

### 📋 Planned - MCP Integration
- [ ] Add official MCP Go SDK dependency (github.com/modelcontextprotocol/go-sdk)
- [ ] Implement MCP server (cmd/wellknown-mcp/main.go)
- [ ] Define MCP tools (create_calendar_event, create_maps_link, etc.)
- [ ] Implement resource handlers for template inspection
- [ ] Set up STDIO transport
- [ ] Create Claude Desktop integration config example
- [ ] Test with Claude Desktop
- [ ] Document MCP server usage

---

## Repository Statistics

- **Branch**: main
- **Commits**: 2
- **Go Files**: 0
- **Test Files**: 0
- **Dependencies**: 0 (pure Go, zero deps)

---

## Next Milestones

1. **Milestone 1: Core Library Structure**
   - Define package structure
   - Create URL builder interfaces
   - Implement basic URL generators

2. **Milestone 2: Platform Support**
   - Google ecosystem integration
   - Apple ecosystem integration
   - Web fallback support

3. **Milestone 3: CLI Tool**
   - Command-line interface
   - Platform detection
   - Auto-opening URLs

4. **Milestone 4: Testing & Documentation**
   - Comprehensive test coverage
   - API documentation
   - Usage examples

---

## Known Issues

None currently - project in initial setup phase.

---

## Notes

- Project emphasizes **deterministic** URL generation (same input → same output)
- **Zero external dependencies** requirement must be maintained
- **Cross-platform** compatibility is a core requirement
- Focus on both programmatic API and CLI usability
