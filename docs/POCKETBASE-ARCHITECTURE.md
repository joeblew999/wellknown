# Pocketbase Integration Architecture

**Pattern**: Importable library + demo server (inspired by Presentator)

**Date**: 2025-10-27

---

## Design Goals

1. **Reusable Library**: Other PB projects can import our Google Calendar integration
2. **Demo Server**: We can test it standalone in this project
3. **Separation**: Keep PB code separate from core library

---

## Proposed Structure

```
wellknown/
├── pb/                          # NEW: Pocketbase library (root module)
│   ├── wellknown.go            # Main entry point
│   ├── hooks.go                # PB lifecycle hooks
│   ├── google_auth.go          # Google OAuth integration
│   ├── calendar_routes.go      # Calendar API routes
│   ├── middleware.go           # Auth middleware
│   └── base/                   # Demo server
│       └── main.go             # Standalone PB server
│
├── pkg/                         # Core library (no PB dependency!)
│   ├── google/calendar/        # URL generation
│   ├── google/console/         # Calendar API client
│   └── server/                 # Demo web server (no auth)
│
├── cmd/
│   └── server/                 # Demo web server (keep as-is)
│
└── go.mod                      # Core library dependencies
```

---

## Pattern Learned from Presentator

### Library Code (Root Package)

```go
// pb/wellknown.go
package wellknown

import "github.com/pocketbase/pocketbase"

type Wellknown struct {
    *pocketbase.PocketBase
}

func New() *Wellknown {
    wk := &Wellknown{pocketbase.New()}

    // Register all hooks and routes
    bindAppHooks(wk)

    return wk
}
```

### Demo Server

```go
// pb/base/main.go
package main

import "github.com/joeblew999/wellknown/pb"

func main() {
    app := wellknown.New()

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Usage by Other Projects

```go
// In another Pocketbase project:
package main

import (
    "github.com/pocketbase/pocketbase"
    "github.com/joeblew999/wellknown/pb"
)

func main() {
    app := pocketbase.New()

    // Add wellknown functionality
    wk := wellknown.NewWithApp(app)
    wk.RegisterRoutes()

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

---

## File Breakdown

### `pb/wellknown.go` - Main Entry Point

```go
package wellknown

import "github.com/pocketbase/pocketbase"

type Wellknown struct {
    *pocketbase.PocketBase
}

// New creates a standalone Wellknown app
func New() *Wellknown {
    return NewWithApp(pocketbase.New())
}

// NewWithApp attaches to existing PB app (for importing)
func NewWithApp(app *pocketbase.PocketBase) *Wellknown {
    wk := &Wellknown{app}
    bindAppHooks(wk)
    return wk
}
```

### `pb/hooks.go` - Lifecycle Hooks

```go
package wellknown

func bindAppHooks(wk *Wellknown) {
    // Register Google OAuth provider
    wk.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        return setupGoogleAuth(wk)
    })

    // Register calendar routes
    wk.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        return registerCalendarRoutes(wk, e)
    })
}
```

### `pb/google_auth.go` - OAuth Integration

```go
package wellknown

func setupGoogleAuth(wk *Wellknown) error {
    // Configure Google OAuth2 provider
    // Store tokens in PB collections
    // Handle token refresh
}
```

### `pb/calendar_routes.go` - API Routes

```go
package wellknown

func registerCalendarRoutes(wk *Wellknown, e *core.ServeEvent) error {
    // GET /api/calendar/events - List user's events
    e.Router.GET("/api/calendar/events", listEvents, requireAuth())

    // POST /api/calendar/events - Create event
    e.Router.POST("/api/calendar/events", createEvent, requireAuth())

    // Uses pkg/google/console/ for Calendar API calls
}
```

### `pb/middleware.go` - Auth Middleware

```go
package wellknown

func requireAuth() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Check PB auth token
            // Verify Google OAuth token
            // Return 401 if invalid
        }
    }
}
```

---

## Module Structure

### Option A: Separate Go Module (Recommended)

```
wellknown/           # Core library (no PB)
└── go.mod          # Dependencies: no pocketbase

wellknown/pb/       # PB extension (separate module)
└── go.mod          # Dependencies: pocketbase + wellknown core
```

**Benefits**:
- Core library has no PB dependency
- Users can import core without PB overhead
- Clear separation

### Option B: Single Module with Build Tags

```
wellknown/
├── go.mod          # All dependencies
└── pb/
    └── *.go       # //go:build pocketbase
```

**Benefits**:
- Simpler for development
- Users opt-in with build tag

---

## Environment Variables

### Demo Server (`pb/base/main.go`)

```bash
# Google OAuth
GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=xxx

# Pocketbase
PB_ADMIN_EMAIL=admin@example.com
PB_ADMIN_PASSWORD=secure-password
```

---

## API Endpoints (After Integration)

### Public (No Auth)
- `GET /` - Homepage
- `GET /google/calendar` - URL generation demo
- `GET /apple/calendar` - ICS generation demo

### Protected (Requires Auth)
- `POST /api/auth/google` - Initiate Google OAuth
- `GET /api/auth/google/callback` - OAuth callback
- `GET /api/calendar/events` - List user's events
- `POST /api/calendar/events` - Create event
- `PUT /api/calendar/events/:id` - Update event
- `DELETE /api/calendar/events/:id` - Delete event

---

## Benefits of This Architecture

### ✅ Reusable
Other PB projects can:
```bash
go get github.com/joeblew999/wellknown/pb
```

### ✅ Testable
We have standalone demo server at `pb/base/main.go`

### ✅ Clean Separation
- `pkg/` = core library (no PB)
- `pb/` = PB extension (imports `pkg/`)
- `cmd/server/` = demo web server (no auth)

### ✅ Flexible
Users can:
- Import full `wellknown.New()` app
- Or use `NewWithApp(existingPB)` to add to existing PB app

---

## Implementation Phases

### Phase 1: Structure Setup
- Create `pb/` directory
- Create `pb/wellknown.go` skeleton
- Create `pb/base/main.go` demo server

### Phase 2: Google OAuth
- Implement `google_auth.go`
- Configure OAuth provider
- Token storage in PB collections

### Phase 3: Calendar Routes
- Implement `calendar_routes.go`
- Connect to `pkg/google/console/`
- CRUD operations

### Phase 4: Testing
- Test standalone PB server
- Test importing into another PB app
- Integration tests with real Calendar API

---

## Related Documents

- [ARCHITECTURE.md](ARCHITECTURE.md) - Two-server design
- [docs/pocketbase-setup.md](pocketbase-setup.md) - Setup guide (TODO)

**Last Updated**: 2025-10-27
