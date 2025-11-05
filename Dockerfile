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
