#!/bin/bash

# O-RAN Near-RT RIC Platform Setup Script
# This script provides a one-command setup for the complete O-RAN platform

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global variables
CLUSTER_NAME="near-rt-ric"
NAMESPACE_DASHBOARD="kubernetes-dashboard"
NAMESPACE_XAPP="xapp-dashboard"
NAMESPACE_ORAN="oran-nearrt-ric"
NAMESPACE_MONITORING="monitoring"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_deps=()
    
    # Check required tools
    if ! command_exists docker; then
        missing_deps+=("docker")
    fi
    
    if ! command_exists kubectl; then
        missing_deps+=("kubectl")
    fi
    
    if ! command_exists kind; then
        missing_deps+=("kind")
    fi
    
    if ! command_exists helm; then
        missing_deps+=("helm")
    fi
    
    if ! command_exists go; then
        missing_deps+=("go")
    fi
    
    if ! command_exists node; then
        missing_deps+=("node")
    fi
    
    if ! command_exists npm; then
        missing_deps+=("npm")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_info "Please install missing dependencies and run this script again."
        log_info "Run './scripts/check-prerequisites.sh' for detailed installation instructions."
        exit 1
    fi
    
    # Check version requirements
    log_info "Checking version requirements..."
    
    # Check Docker version (minimum 20.10)
    docker_version=$(docker version --format '{{.Server.Version}}' 2>/dev/null | cut -d. -f1,2)
    if (( $(echo "$docker_version < 20.10" | bc -l) )); then
        log_warning "Docker version $docker_version is below recommended 20.10+"
    fi
    
    # Check Kubernetes version  
    kubectl_version=$(kubectl version --client --short 2>/dev/null | cut -d' ' -f3 | tr -d 'v' | cut -d. -f1,2)
    if (( $(echo "$kubectl_version < 1.21" | bc -l) )); then
        log_warning "kubectl version $kubectl_version is below required 1.21+"
    fi
    
    # Check Go version (minimum 1.17)
    go_version=$(go version | awk '{print $3}' | tr -d 'go' | cut -d. -f1,2)
    if (( $(echo "$go_version < 1.17" | bc -l) )); then
        log_warning "Go version $go_version is below required 1.17+"
    fi
    
    # Check Node.js version (minimum 16.14)
    node_version=$(node --version | tr -d 'v' | cut -d. -f1,2)
    if (( $(echo "$node_version < 16.14" | bc -l) )); then
        log_warning "Node.js version $node_version is below required 16.14+"
    fi
    
    log_success "Prerequisites check completed"
}

# Create KIND cluster
create_cluster() {
    log_info "Creating KIND cluster: $CLUSTER_NAME"
    
    # Check if cluster already exists
    if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
        log_warning "Cluster $CLUSTER_NAME already exists"
        read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            kind delete cluster --name="$CLUSTER_NAME"
        else
            log_info "Using existing cluster"
            return 0
        fi
    fi
    
    # Create cluster with custom configuration
    kind create cluster --name="$CLUSTER_NAME" --config=kind-config.yaml
    
    # Wait for cluster to be ready
    log_info "Waiting for cluster to be ready..."
    kubectl cluster-info --context="kind-$CLUSTER_NAME"
    kubectl wait --for=condition=Ready nodes --all --timeout=300s
    
    log_success "KIND cluster created successfully"
}

# Setup namespaces
setup_namespaces() {
    log_info "Setting up namespaces..."
    
    # Create namespaces
    kubectl create namespace "$NAMESPACE_DASHBOARD" --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace "$NAMESPACE_XAPP" --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace "$NAMESPACE_ORAN" --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace "$NAMESPACE_MONITORING" --dry-run=client -o yaml | kubectl apply -f -
    
    # Label namespaces
    kubectl label namespace "$NAMESPACE_DASHBOARD" app.kubernetes.io/name=kubernetes-dashboard --overwrite
    kubectl label namespace "$NAMESPACE_XAPP" app.kubernetes.io/name=xapp-dashboard --overwrite
    kubectl label namespace "$NAMESPACE_ORAN" app.kubernetes.io/name=oran-nearrt-ric --overwrite
    kubectl label namespace "$NAMESPACE_MONITORING" app.kubernetes.io/name=monitoring --overwrite
    
    log_success "Namespaces created and labeled"
}

# Deploy monitoring stack
deploy_monitoring() {
    log_info "Deploying monitoring stack (Prometheus, Grafana)..."
    
    # Add Prometheus Helm repository
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    
    # Deploy Prometheus
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
        --namespace "$NAMESPACE_MONITORING" \
        --set prometheus.prometheusSpec.retention=7d \
        --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=10Gi \
        --set grafana.adminPassword=admin123 \
        --set grafana.persistence.enabled=true \
        --set grafana.persistence.size=5Gi \
        --wait --timeout=600s
    
    log_success "Monitoring stack deployed"
}

# Deploy main dashboard
deploy_main_dashboard() {
    log_info "Deploying main dashboard..."
    
    # Build main dashboard
    cd dashboard-master/dashboard-master
    make build
    
    # Deploy using existing manifests
    kubectl apply -f aio/deploy/recommended.yaml
    
    # Wait for deployment
    kubectl wait --for=condition=available deployment/kubernetes-dashboard \
        --namespace="$NAMESPACE_DASHBOARD" --timeout=300s
    
    cd ../../
    log_success "Main dashboard deployed"
}

# Deploy xApp dashboard  
deploy_xapp_dashboard() {
    log_info "Deploying xApp dashboard..."
    
    # Build xApp dashboard
    cd xAPP_dashboard-master
    npm install --silent
    npm run build --silent
    
    # Create deployment from template
    envsubst < ../k8s/xapp-dashboard-deployment.yaml | kubectl apply -f -
    
    # Wait for deployment
    kubectl wait --for=condition=available deployment/xapp-dashboard \
        --namespace="$NAMESPACE_XAPP" --timeout=300s
    
    cd ../
    log_success "xApp dashboard deployed"
}

# Deploy O-RAN components
deploy_oran_components() {
    log_info "Deploying O-RAN Near-RT RIC components..."
    
    # Deploy E2 simulator
    kubectl apply -f k8s/oran/e2-simulator.yaml
    
    # Deploy O-CU simulator
    kubectl apply -f k8s/oran/o-cu-simulator.yaml
    
    # Deploy O-DU simulator  
    kubectl apply -f k8s/oran/o-du-simulator.yaml
    
    # Deploy sample xApps
    kubectl apply -f k8s/oran/sample-xapps/
    
    # Wait for components to be ready
    kubectl wait --for=condition=available deployment/e2-simulator \
        --namespace="$NAMESPACE_ORAN" --timeout=300s
    
    log_success "O-RAN components deployed"
}

# Setup port forwarding
setup_port_forwarding() {
    log_info "Setting up port forwarding..."
    
    # Create port-forward script
    cat > port-forward.sh << 'EOF'
#!/bin/bash
# Port forwarding script for O-RAN Near-RT RIC platform

echo "Starting port forwarding..."

# Main Dashboard
kubectl port-forward -n kubernetes-dashboard service/kubernetes-dashboard 8080:443 &
MAIN_PID=$!

# xApp Dashboard  
kubectl port-forward -n xapp-dashboard service/xapp-dashboard 4200:80 &
XAPP_PID=$!

# Grafana
kubectl port-forward -n monitoring service/prometheus-grafana 3000:80 &
GRAFANA_PID=$!

# Prometheus
kubectl port-forward -n monitoring service/prometheus-kube-prometheus-prometheus 9090:9090 &
PROMETHEUS_PID=$!

echo "Port forwarding started:"
echo "  Main Dashboard: https://localhost:8080"
echo "  xApp Dashboard: http://localhost:4200"  
echo "  Grafana: http://localhost:3000 (admin/admin123)"
echo "  Prometheus: http://localhost:9090"
echo ""
echo "Press Ctrl+C to stop all port forwarding"

# Function to kill all background processes
cleanup() {
    echo "Stopping port forwarding..."
    kill $MAIN_PID $XAPP_PID $GRAFANA_PID $PROMETHEUS_PID 2>/dev/null
    exit 0
}

# Set trap to catch interrupt signal
trap cleanup INT

# Wait for interrupt
while true; do
    sleep 1
done
EOF
    
    chmod +x port-forward.sh
    log_success "Port forwarding script created"
}

# Create authentication token
create_admin_user() {
    log_info "Creating admin user for dashboard access..."
    
    # Create service account and cluster role binding
    kubectl apply -f - << 'EOF'
apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-user
  namespace: kubernetes-dashboard
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: admin-user
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: admin-user
  namespace: kubernetes-dashboard
EOF
    
    # Wait for secret to be created
    sleep 5
    
    # Get token
    TOKEN=$(kubectl -n kubernetes-dashboard get secret $(kubectl -n kubernetes-dashboard get sa/admin-user -o jsonpath="{.secrets[0].name}") -o go-template="{{.data.token | base64decode}}")
    
    # Save token to file
    echo "$TOKEN" > dashboard-admin-token.txt
    
    log_success "Admin user created. Token saved to dashboard-admin-token.txt"
}

# Display access information
display_access_info() {
    log_success "🎉 O-RAN Near-RT RIC Platform Setup Complete!"
    echo ""
    echo "📊 Access URLs:"
    echo "  Main Dashboard: https://localhost:8080"
    echo "  xApp Dashboard: http://localhost:4200"
    echo "  Grafana: http://localhost:3000 (admin/admin123)"
    echo "  Prometheus: http://localhost:9090"
    echo ""
    echo "🔐 Dashboard Authentication:"
    echo "  Use the token from: dashboard-admin-token.txt"
    echo ""
    echo "🚀 Next Steps:"
    echo "  1. Run './port-forward.sh' to start port forwarding"
    echo "  2. Open https://localhost:8080 in your browser"
    echo "  3. Select 'Token' and paste the token from dashboard-admin-token.txt"
    echo "  4. Explore the O-RAN platform components"
    echo ""
    echo "📚 Documentation:"
    echo "  - User Guide: docs/user/README.md"
    echo "  - Developer Guide: dashboard-master/dashboard-master/docs/developer/README.md"
    echo "  - API Reference: docs/developer/api-reference.md"
    echo ""
    echo "🛠️ Management Commands:"
    echo "  - View logs: kubectl logs -f deployment/kubernetes-dashboard -n kubernetes-dashboard"
    echo "  - Scale components: kubectl scale deployment/xapp-dashboard --replicas=2 -n xapp-dashboard"
    echo "  - Cleanup: kind delete cluster --name=$CLUSTER_NAME"
}

# Main setup function
main() {
    log_info "🌐 Starting O-RAN Near-RT RIC Platform Setup"
    echo "This script will set up a complete O-RAN platform with:"
    echo "  - Kubernetes cluster (KIND)"
    echo "  - Main Dashboard (Kubernetes management)"
    echo "  - xApp Dashboard (xApp lifecycle management)"
    echo "  - Monitoring stack (Prometheus + Grafana)"
    echo "  - O-RAN simulators (E2, O-CU, O-DU)"
    echo "  - Sample xApps"
    echo ""
    
    # Confirm setup
    read -p "Continue with setup? (Y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        log_info "Setup cancelled"
        exit 0
    fi
    
    # Run setup steps
    check_prerequisites
    create_cluster
    setup_namespaces
    deploy_monitoring
    deploy_main_dashboard
    deploy_xapp_dashboard
    deploy_oran_components
    setup_port_forwarding
    create_admin_user
    display_access_info
    
    log_success "Setup completed successfully! 🎉"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "O-RAN Near-RT RIC Platform Setup Script"
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --check-only   Only check prerequisites"
        echo "  --cleanup      Delete the cluster and exit"
        echo ""
        exit 0
        ;;
    --check-only)
        check_prerequisites
        exit 0
        ;;
    --cleanup)
        log_info "Cleaning up cluster: $CLUSTER_NAME"
        kind delete cluster --name="$CLUSTER_NAME"
        log_success "Cleanup completed"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown option: $1"
        echo "Run '$0 --help' for usage information"
        exit 1
        ;;
esac