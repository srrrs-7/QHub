# API Server Image
# Go application serving REST API on port 8080
#
# Build context should be repository root:
#   docker build -f apps/api/.images/Dockerfile -t api .

# =============================================================================
# Builder Stage
# =============================================================================
FROM golang:1.25-alpine AS builder

# Install certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy source code
WORKDIR /build
COPY apps/go.work apps/go.work.sum ./
COPY apps/api ./api
COPY apps/pkgs ./pkgs

# Build binary
WORKDIR /build/api/src
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -ldflags="-s -w" -trimpath -o /api ./cmd

# =============================================================================
# Runner Stage
# =============================================================================
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 api && \
    adduser -u 1000 -G api -s /bin/sh -D api

# Copy built binary
COPY --from=builder /api /usr/local/bin/api

# Use non-root user
USER api
WORKDIR /home/api

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run application
CMD ["/usr/local/bin/api"]
