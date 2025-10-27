# Google Console Integration

**Purpose**: Access user's actual Google Calendar for end-to-end testing and integration.

## Why This Package Exists

While `pkg/google/calendar` generates **URLs** that open Google Calendar web forms, this package provides **API access** to:
- Create events directly in user's calendar (no manual form submission)
- Read existing events from user's calendar
- Update/delete events
- Test our URL generation against real API behavior

## Architecture

```
pkg/google/
├── calendar/              # URL Generation (no auth required)
│   ├── event.go          # Event struct + GenerateURL()
│   └── testdata.go       # Test cases
└── console/              # API Integration (OAuth required)
    ├── client.go         # OAuth2 + API client setup
    ├── events.go         # CRUD operations on Calendar API
    └── compare.go        # Compare URL vs API (for validation)
```

## Setup

1. **Run GCP setup tool** (creates project + enables APIs):
   ```bash
   cd tools/gcp-setup
   go run main.go
   ```

2. **Create OAuth credentials** (manual step):
   - Go to: https://console.cloud.google.com/apis/credentials
   - Create OAuth 2.0 Client ID
   - Download credentials.json

3. **Set environment variables**:
   ```bash
   export GOOGLE_CLIENT_ID='your-client-id'
   export GOOGLE_CLIENT_SECRET='your-client-secret'
   ```

## Usage

### Create Event via API
```go
import "github.com/joeblew999/wellknown/pkg/google/console"

client := console.NewClient(ctx, clientID, clientSecret)
event := calendar.Event{
    Title: "Meeting",
    StartTime: time.Now(),
    EndTime: time.Now().Add(1 * time.Hour),
}
apiEventID, err := client.CreateEvent(event)
```

### Compare URL vs API
```go
// Generate URL (our library)
url := event.GenerateURL()

// Create via API (Google's official client)
apiEvent, _ := client.CreateEvent(event)

// Compare: Did our URL generate the same event as the API?
console.CompareEventData(event, apiEvent)
```

## Use Cases

1. **End-to-end testing**: Verify URL generation creates correct events
2. **Integration tests**: Test against real Google Calendar API
3. **User features**: Allow users to create events via API (optional future)
4. **Validation**: Ensure our URL format matches Google's API expectations

## Not in Scope (Yet)

- User authentication flow (OAuth dance)
- Token storage/refresh
- Production-ready API integration

This is a **testing and validation tool**, not a production API client.
