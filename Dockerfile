# Multi-stage Dockerfile for O-RAN Near-RT RIC

# Build arguments
ARG BACKEND_BINARY_PATH=./artifacts/backend/dashboard
ARG FRONTEND_DIST_PATH=./artifacts/frontend

# Base image for runtime
FROM gcr.io/distroless/static-debian11:nonroot AS runtime

# Metadata
LABEL maintainer="O-RAN Near-RT RIC Team"
LABEL version="1.0.0"
LABEL description="O-RAN Near Real-Time RAN Intelligent Controller"

# Create application directory
WORKDIR /app

# Copy backend binary from artifacts
COPY ${BACKEND_BINARY_PATH} /app/dashboard

# Copy frontend dist from artifacts
COPY ${FRONTEND_DIST_PATH} /app/static/

# Copy configuration files
COPY dashboard-master/dashboard-master/aio/deploy/recommended/ /app/config/

# Set up non-root user (distroless already has nonroot user)
USER nonroot:nonroot

# Expose ports
EXPOSE 8080 8443 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/app/dashboard", "--health-check"]

# Set entrypoint
ENTRYPOINT ["/app/dashboard"]
CMD ["--config=/app/config/config.yaml"]