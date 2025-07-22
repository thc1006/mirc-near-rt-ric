# Multi-stage Dockerfile for O-RAN Near-RT RIC
# Production-ready image with security best practices

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    && rm -rf /var/cache/apk/*

# Create non-root user for building
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o near-rt-ric \
    ./cmd/near-rt-ric

# Verify the binary
RUN ldd near-rt-ric || true  # Should show "not a dynamic executable"

# Final stage - distroless for maximum security
FROM gcr.io/distroless/static-debian11:nonroot

# Labels for metadata
LABEL maintainer="O-RAN Near-RT RIC Team" \
      version="1.0.0" \
      description="Production-grade O-RAN Near Real-Time RAN Intelligent Controller" \
      org.opencontainers.image.title="near-rt-ric" \
      org.opencontainers.image.description="O-RAN Near-RT RIC with E2, A1, and O1 interfaces" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.vendor="O-RAN Software Community" \
      org.opencontainers.image.licenses="Apache-2.0"

# Create necessary directories with proper permissions
USER root
RUN mkdir -p /etc/near-rt-ric /var/log/near-rt-ric /var/lib/near-rt-ric
USER nonroot:nonroot

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/near-rt-ric .

# Copy configuration files
COPY --chown=nonroot:nonroot configs/default-config.yaml /etc/near-rt-ric/config.yaml
COPY --chown=nonroot:nonroot configs/yang/ /etc/near-rt-ric/yang/

# Expose ports according to O-RAN specifications
# E2 interface (SCTP)
EXPOSE 36421
# A1 interface (HTTP/HTTPS)
EXPOSE 10020 10021
# O1 interface (NETCONF over SSH)
EXPOSE 830
# Metrics (Prometheus)
EXPOSE 9090
# Health check
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/app/near-rt-ric", "--health-check"]

# Set entrypoint and default command
ENTRYPOINT ["/app/near-rt-ric"]
CMD ["--config=/etc/near-rt-ric/config.yaml"]

# Security: Run as non-root user (already set via distroless)
# Security: Read-only filesystem (can be set at runtime with --read-only)
# Security: No shell access (distroless doesn't include shell)
# Security: Minimal attack surface (distroless contains only necessary files)