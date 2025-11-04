# Wellknown PocketBase Server

Server-side Google Calendar access via OAuth 2.0

## CLAUDE 

MUST keep Makefile, .gitignore, code and air config in sync !!! 

## Quick Start

```bash
# Run server (migrations auto-apply)
make server

# Or with hot-reload
make dev
```

**URLs:**
- Root: http://localhost:8090/
- Admin UI: http://localhost:8090/_/
- Google Login: http://localhost:8090/auth/google

## Data Management Flow

### 1. Migrations (Schema Source of Truth)

```
pb/cmd/pb_migrations/
└── 1730709600_init_google_tokens.go  ← Schema definition
```

**Create new migration:**
```go
package pb_migrations

import "github.com/pocketbase/pocketbase/core"

func init() {
    core.AppMigrations.Register(
        // Up: Create collection
        func(txApp core.App) error {
            collection := core.NewBaseCollection("my_collection")
            collection.Fields.Add(
                &core.TextField{Name: "my_field", Required: true},
            )
            return txApp.Save(collection)
        },
        // Down: Remove collection
        func(txApp core.App) error {
            collection, _ := txApp.FindCollectionByNameOrId("my_collection")
            return txApp.Delete(collection)
        },
    )
}
```

### 2. Code Generation (Type-Safe Models)

**After adding/modifying migrations:**

```bash
# Step 1: Run server to apply migrations
make server  # Creates pb_data/ from migrations

# Step 2: Generate template from database schema
make gen-template  # Creates codegen/_templates/*.go

# Step 3: Generate type-safe models
make gen-models  # Creates codegen/models/*.go
```

**Flow:**
```
Migrations → pb_data/ → pocketbase-gogen template → _templates/*.go
                                                           ↓
                                               pocketbase-gogen generate
                                                           ↓
                                               codegen/models/*.go
```

**Key Points:**
- Migrations define ONLY custom fields (no system fields like `id`, `email`)
- Template includes system fields WITH `// system:` comments
- Generator skips system fields (inherited from `core.BaseRecordProxy`)
- No shadowing errors!

### 3. Using Generated Models

```go
import "github.com/joeblew999/wellknown/pb/codegen/models"

// Create record
tokenProxy := models.NewGoogleTokens(app, record)

// Use type-safe setters for CUSTOM fields
tokenProxy.SetUserId("user123")
tokenProxy.SetAccessToken("token")

// System fields inherited from Record
record.Id()  // NOT tokenProxy.Id() - uses core.Record method
```

## Architecture

```
pb/
├── cmd/
│   ├── main.go              # Entry point
│   └── pb_migrations/       # Schema evolution
├── codegen/
│   ├── _templates/          # Generated from pb_data
│   └── models/              # Generated type-safe code
├── oauth.go                 # Google OAuth flow
├── calendar.go              # Calendar API
└── wellknown.go             # App wrapper
```

## Commands

```bash
make help              # Show all commands
make server            # Run server
make dev               # Run with hot-reload (Air)
make build             # Build binary
make gen-template      # Generate template from pb_data
make gen-models        # Generate type-safe models
make clean             # Clean generated files
make release           # Create GitHub release
make debug             # Show Makefile variables
```

## API Endpoints

**OAuth:**
- `GET /auth/google` - Login
- `GET /auth/google/callback` - OAuth callback
- `GET /auth/logout` - Logout
- `GET /auth/status` - Check auth status

**Calendar (authenticated):**
- `GET /api/calendar/events` - List events
- `POST /api/calendar/events` - Create event

## Plugins Enabled

- **jsvm**: JS hooks (`pb_hooks/*.pb.js`)
- **migratecmd**: Database migrations
- **ghupdate**: Self-update from GitHub

See [cmd/main.go](cmd/main.go) for plugin configuration.
