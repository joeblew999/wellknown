# Google Cloud Platform Setup

Quick guide to get Google OAuth credentials for PocketBase.

## Google vs Apple Calendar Integration

**Google Calendar (this guide)**:
- Requires OAuth 2.0 setup (this guide)
- Server accesses user's calendar on their behalf
- Users must authenticate via Google
- Can create/read/update/delete events in user's calendar

**Apple Calendar** (no setup needed):
- No OAuth required - generates .ics files
- Works completely client-side
- User downloads .ics file and opens in Apple Calendar
- Already implemented in `pkg/apple/calendar/`

## Prerequisites
- Google account
- 5 minutes

## Steps

### 1. Create Google Cloud Project
1. Go to https://console.cloud.google.com/
2. Click **Select a project** → **NEW PROJECT**
3. Name: `Wellknown Calendar` (or anything)
4. Click **CREATE**

### 2. Enable Google Calendar API
1. Go to https://console.cloud.google.com/apis/library
2. Search: **Google Calendar API**
3. Click **ENABLE**

### 3. Configure OAuth Consent Screen
1. Go to https://console.cloud.google.com/apis/credentials/consent
2. Choose **External** → **CREATE**
3. Fill in:
   - App name: `Wellknown Calendar`
   - User support email: (your email)
   - Developer contact: (your email)
4. Click **SAVE AND CONTINUE**
5. Scopes: Click **ADD OR REMOVE SCOPES**
   - Search: `calendar.events`
   - Select: `.../auth/calendar.events` (Read/write access)
   - Click **UPDATE** → **SAVE AND CONTINUE**
6. Test users: Click **ADD USERS**
   - Enter your Google email
   - Click **ADD** → **SAVE AND CONTINUE**
7. Click **BACK TO DASHBOARD**

### 4. Create OAuth Credentials
1. Go to https://console.cloud.google.com/apis/credentials
2. Click **CREATE CREDENTIALS** → **OAuth client ID**
3. Application type: **Web application**
4. Name: `Wellknown PocketBase`
5. Authorized redirect URIs:
   - Click **ADD URI**
   - Enter: `http://localhost:8090/auth/google/callback`
6. Click **CREATE**
7. **IMPORTANT**: Copy the Client ID and Client Secret

### 5. Configure PocketBase
```bash
cd pb/base
cp .env.example .env
```

Edit `.env` and paste your credentials:
```bash
GOOGLE_CLIENT_ID=123456789-abc.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-abc123def456
GOOGLE_REDIRECT_URL=http://localhost:8090/auth/google/callback
```

### 6. Test OAuth Flow
```bash
make pb-server
```

Then visit: http://localhost:8090/
- Click **Sign in with Google**
- Choose your test user account
- Grant calendar permissions
- You should see "Welcome back!" with calendar access

## Troubleshooting

### Error: "redirect_uri_mismatch"
- Check Authorized redirect URIs in GCP Console
- Must be **exactly**: `http://localhost:8090/auth/google/callback`

### Error: "Access blocked: This app's request is invalid"
- Make sure OAuth consent screen is configured
- Add your email to Test users

### Error: "Missing environment variables"
- Check `.env` file exists in `pb/base/`
- Run server from repo root: `make pb-server`

### Error: "Calendar API has not been used"
- Enable Google Calendar API in GCP Console
- Wait 1-2 minutes for it to propagate

## Production Deployment

For production (non-localhost):
1. Update Authorized redirect URIs to use your domain
2. Change OAuth consent screen from "Testing" to "Production"
3. Update `GOOGLE_REDIRECT_URL` in `.env`
