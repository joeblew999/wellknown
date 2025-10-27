# GCP Setup Tool - Unified Web & CLI

**Purpose**: Interactive web dashboard + CLI automation for Google Cloud OAuth setup

## Quick Start

```bash
# Web mode (recommended - no gcloud needed!)
make gcp-setup

# Development mode with hot-reload (Air)
make gcp-setup-dev

# CLI mode (requires gcloud authentication)
make gcp-setup-cli
```

## What's Automated ✅

This tool automates:
1. ✅ **Project creation** - Creates GCP project programmatically
2. ✅ **API enablement** - Enables Calendar API, OAuth2 API, IAM API, Resource Manager API
3. ✅ **Clear instructions** - Provides step-by-step manual OAuth setup guide

## What Requires Manual Setup ⚠️

Due to Google API limitations, these steps require manual configuration:
1. ⚠️ **OAuth Consent Screen** - Must be configured via GCP Console (one-time)
2. ⚠️ **OAuth Client Credentials** - Must be created via GCP Console
3. ⚠️ **.env file** - Must be created manually with credentials

**Why?** Google's OAuth2 client creation API requires special IAM permissions and OAuth brand setup that's complex to automate. The GCP Console provides the best UX for this one-time setup.

---

## Quick Start

### 1. Prerequisites

- **Google Cloud account** with billing enabled
- **gcloud CLI** installed and authenticated:
  ```bash
  gcloud auth application-default login
  ```
- **Project ID** decided (e.g., `wellknown-calendar-dev`)

### 2. Run Automated Setup

```bash
export GCP_PROJECT_ID="wellknown-calendar-dev"
make gcp-setup
```

This will:
- Create the GCP project
- Enable required APIs
- Print manual OAuth setup instructions

### 3. Manual OAuth Setup (One-Time)

The tool will print URLs and instructions. Follow them to:

**Step 1: Configure OAuth Consent Screen**
```
https://console.cloud.google.com/apis/credentials/consent?project=YOUR_PROJECT
```
- User Type: **External**
- App name: **Wellknown Calendar**
- Support email: Your email
- Developer contact: Your email
- Scopes: Leave default
- Test users: Add your email
- Click through all steps

**Step 2: Create OAuth Client**
```
https://console.cloud.google.com/apis/credentials?project=YOUR_PROJECT
```
- Create Credentials → **OAuth client ID**
- Application type: **Web application**
- Name: **Wellknown PB Server**
- Authorized redirect URIs:
  - `http://localhost:8090/auth/google/callback`
  - `http://127.0.0.1:8090/auth/google/callback`
- Click **Create**
- **Copy Client ID and Client Secret**

**Step 3: Create .env File**
```bash
cd pb/base
cp .env.example .env
# Edit .env with your credentials:
# GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
# GOOGLE_CLIENT_SECRET=your-client-secret
```

### 4. Start Server

```bash
make pb-server
```

Or manually:
```bash
cd pb/base
source .env
go run main.go serve
```

Open: http://localhost:8090

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  GCP Setup Tool                         │
│                  (Go Program)                           │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Automated (via Go APIs):                              │
│  ✅ Create project                                      │
│  ✅ Enable APIs                                         │
│  ✅ Generate instructions                               │
│                                                         │
│  Manual (via GCP Console):                             │
│  ⚠️  Configure OAuth consent screen                    │
│  ⚠️  Create OAuth client credentials                   │
│  ⚠️  Copy credentials to .env                          │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## Future Improvements

### Option 1: Full REST API Automation

Implement direct REST API calls to create OAuth clients:
```go
// POST https://iap.googleapis.com/v1/projects/{project}/brands/{brand}/identityAwareProxyClients
func createOAuthClientViaREST(ctx context.Context, projectID string) error {
    // 1. Create OAuth brand (consent screen)
    // 2. Get brand ID
    // 3. Create OAuth client with brand ID
    // 4. Return credentials
}
```

**Challenges**:
- Requires service account with special IAM roles
- Brand creation API is complex
- Authentication flow is non-trivial

### Option 2: gcloud CLI Wrapper

Use `os/exec` to call gcloud CLI commands:
```bash
gcloud alpha iap oauth-clients create \
  --project=$PROJECT_ID \
  --brand=$BRAND_ID \
  --display_name="Wellknown PB Server"
```

**Challenges**:
- Requires user to have gcloud installed
- Requires finding/creating brand ID first
- Not as clean as pure Go solution

### Option 3: Terraform/Pulumi

Use infrastructure-as-code tools that have better OAuth support.

---

## Troubleshooting

### Error: "Project already exists"
**Solution**: That's fine! The tool will skip project creation and continue with API enablement.

### Error: "Permission denied"
**Solution**: Run `gcloud auth application-default login` to authenticate with GCP.

### Error: "Billing account required"
**Solution**: Enable billing on your GCP account first.

### OAuth consent screen says "Unverified app"
**Solution**: This is normal for development. Click "Advanced" → "Go to Wellknown Calendar (unsafe)" during testing. For production, submit app for verification.

---

## Security Notes

- Store `.env` files securely (never commit to git - already in `.gitignore`)
- OAuth credentials grant access to user's calendar - treat like passwords
- For production: Use GCP Secret Manager instead of .env files
- Redirect URIs must match exactly (including protocol and port)

---

## Commands

```bash
# Run setup
make gcp-setup

# Or directly
cd tools/gcp-setup
export GCP_PROJECT_ID="your-project-id"
go run main.go
```

---

**Last Updated**: 2025-10-27
**Status**: Partial automation (APIs automated, OAuth requires manual setup)
