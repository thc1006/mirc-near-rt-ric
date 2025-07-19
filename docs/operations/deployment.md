# Production Deployment Guide

This guide covers deploying the O-RAN Near-RT RIC platform in production environments.

## Prerequisites

### System Requirements

**Minimum Requirements:**
- Kubernetes 1.21+
- Docker 20.10+
- kubectl 1.21+
- Go 1.17+ (for building from source)
- Node.js 16.14.2+ (for frontend builds)

**Recommended Resources:**
- CPU: 4+ cores per node
- Memory: 8GB+ per node
- Storage: 100GB+ persistent storage
- Network: 1Gbps+ bandwidth

### Kubernetes Cluster Setup

TODO: Add specific cluster configuration based on actual deployment manifests found in codebase.

## Deployment Methods

### Method 1: Production Build and Deploy

Based on the Makefile in `dashboard-master/dashboard-master/Makefile`:

```bash
# Navigate to main dashboard directory
cd dashboard-master/dashboard-master

# Build production binaries for all architectures
make build-cross

# Create production Docker images
make docker-build-release

# Run in production mode
make prod
```

### Method 2: Development Deployment

```bash
# Build for local development
make build

# Serve backend with hot reload
make serve-backend

# Or watch for changes with air (live reload)
make watch-backend
```

### Method 3: Quick Start (Development)

```bash
# Build both backend and frontend
make build

# Start development server
make prod
```

### Method 4: Docker Container Deployment

```bash
# Build Docker images for specific architecture
docker buildx build -t dashboard-amd64:v2.5.1 --platform linux/amd64 .

# Run container
docker run -p 8080:8080 dashboard-amd64:v2.5.1
```

## Configuration

### Environment Variables

Based on the Makefile configuration in `dashboard-master/dashboard-master/Makefile`:

```bash
# Core Configuration
export KUBECONFIG="${HOME}/.kube/config"    # Kubernetes config path
export TOKEN_TTL="900"                      # Token expiration (15 minutes)
export BIND_ADDRESS="127.0.0.1"           # Service bind address
export PORT="8080"                         # Service port

# Security Configuration
export AUTO_GENERATE_CERTS="false"        # Auto-generate TLS certificates
export ENABLE_INSECURE_LOGIN="false"      # Allow insecure HTTP login
export ENABLE_SKIP_LOGIN="false"          # Allow login bypass

# System Configuration
export SYSTEM_BANNER="Local test environment"  # System banner text
export SYSTEM_BANNER_SEVERITY="INFO"          # Banner severity level
export SIDECAR_HOST="http://localhost:8000"   # Sidecar service host

# Build Configuration
export GOOS="linux"                        # Target OS for builds
export GOARCH="amd64"                      # Target architecture
```

### Application Arguments

The dashboard binary accepts the following command-line arguments:

```bash
./dashboard \
  --kubeconfig=/path/to/kubeconfig \
  --sidecar-host=http://localhost:8000 \
  --system-banner="Production Environment" \
  --system-banner-severity=WARNING \
  --token-ttl=900 \
  --auto-generate-certificates=true \
  --enable-insecure-login=false \
  --enable-skip-login=false \
  --locale-config=./locale_conf.json \
  --bind-address=0.0.0.0 \
  --port=8080
```

### ConfigMaps and Secrets

Required Kubernetes resources for authentication and certificates:

```yaml
# Secret for encryption keys
apiVersion: v1
kind: Secret
metadata:
  name: kubernetes-dashboard-key-holder
  namespace: kubernetes-dashboard
type: Opaque

---
# Secret for custom certificates
apiVersion: v1
kind: Secret
metadata:
  name: kubernetes-dashboard-certs
  namespace: kubernetes-dashboard
type: kubernetes.io/tls
```

### Resource Limits

Recommended resource configuration:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 200Mi
  limits:
    cpu: 500m
    memory: 500Mi
```

## High Availability Setup

### Multi-node Deployment

TODO: Add HA configuration based on actual Kubernetes deployment patterns found in codebase.

### Load Balancing

TODO: Document load balancing setup.

### Backup and Recovery

TODO: Add backup procedures.

## Security Configuration

### RBAC Setup

Based on existing access control documentation in `dashboard-master/dashboard-master/docs/user/access-control/`:

TODO: Reference actual RBAC configurations from the codebase.

### TLS Configuration

TODO: Document TLS setup based on actual implementation.

### Network Policies

TODO: Add network policy configurations.

## Post-Deployment Verification

### Health Checks

TODO: Add health check procedures based on actual endpoints.

### Functional Testing

TODO: Add testing procedures.

### Performance Validation

TODO: Add performance validation steps.

---

**Status**: This is a documentation stub. Content will be populated based on actual deployment configurations found in the repository.