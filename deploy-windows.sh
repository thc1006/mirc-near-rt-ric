#\!/bin/bash
# O-RAN Near-RT RIC Windows Deployment Script
set -euo pipefail

CLUSTER_NAME="near-rt-ric"
NAMESPACE="oran-nearrt-ric"
HELM_RELEASE="oran-nearrt-ric"
HELM_EXE="$HOME/bin/helm.exe"
KIND_EXE="$HOME/bin/kind.exe"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

echo "🚀 O-RAN Near-RT RIC Windows Deployment"

# Clean up any existing cluster
log_info "Cleaning up existing cluster..."
$KIND_EXE delete cluster --name "$CLUSTER_NAME" 2>/dev/null || true

# Create cluster
log_info "Creating KIND cluster..."
$KIND_EXE create cluster --name "$CLUSTER_NAME"

# Setup Helm repos
log_info "Setting up Helm repositories..."
$HELM_EXE repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
$HELM_EXE repo update

# Create namespace
log_info "Creating namespace..."
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml  < /dev/null |  kubectl apply -f -

# Deploy O-RAN components
log_info "Deploying O-RAN components..."
if [[ -d "k8s/oran" ]]; then
    kubectl apply -f k8s/oran/ -n "$NAMESPACE"
fi

# Deploy additional components
if [[ -f "k8s/fl-coordinator-deployment.yaml" ]]; then
    kubectl apply -f k8s/fl-coordinator-deployment.yaml -n "$NAMESPACE"
fi

if [[ -f "k8s/xapp-dashboard-deployment.yaml" ]]; then
    kubectl apply -f k8s/xapp-dashboard-deployment.yaml -n "$NAMESPACE"
fi

# Wait for pods
log_info "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod --all -n "$NAMESPACE" --timeout=300s || true

# Show status
log_info "Deployment status:"
kubectl get pods -n "$NAMESPACE" -o wide

log_success "🎉 O-RAN Near-RT RIC deployed successfully!"
echo "📊 Cluster info:"
kubectl cluster-info
echo "🔧 To access services, use kubectl port-forward"
echo "🗑️ To cleanup: $KIND_EXE delete cluster --name $CLUSTER_NAME"
