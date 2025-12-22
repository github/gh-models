# BrainSait AI Platform - Production Dockerfile
# Multi-stage build for minimal image size

# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always)" \
    -o brainsait-models \
    main.go

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:3.19

# Security: run as non-root user
RUN addgroup -g 1000 brainsait && \
    adduser -u 1000 -G brainsait -s /bin/sh -D brainsait

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/brainsait-models .

# Copy domain-specific prompts (if any)
COPY --from=builder /app/examples ./prompts

# Set ownership
RUN chown -R brainsait:brainsait /app

# Switch to non-root user
USER brainsait

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./brainsait-models list || exit 1

# Default command
ENTRYPOINT ["./brainsait-models"]
CMD ["--help"]
