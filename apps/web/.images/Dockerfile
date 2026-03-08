# Web Server Image (templ SSR)
# Go application serving HTML on port 3000
#
# Build context should be repository root:
#   docker build -f apps/web/.images/Dockerfile -t web .

# =============================================================================
# Builder Stage
# =============================================================================
FROM golang:1.25-alpine AS builder

# Install certificates and git for dependencies
RUN apk add --no-cache ca-certificates git

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy source code
WORKDIR /build
COPY apps/go.work apps/go.work.sum ./
COPY apps/web ./web
COPY apps/pkgs ./pkgs

# Generate templ files
WORKDIR /build/web/src
RUN templ generate

# Build binary
WORKDIR /build/web/src
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -ldflags="-s -w" -trimpath -o /web ./cmd

# =============================================================================
# Runner Stage
# =============================================================================
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 web && \
    adduser -u 1000 -G web -s /bin/sh -D web

# Copy built binary
COPY --from=builder /web /usr/local/bin/web

# Use non-root user
USER web
WORKDIR /home/web

# Expose web port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Run application
CMD ["/usr/local/bin/web"]
