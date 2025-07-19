# Development Setup Guide

This guide explains how to set up a local development environment for the Near-RT RIC platform using KIND (Kubernetes in Docker), build the Go backend, and serve the Angular dashboards.

## Prerequisites

Ensure you have the following tools installed:

- **Docker**: For running KIND clusters
- **kubectl**: Kubernetes command-line tool  
- **KIND**: Kubernetes in Docker (v0.11.0 or later)
- **Go**: Version 1.17 or later
- **Node.js**: Version 16.14.2 or later
- **npm**: Comes with Node.js
- **Angular CLI**: Version 13.3.3 (`npm install -g @angular/cli@13.3.3`)

## 1. Setting Up KIND Cluster

### Create KIND Cluster

```bash
# Create a KIND cluster with a specific name
kind create cluster --name near-rt-ric

# Verify the cluster is running
kubectl cluster-info --context kind-near-rt-ric

# Set the context
kubectl config use-context kind-near-rt-ric
```

### Configure Local Registry (Optional)

For faster image pulls during development:

```bash
# Create local registry
docker run -d --restart=always -p 5000:5000 --name registry registry:2

# Configure KIND to use local registry
cat <<EOF | kind create cluster --name near-rt-ric --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
    endpoint = ["http://registry:5000"]
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

# Connect registry to KIND network
docker network connect kind registry
```

## 2. Backend Setup (dashboard-master/dashboard-master)

Navigate to the main dashboard directory:

```bash
cd dashboard-master/dashboard-master
```

### Build and Start Backend

```bash
# Build the Go backend
make build

# Start development server with backend
npm start

# Or start with HTTPS
npm run start:https
```

### Backend Development Workflow

For active Go development with live reload:

```bash
# Watch backend for changes (if available)
make watch-backend

# Run backend tests
make test

# Code quality checks
make check
npm run fix
```

### Environment Configuration

The backend requires proper Kubernetes RBAC configuration. Create a service account:

```bash
# Apply RBAC configuration
kubectl apply -f aio/deploy/recommended/05_dashboard-rbac.yaml

# Or create minimal RBAC for development
kubectl create serviceaccount dashboard-admin -n default
kubectl create clusterrolebinding dashboard-admin \
  --clusterrole=cluster-admin \
  --serviceaccount=default:dashboard-admin
```

## 3. Frontend Setup

### Main Dashboard Frontend

The main dashboard frontend is served alongside the backend:

```bash
cd dashboard-master/dashboard-master

# Install dependencies (if not already done)
npm install

# Start development server (includes both frontend and backend)
npm start

# The application will be available at:
# HTTP: http://localhost:8080
# HTTPS: https://localhost:8443
```

### xApp Dashboard Frontend

Set up the separate xApp dashboard:

```bash
cd xAPP_dashboard-master

# Install dependencies
npm install

# Start development server
npm start

# The xApp dashboard will be available at:
# http://localhost:4200
```

## 4. Testing

### Backend Tests

```bash
cd dashboard-master/dashboard-master
make test
```

### Frontend Tests

```bash
# Main dashboard tests
cd dashboard-master/dashboard-master
npm test

# xApp dashboard tests
cd xAPP_dashboard-master
npm test
```

### End-to-End Tests

```bash
# Run Cypress tests (main dashboard)
cd dashboard-master/dashboard-master
npm run e2e

# Run Cypress tests (xApp dashboard)
cd xAPP_dashboard-master
npm run e2e
```

## 5. Deployment to KIND

Deploy the complete platform to your KIND cluster:

```bash
cd dashboard-master/dashboard-master

# Build and deploy to KIND
make deploy

# Check deployment status
kubectl get pods -n kubernetes-dashboard
kubectl get services -n kubernetes-dashboard
```

## Troubleshooting

### RBAC Errors

If you encounter permission errors like "forbidden: User cannot list resources":

1. **Check Service Account**:
   ```bash
   kubectl get serviceaccount dashboard-admin -n default
   ```

2. **Verify ClusterRoleBinding**:
   ```bash
   kubectl get clusterrolebinding dashboard-admin
   kubectl describe clusterrolebinding dashboard-admin
   ```

3. **Create Admin Access** (for development only):
   ```bash
   kubectl create clusterrolebinding dashboard-admin \
     --clusterrole=cluster-admin \
     --serviceaccount=default:dashboard-admin
   ```

4. **Get Service Account Token**:
   ```bash
   # For Kubernetes 1.24+
   kubectl create token dashboard-admin
   
   # For older versions
   kubectl get secret $(kubectl get serviceaccount dashboard-admin -o jsonpath='{.secrets[0].name}') -o jsonpath='{.data.token}' | base64 --decode
   ```

### Image Pull Issues

If you encounter image pull errors in KIND:

1. **Load Images into KIND**:
   ```bash
   # Build local image
   docker build -t dashboard:local .
   
   # Load into KIND
   kind load docker-image dashboard:local --name near-rt-ric
   ```

2. **Configure ImagePullPolicy**:
   Edit deployment manifests to use `imagePullPolicy: Never` for local images.

### Build Failures

1. **Node.js Version Issues**:
   ```bash
   # Check Node.js version
   node --version  # Should be >= 16.14.2
   
   # Use nvm to manage versions
   nvm install 16.14.2
   nvm use 16.14.2
   ```

2. **Go Version Issues**:
   ```bash
   # Check Go version
   go version  # Should be >= 1.17
   ```

3. **Clear Node Modules**:
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   ```

### Port Conflicts

If ports 8080 or 4200 are in use:

1. **Main Dashboard**:
   ```bash
   npm start -- --port 8081
   ```

2. **xApp Dashboard**:
   ```bash
   ng serve --port 4201
   ```

### KIND Cluster Issues

1. **Reset KIND Cluster**:
   ```bash
   kind delete cluster --name near-rt-ric
   kind create cluster --name near-rt-ric
   ```

2. **Check Docker Resources**:
   Ensure Docker has sufficient memory (at least 4GB) and CPU resources.

### Network Issues

If you cannot access the dashboards:

1. **Check Port Forwarding**:
   ```bash
   # Forward dashboard service
   kubectl port-forward -n kubernetes-dashboard service/kubernetes-dashboard 8443:443
   ```

2. **Verify Service Status**:
   ```bash
   kubectl get services -n kubernetes-dashboard
   kubectl describe service kubernetes-dashboard -n kubernetes-dashboard
   ```

## Quick Start Commands

Once everything is set up, use these commands for daily development:

```bash
# Start KIND cluster
kind create cluster --name near-rt-ric

# Start main dashboard (backend + frontend)
cd dashboard-master/dashboard-master && npm start

# Start xApp dashboard (in separate terminal)
cd xAPP_dashboard-master && npm start

# Run tests
cd dashboard-master/dashboard-master && make test
cd xAPP_dashboard-master && npm test
```

## Useful Links

- [KIND Documentation](https://kind.sigs.k8s.io/)
- [Kubernetes Dashboard Documentation](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/)
- [Angular CLI Documentation](https://angular.io/cli)
- [O-RAN Alliance Specifications](https://www.o-ran.org/specifications)
