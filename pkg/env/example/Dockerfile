# Multi-stage Dockerfile for env-demo (Fly.io deployment demo)
# Build with: make docker-build
# Run with: make docker-run
#
# Commands available in container:
#   serve                Start HTTP server
#   health               Health check
#   sync-registry        Sync deployment configs
#   sync-environments    Merge secrets into environments
#   finalize             Encrypt files

# ================================================================
# Stage 1: Build
# ================================================================
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Set GOTOOLCHAIN=auto globally
ENV GOTOOLCHAIN=auto

# Copy parent module first (for replace directive)
COPY . .

# Navigate to example directory
WORKDIR /build/pkg/env/example

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
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

# ================================================================
# Environment Variables (injected at runtime by )
# AUTO-GENERATED - DO NOT EDIT MANUALLY
# ================================================================
#
# Required (set via ):
#   (none)
#
# Required (set via ):
#   DATABASE_URL
#   STRIPE_API_KEY
#
# Optional (set via  if needed):
#   # APIs
#   SENDGRID_API_KEY
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
