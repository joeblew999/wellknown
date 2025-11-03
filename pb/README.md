# Wellknown PocketBase Integration

**Server-side access to users' Google Calendar data via OAuth 2.0**

This PocketBase integration enables server-side Google Calendar operations:
- Users authenticate via Google OAuth (consent flow)
- OAuth tokens stored securely in PocketBase database
- Server can read/write to users' Google Calendar on their behalf
- Supports listing events, creating events, and more

## Quick Start

```bash
# 1. Setup (optional - only needed for Google OAuth)
cd pb/base
cp .env.example .env
# Edit .env with your Google OAuth credentials

# 2. Run server (auto-creates collections, plugins enabled)
make pb-server
```

Access:
- **Root**: http://localhost:8090/ (dynamic HTML with all endpoints)
- **API Index**: http://localhost:8090/api/ (JSON version)
- **Admin UI**: http://localhost:8090/_/
- **OAuth**: http://localhost:8090/auth/google
- **Calendar API**: http://localhost:8090/api/calendar/events

## Commands

```bash
make help  # See all available commands
```

**Key Commands**:
- `make pb-server` - Run server (no hot-reload)
- `make pb-dev` - Run server with hot-reload (Air)
- `make pb-build` - Build binary
- `make pb-release` - Create GitHub release (multi-platform)
- `make pb-update` - Update from GitHub releases
- `make pb-gen-template` - Generate type-safe model template
- `make pb-gen-models` - Generate type-safe model code

## Architecture

**Data-First**: Collections defined in [collections.go](collections.go), auto-created on startup.

**Plugins Enabled** (see [base/main.go](base/main.go)):
- jsvm: JS hooks (`pb_hooks/*.pb.js`)
- migratecmd: Database migrations
- ghupdate: Self-update from GitHub

**Type-Safe Models**: Optional code generation with pocketbase-gogen.

## API Endpoints

**Dynamic Discovery**: The root endpoint (`/`) serves a dynamically-generated HTML page listing all available endpoints. This ensures the index stays in sync as routes are added/removed. For programmatic access, use `/api/` which returns the same information as JSON.

### OAuth Routes
- `GET /auth/google` - Initiate Google OAuth flow
- `GET /auth/google/callback` - OAuth callback
- `GET /auth/logout` - Logout
- `GET /auth/status` - Check auth status

### Calendar API Routes (Authenticated - requires OAuth)
- `GET /api/calendar/events` - List user's Google Calendar events
- `POST /api/calendar/events` - Create new event in Google Calendar

**Example - List Events**:
```bash
curl http://localhost:8090/api/calendar/events \
  -H "Cookie: pb_auth=YOUR_AUTH_TOKEN"
```

**Example - Create Event**:
```bash
curl -X POST http://localhost:8090/api/calendar/events \
  -H "Cookie: pb_auth=YOUR_AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Team Meeting",
    "start_time": "2024-03-20T10:00:00Z",
    "end_time": "2024-03-20T11:00:00Z",
    "location": "Office",
    "description": "Weekly sync"
  }'
```

## Files

```
pb/
├── base/main.go         # Follows .src/presentator pattern
├── collections.go       # Schema as code
├── oauth.go            # Google OAuth (stores tokens in DB)
├── calendar.go         # Calendar API (list/create events)
├── wellknown.go        # App wrapper
├── _templates/         # pocketbase-gogen templates
└── models/             # Generated type-safe models
```

**Architecture Separation**:

- **PocketBase** (`pb/`) - **Server-side Google Calendar access**
  - OAuth 2.0 flow (user consent → token storage)
  - Securely stores OAuth tokens in database
  - Provides authenticated API to read/write users' Google Calendar
  - Example: Your app can list a user's calendar events, create events, etc.

- **Main Server** (`pkg/server`) - **Client-side calendar deep links**
  - Generates Google/Apple Calendar URLs (no auth needed)
  - Stateless URL/ICS generation
  - Public-facing web UI for creating calendar links

**Why separate?**
- **Different use cases**:
  - PocketBase = Server acts on behalf of user (requires OAuth)
  - Main Server = User opens calendar link in their own app (no OAuth)
- **Security**: OAuth tokens stay in PocketBase database, never exposed to client
- **Flexibility**: Use PocketBase for server-side operations, main server for public links

See `make help` for all commands.
