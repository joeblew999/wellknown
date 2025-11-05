# Multi-stage build for wellknown PocketBase app
# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO disabled (pure Go SQLite via modernc.org/sqlite)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o wellknown-pb .

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for OAuth/API calls and tzdata for timezone support
RUN apk --no-cache add ca-certificates tzdata wget

# ================================================================
# Environment Variables (injected at runtime by Fly.io)
# ================================================================
#
# Required (set via fly.toml [env] section):
#   PB_DATA_DIR=/app/.data/pb     - PocketBase data directory
#   SERVER_HOST=0.0.0.0            - Server bind address
#   SERVER_PORT=8090               - Server port
#
# Required (set via Fly.io secrets):
#   GOOGLE_CLIENT_ID               - Google OAuth client ID
#   GOOGLE_CLIENT_SECRET           - Google OAuth client secret
#   GOOGLE_REDIRECT_URL            - Google OAuth callback URL
#
# Optional (set via Fly.io secrets if needed):
#   PB_ADMIN_EMAIL                 - Admin email for PocketBase
#   PB_ADMIN_PASSWORD              - Admin password for PocketBase
#   APPLE_TEAM_ID                  - Apple Developer Team ID
#   APPLE_CLIENT_ID                - Apple OAuth client ID
#   APPLE_KEY_ID                   - Apple private key ID
#   APPLE_PRIVATE_KEY              - Apple private key (inline PEM)
#   APPLE_REDIRECT_URL             - Apple OAuth callback URL
#   SMTP_HOST                      - SMTP server hostname
#   SMTP_PORT                      - SMTP server port
#   SMTP_USERNAME                  - SMTP username
#   SMTP_PASSWORD                  - SMTP password
#   SMTP_FROM_EMAIL                - From email address
#   SMTP_FROM_NAME                 - From name
#   S3_ENDPOINT                    - S3 endpoint URL
#   S3_REGION                      - S3 region
#   S3_BUCKET                      - S3 bucket name
#   S3_ACCESS_KEY                  - S3 access key
#   S3_SECRET_KEY                  - S3 secret key
#   S3_FORCE_PATH_STYLE            - S3 path style (true/false)
#   ANTHROPIC_API_KEY              - Anthropic Claude API key
#   ANTHROPIC_MODEL                - Claude model name (optional, has default)
#
# Sync secrets: make fly-secrets
# ================================================================

# Copy binary from builder
COPY --from=builder /build/wellknown-pb .

# Copy PocketBase hooks and migrations
COPY --from=builder /build/pkg/cmd/pocketbase/pb_hooks ./pb_hooks
COPY --from=builder /build/pkg/cmd/pocketbase/pb_migrations ./pb_migrations

# Create .data directory structure for multi-service architecture
# - .data/pb/      PocketBase SQLite databases (mounted from volume)
# - .data/nats/    Future: NATS JetStream state for HA
# - .data/uploads/ Future: File uploads
RUN mkdir -p /app/.data/pb /app/.data/nats

# Expose PocketBase port
EXPOSE 8090

# Health check - PocketBase provides /api/health endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/api/health || exit 1

# Run PocketBase service with custom data directory
# The "pb" command starts the PocketBase server
# --dir specifies the data directory (overridden by Go code default)
# Use serve command with --http flag to bind to 0.0.0.0
CMD ["./wellknown-pb", "pb", "serve", "--http=0.0.0.0:8090"]
