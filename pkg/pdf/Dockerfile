# Multi-stage build for minimal image size
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY ../../go.mod ../../go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
WORKDIR /build/pkg/pdf/examples/pdfform
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pdfform .

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS downloads
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/pkg/pdf/examples/pdfform/pdfform /app/pdfform

# Create .data directory structure and set permissions
RUN mkdir -p /app/.data/catalog \
             /app/.data/downloads \
             /app/.data/templates \
             /app/.data/outputs \
             /app/.data/cases/test_scenarios \
             /app/.data/temp && \
    chmod -R 755 /app/.data

# Copy catalog file if it exists
COPY .data/catalog/australian_transfer_forms.csv /app/.data/catalog/ 2>/dev/null || true

# Environment variables
ENV PDFFORM_DATA_DIR=/app/.data

# Expose port for web server
EXPOSE 8080

# Volume mount point for persistent data
VOLUME ["/app/.data"]

# Run as non-root user
RUN adduser -D -u 1000 pdfform && \
    chown -R pdfform:pdfform /app
USER pdfform

# Default command - run web server
CMD ["/app/pdfform", "serve", "--port", "8080"]
