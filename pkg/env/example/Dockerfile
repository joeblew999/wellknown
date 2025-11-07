# Multi-stage Dockerfile for env-demo example
# Build with: make docker-build
# Run with: make docker-run

# ================================================================
# Stage 1: Build
# ================================================================
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /app/server \
    .

# ================================================================
# Stage 2: Runtime
# ================================================================
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# === START AUTO-GENERATED ENV DOCS ===
# ================================================================
# Environment Variables (injected at runtime by Docker)
# AUTO-GENERATED - DO NOT EDIT MANUALLY
# To update: go run . dockerfile-sync
# ================================================================
#
# Required (set via docker-compose.yml or Dockerfile ENV):
#   (none)
#
# Required (set via Docker secrets or environment):
#   DATABASE_URL
#     Database connection URL
#   STRIPE_API_KEY
#     Stripe API secret key
#   STRIPE_WEBHOOK_SECRET
#     Stripe webhook signing secret
#   OAUTH_GOOGLE_CLIENT_ID
#     Google OAuth client ID
#   OAUTH_GOOGLE_CLIENT_SECRET
#     Google OAuth client secret
#   OAUTH_GOOGLE_REDIRECT_URL
#     Google OAuth redirect URL
#
# Optional (set via Docker secrets or environment if needed):
#   # Cache
#   REDIS_URL
#     Redis connection URL
#   REDIS_PASSWORD
#     Redis password
#   # External APIs
#   SENDGRID_API_KEY
#     SendGrid API key for email
#   # OAuth
#   OAUTH_GITHUB_CLIENT_ID
#     GitHub OAuth client ID
#   OAUTH_GITHUB_CLIENT_SECRET
#     GitHub OAuth client secret
#
# Sync secrets: docker-compose up
# ================================================================
# === END AUTO-GENERATED ENV DOCS ===
#   SERVER_TIMEOUT (default: 30)
#     Server request timeout in seconds
#   LOG_LEVEL (default: info)
#     Logging level (debug, info, warn, error)
#   DATABASE_MAX_CONNECTIONS (default: 25)
#     Maximum database connections
#   DATABASE_SSL_MODE (default: require)
#     Database SSL mode (disable, require, verify-full)
#   CACHE_TTL (default: 3600)
#     Cache TTL in seconds
#   FEATURE_NEW_UI (default: false)
#     Enable new UI (true/false)
#   FEATURE_BETA_API (default: false)
#     Enable beta API endpoints (true/false)
#   FEATURE_ANALYTICS (default: true)
#     Enable analytics tracking (true/false)
#
# Optional (no defaults):
#   SENDGRID_API_KEY
#     SendGrid API key for email
#   OAUTH_GITHUB_CLIENT_ID
#     GitHub OAuth client ID
#   OAUTH_GITHUB_CLIENT_SECRET
#     GitHub OAuth client secret
#   REDIS_URL
#     Redis connection URL
#   REDIS_PASSWORD
#     Redis password
#
# ================================================================

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "health"] || exit 1

# Run the application
ENTRYPOINT ["/app/server"]
CMD ["serve"]
