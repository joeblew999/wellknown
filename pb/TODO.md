# PocketBase TODO

**Purpose**: Server-side access to users' Google Calendar via OAuth 2.0

## What Works Now

✅ Google OAuth 2.0 flow (server-based, stores tokens in DB)
✅ Calendar API - list events (`GET /api/calendar/events`)
✅ Calendar API - create events (`POST /api/calendar/events`)
✅ Type-safe models with pocketbase-gogen
✅ Dynamic HTML/JSON endpoint discovery (`/` and `/api/`)
✅ Air hot-reload for development (`make pb-dev`)
✅ Multi-platform builds (darwin, linux, windows × arm64/amd64)
✅ GitHub releases with automatic updates (`make pb-release`, `make pb-update`)

## Near-Term (Next Sprint)

### Missing Calendar Operations
- [ ] **Update event** - `PUT /api/calendar/events/:id`
- [ ] **Delete event** - `DELETE /api/calendar/events/:id`
- [ ] **Get single event** - `GET /api/calendar/events/:id`

### OAuth Improvements
- [ ] Add `.env.example` file with required OAuth vars
- [ ] Better error messages when OAuth creds are missing
- [ ] Add OAuth token expiry display in admin UI

### Testing
- [ ] Create test user flow (OAuth → create event → list events)
- [ ] Test token refresh when token expires
- [ ] Test error handling (no token, invalid token, expired token)

## Later (When Needed)

### Calendar Features (if users ask for them)
- [ ] List all user's calendars (not just primary)
- [ ] Support for recurring events
- [ ] Attendees (add/remove/invite)
- [ ] Event reminders/notifications

### API Improvements (if needed)
- [ ] Pagination for event listing (currently returns all)
- [ ] Date range filtering (e.g., events this week)
- [ ] Search events by title/description

### Production (when deploying)
- [ ] Add proper logging (not just `log.Println`)
- [ ] Add metrics/health check endpoint
- [ ] Document deployment process
- [ ] Add backup/restore for pb_data

## Won't Do (Out of Scope)

❌ Client-side calendar links (that's in `pkg/server/`)
❌ ICS file generation (that's in `pkg/apple/calendar/`)
❌ Google Calendar URL generation (that's in `pkg/google/calendar/`)
❌ Multi-provider OAuth (Apple, Microsoft) - Google only for now
❌ Calendar sharing/permissions - use Google Calendar UI for that

## Notes

- **Keep it simple**: Only server-side Google Calendar access via OAuth
- **Separation**: Calendar URL/ICS generation stays in main server
- **Type-safe**: Always use generated models from `pb/models/`
- **OAuth**: Server acts on behalf of user after consent
- **No bloat**: Don't add features until they're actually needed


## Other

Highly available leaderless PocketBase cluster powered by go-ha database/sql driver

https://github.com/litesql/pocketbase-ha

---

Automatic JWT generation & management for NATS users, triggered from PocketBase collections (like “nats_system_operator”, “nats_accounts”, “nats_users”) inside a PocketBase app

https://github.com/skeeeon/pb-nats

https://github.com/skeeeon/pb-cli

https://github.com/skeeeon/rule-router

---

NATS emebdded

---

https://github.com/infogulch/xtemplate

---

**Last Updated**: 2025-11-03
