# Multi-stage Dockerfile for Go backend
# Build stage produces static binary, runtime uses distroless for security

# Build stage - compile Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files first for layer caching
# Docker caches layers, so dependencies only rebuild when go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build static binary with optimization flags
# CGO_ENABLED=0: Build pure Go binary (no C dependencies)
# GOOS=linux GOARCH=amd64: Target Linux x64
# -trimpath: Remove file system paths from binary
# -ldflags="-s -w": Strip debug info (-s) and symbol table (-w) for smaller size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o server \
    ./cmd/server

# Runtime stage - minimal distroless image
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /app/server .

# Copy necessary data files
COPY --from=builder /app/data ./data
COPY --from=builder /app/db ./db

# Security: Run as non-root user (distroless provides nonroot user)
USER nonroot:nonroot

# Trading engine API port
EXPOSE 8080

# Run the server
CMD ["./server"]
