# Status and Ideas

## 🎉 Major Milestones

### Phase 5: Full Platform Separation (Completed 2025-10-27)

**Problem**: Google Calendar and Apple Calendar have fundamentally different capabilities:
- Google Calendar uses web URLs with 5 basic fields only (limited by URL length)
- Apple Calendar uses ICS format with full RFC 5545 spec (recurring events, attendees, reminders, etc.)

**Solution**: Complete architectural separation
- ✅ Created `pkg/google/calendar/` with simple 5-field Event type
- ✅ Created `pkg/apple/calendar/` with full ICS Event type + advanced types (Recurrence, Attendee, etc.)
- ✅ Platform-specific examples (Google: 6 basic, Apple: 6 basic + advanced)
- ✅ Updated handlers to use `interface{}` with platform-specific type assertions
- ✅ Deleted old shared types (`pkg/types/calendar.go`, `pkg/examples/calendar.go`)
- ✅ Both showcase pages working and tested

**Benefits**:
- Each platform can evolve independently
- Full access to platform-native features
- No compromises or lowest-common-denominator APIs
- Clear separation of concerns

**Files Changed**:
- Created: `pkg/google/calendar/{event.go, examples.go}`
- Created: `pkg/apple/calendar/{types.go, event.go, examples.go}`
- Updated: `examples/basic/handlers/{generic.go, google_calendar.go, apple_calendar.go}`
- Deleted: `pkg/types/calendar.go`, `pkg/examples/calendar.go`, `pkg/google/calendar.go`, `pkg/apple/calendar.go`
- Updated: `CLAUDE.md` with full separation architecture documentation

---

### Phase 6: Server Infrastructure Refactoring (Completed 2025-10-27)

**Problem**: Web server was in `examples/basic/` but it's actually ESSENTIAL infrastructure for testing deep links on real devices (not optional demo code).

**Solution**: Moved server to core library infrastructure
- ✅ Created `pkg/server/` with embedded templates (via `//go:embed`)
- ✅ Moved all handlers from `examples/basic/handlers/` to `pkg/server/`
- ✅ Created minimal `cmd/wellknown-server/main.go` (18 lines)
- ✅ Fixed test infrastructure (platform-specific tests in correct packages)
- ✅ Updated `go.work` to remove deleted `examples/basic/`
- ✅ Deleted obsolete `examples/basic/` directory
- ✅ Updated all documentation (README.md, CLAUDE.md, examples/README.md)

**Benefits**:
- Server is recognized as essential testing infrastructure
- Templates embedded in binary (zero-config deployment)
- Clean separation: `pkg/server/` (library), `cmd/wellknown-server/` (binary)
- Tests properly organized in platform-specific packages
- Air hot-reload still works from new location

**Files Changed**:
- Created: `pkg/server/{server.go, handlers.go, generic.go, google_calendar.go, apple_calendar.go, stub.go}`
- Created: `pkg/server/templates/*.html` (embedded via `//go:embed`)
- Created: `cmd/wellknown-server/{main.go, .air.toml}`
- Created: `pkg/google/calendar/{event_test.go, testdata.go}`
- Created: `pkg/apple/calendar/event_test.go`
- Updated: `go.work` (removed `examples/basic/` reference)
- Updated: `README.md`, `CLAUDE.md`, `examples/README.md`
- Deleted: `examples/basic/` (entire directory)
- Deleted: Old test files at wrong level (`pkg/google/calendar_test.go`, etc.)

**Test Results**:
- All 36 tests passing (24 Google Calendar, 12 Apple Calendar)
- Server builds successfully (`11MB` binary)
- Air hot-reload working from `cmd/wellknown-server/`
- Playwright tests passing

---

## Ideas and Future Work

### JSON Schema Forms

think about https://github.com/gedw99/goPocJsonSchemaForm

https://github.com/warlockxins/goPocJsonSchemaForm

---

Pocketbase

think aboout Pocketbase auth so then we can see the end users actual cal and know if its an update or create or delete. https://github.com/presentator/presentator has base code thats good.

so we need a pkg/pocketbase and a google cloud console and login screens.

---

timelinze

https://github.com/timelinize/timelinize

https://github.com/joeblew999/timelinize-plug

they have stuff related.








