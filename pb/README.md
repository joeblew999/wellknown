# Wellknown Pocketbase Integration

Server-based Google OAuth + Calendar API integration for Pocketbase.

## Architecture

```
pb/
├── wellknown.go          # Main entry point (wraps Pocketbase)
├── collections.go        # Collection setup (google_tokens)
├── oauth.go              # Server-based Google OAuth flow
├── calendar.go           # Google Calendar API routes
└── base/                 # Demo server
    ├── main.go           # Standalone server
    ├── pb_public/        # Static files served by PB
    │   └── index.html    # OAuth login UI (no JS SDK)
    └── .env.example      # Environment template
```

## Features

✅ **Server-Based OAuth** - No client-side JS SDK, all OAuth flow handled server-side
✅ **Google Calendar API** - List and create events via authenticated API
✅ **Token Storage** - OAuth tokens stored in Pocketbase collections
✅ **Automatic Token Refresh** - Expired tokens refreshed automatically
✅ **Importable Library** - Can be used by other Pocketbase projects

## Setup

### 1. Google Cloud Console

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create or select a project
3. Enable APIs:
   - Google Calendar API
   - Google+ API (for user info)
4. Create OAuth 2.0 credentials:
   - Application type: Web application
   - Authorized redirect URIs: `http://localhost:8090/auth/google/callback`
5. Copy Client ID and Client Secret

### 2. Environment Variables

```bash
cd pb/base
cp .env.example .env
# Edit .env with your credentials
```

Required:
- `GOOGLE_CLIENT_ID` - From Google Cloud Console
- `GOOGLE_CLIENT_SECRET` - From Google Cloud Console
- `GOOGLE_REDIRECT_URL` - `http://localhost:8090/auth/google/callback`

### 3. Run Server

```bash
# Install dependencies
cd pb
go mod tidy

# Run server
cd base
source .env  # Load environment variables
go run main.go serve
```

Access:
- **Homepage**: http://localhost:8090
- **Admin UI**: http://localhost:8090/_/

## How It Works

### OAuth Flow (Server-Based)

1. **User clicks "Sign in with Google"** → `/auth/google`
2. **Server redirects to Google** with OAuth request
3. **User authorizes** in Google's consent screen
4. **Google redirects back** → `/auth/google/callback?code=...`
5. **Server exchanges code for token** (server-to-server)
6. **Server stores token** in `google_tokens` collection
7. **Server creates PB session** with auth cookie
8. **User is logged in** ✅

### API Routes

#### Public Routes
- `GET /` - Homepage with login UI
- `GET /auth/google` - Initiate OAuth flow
- `GET /auth/google/callback` - OAuth callback
- `GET /auth/status` - Check if authenticated

#### Protected Routes (require auth)
- `GET /api/calendar/events` - List user's calendar events
- `POST /api/calendar/events` - Create new event
- `GET /auth/logout` - Sign out

### Collections

#### `google_tokens`
Stores OAuth tokens for each user:

```
- user_id: text (PB user ID)
- access_token: text
- refresh_token: text
- token_type: text
- expiry: date
- email: email
```

Auto-created on first server start.

## Usage in Other Projects

You can import this as a library:

```go
package main

import (
    "github.com/joeblew999/wellknown/pb"
    "github.com/pocketbase/pocketbase"
)

func main() {
    // Option 1: Standalone
    app := wellknown.New()
    app.Start()

    // Option 2: Add to existing PB app
    pbApp := pocketbase.New()
    wk := wellknown.NewWithApp(pbApp)
    pbApp.Start()
}
```

## Development Tools

### pocketbase-gogen

Generate type-safe Go code from PB collections:

```bash
# Install
go install github.com/snonky/pocketbase-gogen@latest

# Generate template
pocketbase-gogen template ./pb_data ./pb/generated/template.go

# Generate proxies
pocketbase-gogen generate ./pb/generated/template.go ./pb/generated/proxies.go --utils --hooks
```

See: [.src/pocketbase-gogen](.src/pocketbase-gogen) for reference.

## Security Notes

- OAuth state tokens prevent CSRF attacks
- Tokens stored server-side (not exposed to client)
- HTTPS required in production
- Use secure cookies (`Secure`, `HttpOnly`, `SameSite`)

## Testing

1. Start server: `go run main.go serve`
2. Open: http://localhost:8090
3. Click "Sign in with Google"
4. Authorize access
5. View your calendar events
6. Check PB admin: http://localhost:8090/_/

## Related Documentation

- [POCKETBASE-ARCHITECTURE.md](../docs/POCKETBASE-ARCHITECTURE.md) - Architecture decisions
- [JSONSCHEMA-POCKETBASE.md](../docs/JSONSCHEMA-POCKETBASE.md) - JSON Schema integration (future)

---

**Last Updated**: 2025-10-27
