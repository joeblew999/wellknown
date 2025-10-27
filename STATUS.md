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

### Phase 7: JSON Schema Dynamic Forms (Completed 2025-10-27)

**Goal**: Auto-generate forms from JSON Schema definitions (inspired by goPocJsonSchemaForm)

**Implementation**:
- ✅ Created `pkg/server/schema.go` - JSON Schema parser (stdlib only!)
- ✅ Created `pkg/google/calendar/schema.go` - Schema definition for Google Calendar
- ✅ Auto-form generation with full type support (string, number, boolean, array, object)
- ✅ Format support (datetime-local, email, uri, date, time)
- ✅ Constraints (minLength, maxLength, minimum, maximum, enum, required)
- ✅ Created route `/google/calendar-schema`
- ✅ Fixed template embedding and rendering

**Benefits**:
- Declarative form definition (define schema once, form generates automatically)
- Type-safe data validation
- Self-documenting API
- Zero external dependencies (stdlib only!)

**Testing**: ✅ Schema-based forms working end-to-end

---

### Phase 8: UI Schema for Layout Control (Completed 2025-10-27)

**Goal**: Separate presentation (UI Schema) from validation (JSON Schema)

**Implementation**:
- ✅ Created `pkg/server/uischema.go` (370 lines) - UI Schema renderer
- ✅ Created `pkg/google/calendar/uischema.go` - Layout definition
- ✅ Layout types: VerticalLayout, HorizontalLayout, Control, Label, Group
- ✅ Options: placeholder, multi-line, format override, showLabel, suggestions
- ✅ Created route `/google/calendar-uischema`
- ✅ Added CSS styling for layouts (responsive horizontal -> vertical on mobile)

**UI Features**:
- Side-by-side start/end time fields (HorizontalLayout)
- Section labels with emoji (Label)
- Grouped fields (Group)
- Custom placeholders and descriptions

**Benefits**:
- Flexible layout control without changing validation
- Reusable UI patterns
- Better UX with proper field grouping

**Testing**: ✅ UI Schema forms rendering correctly with layout control

---

### Phase 9: Server-Side Validation (Completed 2025-10-27)

**Goal**: Validate form submissions and display errors inline

**Phase 9a: Validation Infrastructure** ✅
- Created `pkg/server/validator.go` (280 lines, stdlib only!)
- `ValidateAgainstSchema()` - validates form data against JSON Schema
- Format validation: email, uri, date, datetime-local, time
- Type validation: string, number, integer, boolean
- Constraint validation: required, minLength, maxLength, minimum, maximum, enum
- `FormDataToMap()` - converts flat forms to nested maps (dot notation support)
- No external dependencies!

**Phase 9b: Wire Validation** ✅
- Created `handleGoogleCalendarUISchemaPost()` in `pkg/server/google_calendar.go`
- Validates POST data against schema
- Re-renders form with errors if validation fails
- Only generates URL if validation passes
- Added `ValidationErrors` and `FormData` to PageData struct

**Phase 9c: Display Validation Errors** ✅
- Updated `pkg/server/uischema.go`:
  - Added `GenerateFormHTMLWithData()` method
  - Pre-fills form values from `FormData`
  - Displays validation errors inline below each field
  - All input types support value attributes (text, datetime-local, email, textarea, number)
- Updated `pkg/server/helpers.go` to pass ValidationErrors and FormData to renderer
- Added CSS styling:
  - `.field-error` class: red text with proper spacing
  - Error state styling: red border + pink background on inputs with errors
  - Uses `:has()` selector for automatic error highlighting

**Testing**: ✅ Complete validation UX working
- Submitted form with missing required fields → errors appear below fields
- Submitted form with partial data → values preserved after validation failure
- Error styling applied correctly (red text, red border, pink background)
- Users don't lose their data when validation fails!

**Benefits**:
- Better UX - users see exactly what's wrong
- Form values preserved after validation errors
- Stdlib only implementation (no external validation libraries)
- Follows reference code patterns from goPocJsonSchemaForm

**Routes Strategy** (all 3 approaches working):
1. `/google/calendar` - Hardcoded HTML form ✅
2. `/google/calendar-schema` - JSON Schema only ✅
3. `/google/calendar-uischema` - UI Schema + JSON Schema + Validation ✅

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


---

xtemplate 

https://github.com/infogulch/xtemplate

Check how this might help us. use .src to get the code and research what it gives us







