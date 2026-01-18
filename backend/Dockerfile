# Multi-stage Dockerfile for RTX Trading Engine Backend

# Stage 1: Build stage
FROM golang:1.24.0-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME=unknown

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -extldflags '-static'" \
    -o rtx-backend \
    ./cmd/server/main.go

# Verify the binary
RUN chmod +x rtx-backend && ./rtx-backend --version || true

# Stage 2: Runtime stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1000 rtx && \
    adduser -D -u 1000 -G rtx rtx

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=rtx:rtx /build/rtx-backend /app/rtx-backend

# Copy configuration files (if any)
COPY --chown=rtx:rtx config/ /app/config/ || true
COPY --chown=rtx:rtx migrations/ /app/migrations/ || true

# Create data directories
RUN mkdir -p /app/data /app/logs && \
    chown -R rtx:rtx /app

# Switch to non-root user
USER rtx

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release \
    LOG_LEVEL=info \
    PORT=8080 \
    METRICS_PORT=9090

# Run the binary
ENTRYPOINT ["/app/rtx-backend"]
CMD ["server"]
