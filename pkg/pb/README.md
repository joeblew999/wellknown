# Wellknown PocketBase Server


## Quick Start

```bash

# 1. Create migration
# Edit: pb/cmd/pb_migrations/TIMESTAMP_name.go

# 2. run
make run   # Migrations auto-apply

# 3. will pick up new schema in PB and gen 
make gen  # Regenerate templates from pb_data and Regenerate models from new templates

# 4. run
make run   # new golang code now used.

```

**URLs:**
- Root: http://localhost:8090/
- Admin UI: http://localhost:8090/_/
- Google Login: http://localhost:8090/auth/google


## Development Flow (CORRECT - Discovered via Banking Example)

### 1. Migrations (Schema Source of Truth)

```
pb/cmd/pb_migrations/
├── 1730709600_init_google_tokens.go  ← OAuth tokens
└── 1730710000_init_banking.go        ← Banking example
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

**Migration timestamps MUST be sequential!**

### 2. Code Generation (Type-Safe Models)

**IMPORTANT**: `make gen-models` works as long as templates exist in `codegen/_templates/`. You DON'T need pb_data to generate models!

```bash
# Generate models from existing templates
make gen-models  # Creates codegen/models/*.go
```

**Only regenerate templates when adding NEW collections:**
```bash
# After running server with new migration:
make gen-template  # Extracts schema from pb_data → _templates/*.go
make gen-models    # Regenerates models with new collections
```

**Flow:**
```
Migrations (source of truth)
    ↓
pb_data/ (auto-created on server start)
    ↓ (only if NEW collections)
_templates/*.go (editable schema-as-code)
    ↓
codegen/models/*.go (generated type-safe wrappers)
```

**Key Points:**
- Migrations define ONLY custom fields (no system fields like `id`, `email`)
- Templates include system fields WITH `// system:` comments
- Generator skips system fields (inherited from `core.BaseRecordProxy`)
- Templates can exist without pb_data!

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
make clean             # Clean generated files (preserves pb_data)
make release           # Create GitHub release
make debug             # Show Makefile variables
```

**⚠️ Important Notes:**
- `make clean` does NOT delete `pb_data/` (contains database, collections, and admin users)
- If you manually delete `pb_data/`, PocketBase will show the admin setup screen on next start
- To completely reset: `rm -rf cmd/pb_data && make server` (will re-run migrations and prompt for new admin)

## API Endpoints

**OAuth:**
- `GET /auth/google` - Login
- `GET /auth/google/callback` - OAuth callback
- `GET /auth/logout` - Logout
- `GET /auth/status` - Check auth status

**Calendar (authenticated):**
- `GET /api/calendar/events` - List events
- `POST /api/calendar/events` - Create event

**Banking (example feature):**
- `GET /api/banking/accounts?user_id=<id>` - List accounts for user
- `GET /api/banking/accounts/:id` - Get account details
- `GET /api/banking/accounts/:id/transactions` - List account transactions
- `POST /api/banking/accounts` - Create new account
- `POST /api/banking/transactions` - Create transaction (auto-updates balance)

## Plugins Enabled

- **jsvm**: JS hooks (`pb_hooks/*.pb.js`)
- **migratecmd**: Database migrations
- **ghupdate**: Self-update from GitHub

See [cmd/main.go](cmd/main.go) for plugin configuration.
