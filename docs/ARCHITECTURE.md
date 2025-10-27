# Architecture - Two-Server Design

**Decision**: Separate demo server and production server

**Date**: 2025-10-27

---

## Overview

This project has **two separate servers** serving different purposes:

```
┌─────────────────────────────────────────────────────────────┐
│                    wellknown Project                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Demo Server              Production Server                 │
│  (cmd/server)             (cmd/pocketbase)                  │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │ No Auth      │         │ Google OAuth │                 │
│  │ Showcase     │         │ User Login   │                 │
│  │ Test URLs    │         │ Calendar API │                 │
│  └──────────────┘         └──────────────┘                 │
│         │                         │                         │
│         └─────────┬───────────────┘                         │
│                   │                                         │
│                   ▼                                         │
│          ┌────────────────┐                                │
│          │ Core Libraries  │                                │
│          │ pkg/google/     │                                │
│          │ pkg/server/     │                                │
│          └────────────────┘                                │
└─────────────────────────────────────────────────────────────┘
```

---

## Server #1: Demo Server (`cmd/server`)

### Purpose
**Testing and showcase** - Let anyone try deep link generation without auth

### Features
- ✅ Homepage with service cards
- ✅ Google Calendar URL generation
- ✅ Apple Calendar ICS generation
- ✅ Showcase pages with examples
- ✅ Custom form with validation
- ✅ QR code generation
- ❌ No authentication required
- ❌ No user accounts
- ❌ No Calendar API access

### Use Cases
- Testing URL generation
- Showcasing library capabilities
- Local development
- Quick demos

### Run
```bash
make dev        # Hot-reload dev mode
make run        # Production mode
```

### URL
`http://localhost:8080`

---

## Server #2: Production Server (`cmd/pocketbase`)

### Purpose
**Production app with authentication** - Let users access their real Google Calendar

### Features
- ✅ User signup/login (via Google OAuth)
- ✅ Session management (Pocketbase)
- ✅ Access user's Google Calendar (OAuth tokens)
- ✅ Create/read/update/delete events (Calendar API)
- ✅ User database (Pocketbase)
- ✅ Protected routes (auth middleware)
- ✅ Token refresh (OAuth token management)

### Use Cases
- Production deployment
- Real user accounts
- Calendar integration
- Event management

### Run
```bash
cd cmd/pocketbase
go run main.go serve
```

### URL
`http://localhost:8090` (Pocketbase default)

---

## Architecture Benefits

### Why Separate Servers?

1. **Simple Demo**
   - Demo server has zero auth complexity
   - Anyone can test URL generation immediately
   - No OAuth setup required for development
   - Fast iteration on UI/UX

2. **Secure Production**
   - Production server has proper authentication
   - User data is protected
   - OAuth tokens stored securely
   - Clear security boundaries

3. **Independent Development**
   - Work on demo without breaking auth
   - Work on auth without breaking demo
   - Different deployment strategies
   - Different scaling needs

4. **Clear Separation of Concerns**
   - Demo: "Try our library"
   - Production: "Use our app"
   - Different audiences, different needs

---

## Directory Structure

```
wellknown/
├── cmd/
│   ├── server/              # Demo server (no auth)
│   │   ├── main.go          # Entry point
│   │   └── .air.toml        # Hot-reload config
│   │
│   └── pocketbase/          # Production server (with auth)
│       ├── main.go          # Pocketbase + custom routes
│       ├── migrations/      # DB schema
│       ├── pb_hooks/        # Lifecycle hooks (optional)
│       └── .env.example     # Config template
│
├── pkg/
│   ├── google/
│   │   ├── calendar/        # URL generation (no auth needed)
│   │   └── console/         # Calendar API client (uses OAuth tokens)
│   │
│   ├── server/              # Demo server handlers (no auth)
│   │   ├── handlers.go
│   │   ├── routes.go
│   │   └── templates/
│   │
│   └── pb/                  # Pocketbase integration (NEW)
│       ├── auth.go          # Google OAuth setup
│       ├── middleware.go    # Auth middleware
│       ├── calendar.go      # User-scoped Calendar operations
│       └── tokens.go        # Token storage/refresh
│
├── tools/
│   └── gcp-setup/           # GCP project setup
│
└── docs/
    ├── ARCHITECTURE.md      # This file
    ├── pocketbase-setup.md  # Setup guide (TODO)
    └── api-endpoints.md     # API docs (TODO)
```

---

## Shared Components

Both servers use the same core libraries:

### `pkg/google/calendar/`
- URL generation (no auth)
- Used by both servers
- Pure functions, no side effects

### `pkg/google/console/`
- Calendar API client (OAuth tokens required)
- Used only by production server
- Requires authenticated user context

### `pkg/server/`
- Demo server handlers
- Templates and static assets
- Used only by demo server

---

## Development Workflow

### Working on Demo Server
```bash
cd /path/to/wellknown
make dev                    # Start with hot-reload
# Edit pkg/server/* files
# Browser auto-refreshes
```

### Working on Production Server
```bash
cd cmd/pocketbase
go run main.go serve        # Start Pocketbase
# Edit pb hooks or routes
# Restart server to see changes
```

### Working on Core Library
```bash
# Edit pkg/google/calendar/*
go test ./pkg/google/calendar  # Test changes
# Both servers pick up changes
```

---

## Deployment Strategies

### Demo Server
- Static binary deployment
- No database required
- Can run on serverless (Cloud Run, Lambda)
- Scales horizontally easily

### Production Server
- Pocketbase + SQLite (or PostgreSQL)
- Needs persistent storage for database
- Needs volume for pb_data/
- Can run on VPS, Docker, Fly.io

---

## Environment Variables

### Demo Server (cmd/server)
```bash
# No required env vars!
# Optional:
PORT=8080
```

### Production Server (cmd/pocketbase)
```bash
# Required:
GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=xxx
GOOGLE_REDIRECT_URL=http://localhost:8090/api/oauth2-redirect

# Pocketbase:
PB_ADMIN_EMAIL=admin@example.com
PB_ADMIN_PASSWORD=secure-password
```

---

## Future Considerations

### Could We Merge Them?
Yes, but we'd lose the simplicity of the demo server:
- Demo pages would need "public" mode
- Auth would add complexity to codebase
- Harder to showcase library capabilities
- More cognitive overhead for contributors

### Keep Separate Because:
- Demo server is a great library showcase
- Production server focuses on user features
- Clear mental model: "Try it" vs "Use it"
- Easier onboarding for contributors

---

## Related Documents

- [PREVENTION.md](PREVENTION.md) - Lessons learned
- [STATUS.md](../STATUS.md) - Implementation status
- [CLAUDE.md](../CLAUDE.md) - AI agent instructions

**Last Updated**: 2025-10-27
