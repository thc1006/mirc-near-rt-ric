# O-RAN Near-RT RIC Platform Setup Script for Windows
# PowerShell version of the setup script for Windows environments

param(
    [switch]$CheckOnly,
    [switch]$Cleanup,
    [switch]$Help
)

# Color output functions
function Write-Success { param($Message) Write-Host "✓ $Message" -ForegroundColor Green }
function Write-Error { param($Message) Write-Host "✗ $Message" -ForegroundColor Red }
function Write-Warning { param($Message) Write-Host "⚠ $Message" -ForegroundColor Yellow }
function Write-Info { param($Message) Write-Host "ℹ $Message" -ForegroundColor Blue }

# Global variables
$CLUSTER_NAME = "near-rt-ric"
$NAMESPACE_DASHBOARD = "kubernetes-dashboard"
$NAMESPACE_XAPP = "xapp-dashboard"
$NAMESPACE_ORAN = "oran-nearrt-ric"
$NAMESPACE_MONITORING = "monitoring"

function Show-Help {
    Write-Host @"
O-RAN Near-RT RIC Platform Setup Script for Windows

Usage: .\setup.ps1 [OPTIONS]

Options:
  -CheckOnly    Only check prerequisites
  -Cleanup      Delete the cluster and exit  
  -Help         Show this help message

Description:
  This script sets up a complete O-RAN platform with:
  - Kubernetes cluster (KIND)
  - Main Dashboard (Kubernetes management)
  - xApp Dashboard (xApp lifecycle management)
  - Monitoring stack (Prometheus + Grafana)
  - O-RAN simulators (E2, O-CU, O-DU)
  - Sample xApps

Prerequisites:
  - Docker Desktop for Windows
  - kubectl
  - kind
  - helm
  - Go 1.17+
  - Node.js 16.14.2+
  - npm 8.5.0+

Examples:
  .\setup.ps1                # Full setup
  .\setup.ps1 -CheckOnly    # Check prerequisites only
  .\setup.ps1 -Cleanup      # Clean up cluster
"@
}

function Test-Command {
    param($CommandName)
    try {
        Get-Command $CommandName -ErrorAction Stop | Out-Null
        return $true
    } catch {
        return $false
    }
}

function Test-Prerequisites {
    Write-Info "Checking prerequisites..."
    
    $missingDeps = @()
    $warnings = @()
    
    # Check required tools
    if (-not (Test-Command "docker")) { $missingDeps += "docker" }
    if (-not (Test-Command "kubectl")) { $missingDeps += "kubectl" }
    if (-not (Test-Command "kind")) { $missingDeps += "kind" }
    if (-not (Test-Command "helm")) { $missingDeps += "helm" }
    if (-not (Test-Command "go")) { $missingDeps += "go" }
    if (-not (Test-Command "node")) { $missingDeps += "node" }
    if (-not (Test-Command "npm")) { $missingDeps += "npm" }
    
    if ($missingDeps.Count -gt 0) {
        Write-Error "Missing required dependencies: $($missingDeps -join ', ')"
        Write-Info "Please install missing dependencies and run this script again."
        Write-Info "Run '.\scripts\check-prerequisites.ps1' for detailed installation instructions."
        return $false
    }
    
    # Check version requirements
    Write-Info "Checking version requirements..."
    
    # Check Docker version
    try {
        $dockerVersion = docker version --format '{{.Server.Version}}' 2>$null
        if ($dockerVersion -lt [version]"20.10") {
            $warnings += "Docker version $dockerVersion is below recommended 20.10+"
        }
    } catch {
        Write-Warning "Could not check Docker version"
    }
    
    # Check kubectl version
    try {
        $kubectlVersion = (kubectl version --client -o json 2>$null | ConvertFrom-Json).clientVersion.gitVersion -replace '^v'
        if ([version]$kubectlVersion -lt [version]"1.21.0") {
            $warnings += "kubectl version $kubectlVersion is below required 1.21+"
        }
    } catch {
        Write-Warning "Could not check kubectl version"
    }
    
    # Check Go version
    try {
        $goVersion = (go version) -replace '.*go(\d+\.\d+).*','$1'
        if ([version]$goVersion -lt [version]"1.17") {
            $warnings += "Go version $goVersion is below required 1.17+"
        }
    } catch {
        Write-Warning "Could not check Go version"
    }
    
    # Check Node.js version
    try {
        $nodeVersion = (node --version) -replace '^v'
        if ([version]$nodeVersion -lt [version]"16.14") {
            $warnings += "Node.js version $nodeVersion is below required 16.14+"
        }
    } catch {
        Write-Warning "Could not check Node.js version"
    }
    
    # Display warnings
    foreach ($warning in $warnings) {
        Write-Warning $warning
    }
    
    Write-Success "Prerequisites check completed"
    return $true
}

function New-KindCluster {
    Write-Info "Creating KIND cluster: $CLUSTER_NAME"
    
    # Check if cluster already exists
    $existingClusters = kind get clusters 2>$null
    if ($existingClusters -contains $CLUSTER_NAME) {
        Write-Warning "Cluster $CLUSTER_NAME already exists"
        $response = Read-Host "Do you want to delete and recreate it? (y/N)"
        if ($response -eq 'y' -or $response -eq 'Y') {
            kind delete cluster --name=$CLUSTER_NAME
        } else {
            Write-Info "Using existing cluster"
            return $true
        }
    }
    
    # Create cluster with custom configuration
    if (-not (Test-Path "kind-config.yaml")) {
        Write-Error "kind-config.yaml not found. Please ensure you're running from the repository root."
        return $false
    }
    
    kind create cluster --name=$CLUSTER_NAME --config=kind-config.yaml
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to create KIND cluster"
        return $false
    }
    
    # Wait for cluster to be ready
    Write-Info "Waiting for cluster to be ready..."
    kubectl cluster-info --context="kind-$CLUSTER_NAME"
    kubectl wait --for=condition=Ready nodes --all --timeout=300s
    
    Write-Success "KIND cluster created successfully"
    return $true
}

function New-Namespaces {
    Write-Info "Setting up namespaces..."
    
    # Create namespaces
    kubectl create namespace $NAMESPACE_DASHBOARD --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace $NAMESPACE_XAPP --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace $NAMESPACE_ORAN --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace $NAMESPACE_MONITORING --dry-run=client -o yaml | kubectl apply -f -
    
    # Label namespaces
    kubectl label namespace $NAMESPACE_DASHBOARD app.kubernetes.io/name=kubernetes-dashboard --overwrite
    kubectl label namespace $NAMESPACE_XAPP app.kubernetes.io/name=xapp-dashboard --overwrite
    kubectl label namespace $NAMESPACE_ORAN app.kubernetes.io/name=oran-nearrt-ric --overwrite
    kubectl label namespace $NAMESPACE_MONITORING app.kubernetes.io/name=monitoring --overwrite
    
    Write-Success "Namespaces created and labeled"
    return $true
}

function Install-MonitoringStack {
    Write-Info "Deploying monitoring stack (Prometheus, Grafana)..."
    
    # Add Helm repositories
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    
    # Deploy Prometheus
    helm upgrade --install prometheus prometheus-community/kube-prometheus-stack `
        --namespace $NAMESPACE_MONITORING `
        --set prometheus.prometheusSpec.retention=7d `
        --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=10Gi `
        --set grafana.adminPassword=admin123 `
        --set grafana.persistence.enabled=true `
        --set grafana.persistence.size=5Gi `
        --wait --timeout=600s
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to deploy monitoring stack"
        return $false
    }
    
    Write-Success "Monitoring stack deployed"
    return $true
}

function Install-MainDashboard {
    Write-Info "Deploying main dashboard..."
    
    # Build main dashboard
    Set-Location "dashboard-master\dashboard-master"
    make build
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build main dashboard"
        Set-Location "..\..\"
        return $false
    }
    
    # Deploy using existing manifests
    kubectl apply -f aio\deploy\recommended.yaml
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to deploy main dashboard"
        Set-Location "..\..\"
        return $false
    }
    
    # Wait for deployment
    kubectl wait --for=condition=available deployment/kubernetes-dashboard `
        --namespace=$NAMESPACE_DASHBOARD --timeout=300s
    
    Set-Location "..\..\"
    Write-Success "Main dashboard deployed"
    return $true
}

function Install-XappDashboard {
    Write-Info "Deploying xApp dashboard..."
    
    # Build xApp dashboard
    Set-Location "xAPP_dashboard-master"
    
    # Install dependencies and build
    npm install --silent
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to install xApp dashboard dependencies"
        Set-Location "..\"
        return $false
    }
    
    npm run build --silent
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build xApp dashboard"
        Set-Location "..\"
        return $false
    }
    
    Set-Location "..\"
    
    # Create deployment from template
    if (Test-Path "k8s\xapp-dashboard-deployment.yaml") {
        kubectl apply -f k8s\xapp-dashboard-deployment.yaml
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to deploy xApp dashboard"
            return $false
        }
        
        # Wait for deployment
        kubectl wait --for=condition=available deployment/xapp-dashboard `
            --namespace=$NAMESPACE_XAPP --timeout=300s
    } else {
        Write-Warning "xApp dashboard Kubernetes deployment not found, skipping K8s deployment"
    }
    
    Write-Success "xApp dashboard deployed"
    return $true
}

function Install-OranComponents {
    Write-Info "Deploying O-RAN Near-RT RIC components..."
    
    # Deploy E2 simulator
    if (Test-Path "k8s\oran\e2-simulator.yaml") {
        kubectl apply -f k8s\oran\e2-simulator.yaml
    }
    
    # Deploy sample xApps
    if (Test-Path "k8s\oran\sample-xapps") {
        kubectl apply -f k8s\oran\sample-xapps\
    }
    
    # Wait for components to be ready (with timeout)
    Write-Info "Waiting for O-RAN components to be ready..."
    Start-Sleep -Seconds 30
    
    Write-Success "O-RAN components deployed"
    return $true
}

function New-PortForwardScript {
    Write-Info "Setting up port forwarding..."
    
    # Create PowerShell port-forward script
    $portForwardScript = @'
# Port forwarding script for O-RAN Near-RT RIC platform

Write-Host "Starting port forwarding..."

# Start port forwarding jobs
$jobs = @()

# Main Dashboard
$jobs += Start-Job -ScriptBlock {
    kubectl port-forward -n kubernetes-dashboard service/kubernetes-dashboard 8080:443
}

# xApp Dashboard  
$jobs += Start-Job -ScriptBlock {
    kubectl port-forward -n xapp-dashboard service/xapp-dashboard 4200:80
}

# Grafana
$jobs += Start-Job -ScriptBlock {
    kubectl port-forward -n monitoring service/prometheus-grafana 3000:80
}

# Prometheus
$jobs += Start-Job -ScriptBlock {
    kubectl port-forward -n monitoring service/prometheus-kube-prometheus-prometheus 9090:9090
}

Write-Host "Port forwarding started:"
Write-Host "  Main Dashboard: https://localhost:8080"
Write-Host "  xApp Dashboard: http://localhost:4200"  
Write-Host "  Grafana: http://localhost:3000 (admin/admin123)"
Write-Host "  Prometheus: http://localhost:9090"
Write-Host ""
Write-Host "Press Ctrl+C to stop all port forwarding"

# Wait for interrupt
try {
    while ($true) {
        Start-Sleep -Seconds 1
    }
} finally {
    Write-Host "Stopping port forwarding..."
    $jobs | Stop-Job
    $jobs | Remove-Job
}
'@
    
    $portForwardScript | Out-File -FilePath "port-forward.ps1" -Encoding utf8
    Write-Success "Port forwarding script created: port-forward.ps1"
    return $true
}

function New-AdminUser {
    Write-Info "Creating admin user for dashboard access..."
    
    # Create service account and cluster role binding
    $adminUserYaml = @'
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
'@
    
    $adminUserYaml | kubectl apply -f -
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to create admin user"
        return $false
    }
    
    # Wait for secret to be created
    Start-Sleep -Seconds 5
    
    # Get token (handle both old and new Kubernetes versions)
    try {
        # Try new method first (K8s 1.24+)
        $token = kubectl create token admin-user -n kubernetes-dashboard 2>$null
        if (-not $token) {
            # Fallback to old method
            $secretName = kubectl -n kubernetes-dashboard get sa/admin-user -o jsonpath="{.secrets[0].name}" 2>$null
            if ($secretName) {
                $token = kubectl -n kubernetes-dashboard get secret $secretName -o go-template="{{.data.token | base64decode}}" 2>$null
            }
        }
        
        if ($token) {
            $token | Out-File -FilePath "dashboard-admin-token.txt" -Encoding utf8
            Write-Success "Admin user created. Token saved to dashboard-admin-token.txt"
        } else {
            Write-Warning "Could not retrieve admin token. You may need to create it manually."
        }
    } catch {
        Write-Warning "Could not retrieve admin token: $($_.Exception.Message)"
    }
    
    return $true
}

function Show-AccessInfo {
    Write-Success "🎉 O-RAN Near-RT RIC Platform Setup Complete!"
    Write-Host ""
    Write-Host "📊 Access URLs:" -ForegroundColor Cyan
    Write-Host "  Main Dashboard: https://localhost:8080"
    Write-Host "  xApp Dashboard: http://localhost:4200"
    Write-Host "  Grafana: http://localhost:3000 (admin/admin123)"
    Write-Host "  Prometheus: http://localhost:9090"
    Write-Host ""
    Write-Host "🔐 Dashboard Authentication:" -ForegroundColor Cyan
    if (Test-Path "dashboard-admin-token.txt") {
        Write-Host "  Use the token from: dashboard-admin-token.txt"
    } else {
        Write-Host "  Create admin token manually or use kubeconfig authentication"
    }
    Write-Host ""
    Write-Host "🚀 Next Steps:" -ForegroundColor Cyan
    Write-Host "  1. Run '.\port-forward.ps1' to start port forwarding"
    Write-Host "  2. Open https://localhost:8080 in your browser"
    Write-Host "  3. Select 'Token' and paste the token from dashboard-admin-token.txt"
    Write-Host "  4. Explore the O-RAN platform components"
    Write-Host ""
    Write-Host "📚 Documentation:" -ForegroundColor Cyan
    Write-Host "  - User Guide: docs\user\README.md"
    Write-Host "  - Developer Guide: dashboard-master\dashboard-master\docs\developer\README.md"
    Write-Host "  - API Reference: docs\developer\api-reference.md"
    Write-Host ""
    Write-Host "🛠️ Management Commands:" -ForegroundColor Cyan
    Write-Host "  - View logs: kubectl logs -f deployment/kubernetes-dashboard -n kubernetes-dashboard"
    Write-Host "  - Scale components: kubectl scale deployment/xapp-dashboard --replicas=2 -n xapp-dashboard"
    Write-Host "  - Cleanup: .\setup.ps1 -Cleanup"
}

function Remove-Cluster {
    Write-Info "Cleaning up cluster: $CLUSTER_NAME"
    kind delete cluster --name=$CLUSTER_NAME
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Cleanup completed"
    } else {
        Write-Error "Failed to cleanup cluster"
    }
}

function Start-Setup {
    Write-Info "🌐 Starting O-RAN Near-RT RIC Platform Setup"
    Write-Host "This script will set up a complete O-RAN platform with:"
    Write-Host "  - Kubernetes cluster (KIND)"
    Write-Host "  - Main Dashboard (Kubernetes management)"
    Write-Host "  - xApp Dashboard (xApp lifecycle management)"
    Write-Host "  - Monitoring stack (Prometheus + Grafana)"
    Write-Host "  - O-RAN simulators (E2, O-CU, O-DU)"
    Write-Host "  - Sample xApps"
    Write-Host ""
    
    # Confirm setup
    $response = Read-Host "Continue with setup? (Y/n)"
    if ($response -eq 'n' -or $response -eq 'N') {
        Write-Info "Setup cancelled"
        return
    }
    
    # Run setup steps
    if (-not (Test-Prerequisites)) { return }
    if (-not (New-KindCluster)) { return }
    if (-not (New-Namespaces)) { return }
    if (-not (Install-MonitoringStack)) { return }
    if (-not (Install-MainDashboard)) { return }
    if (-not (Install-XappDashboard)) { return }
    if (-not (Install-OranComponents)) { return }
    if (-not (New-PortForwardScript)) { return }
    if (-not (New-AdminUser)) { return }
    
    Show-AccessInfo
    Write-Success "Setup completed successfully! 🎉"
}

# Main script logic
if ($Help) {
    Show-Help
    exit 0
}

if ($CheckOnly) {
    Test-Prerequisites
    exit 0
}

if ($Cleanup) {
    Remove-Cluster
    exit 0
}

# Run main setup
Start-Setup