#!/bin/bash

# O-RAN Near-RT RIC Platform Prerequisites Checker
# This script checks and provides installation instructions for all required dependencies

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_header() {
    echo -e "${CYAN}=== $1 ===${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Get OS information
get_os_info() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

# Get package manager
get_package_manager() {
    local os=$(get_os_info)
    
    case $os in
        linux)
            if command_exists apt-get; then
                echo "apt"
            elif command_exists yum; then
                echo "yum"
            elif command_exists dnf; then
                echo "dnf"
            elif command_exists pacman; then
                echo "pacman"
            else
                echo "unknown"
            fi
            ;;
        macos)
            if command_exists brew; then
                echo "brew"
            else
                echo "none"
            fi
            ;;
        windows)
            if command_exists choco; then
                echo "choco"
            elif command_exists winget; then
                echo "winget"
            else
                echo "none"
            fi
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Version comparison
version_ge() {
    printf '%s\n%s\n' "$2" "$1" | sort -V -C
}

# Check Docker
check_docker() {
    log_header "Docker Engine"
    
    if command_exists docker; then
        local version=$(docker version --format '{{.Server.Version}}' 2>/dev/null || echo "unknown")
        if [[ "$version" != "unknown" ]]; then
            log_success "Docker installed: $version"
            if version_ge "$version" "20.10.0"; then
                log_success "Docker version meets requirements (≥20.10.0)"
            else
                log_warning "Docker version $version is below recommended 20.10.0"
            fi
        else
            log_error "Docker daemon not running"
            echo "  Start Docker daemon: sudo systemctl start docker"
        fi
    else
        log_error "Docker not installed"
        show_docker_install_instructions
    fi
    echo
}

show_docker_install_instructions() {
    local os=$(get_os_info)
    local pm=$(get_package_manager)
    
    echo "  📦 Installation Instructions:"
    case $os in
        linux)
            case $pm in
                apt)
                    echo "    curl -fsSL https://get.docker.com -o get-docker.sh"
                    echo "    sudo sh get-docker.sh"
                    echo "    sudo usermod -aG docker \$USER"
                    ;;
                yum|dnf)
                    echo "    sudo $pm install -y docker"
                    echo "    sudo systemctl enable --now docker"
                    echo "    sudo usermod -aG docker \$USER"
                    ;;
                *)
                    echo "    Visit: https://docs.docker.com/engine/install/"
                    ;;
            esac
            ;;
        macos)
            if [[ "$pm" == "brew" ]]; then
                echo "    brew install --cask docker"
            else
                echo "    Download Docker Desktop: https://docker.com/products/docker-desktop"
            fi
            ;;
        windows)
            case $pm in
                choco)
                    echo "    choco install docker-desktop"
                    ;;
                winget)
                    echo "    winget install Docker.DockerDesktop"
                    ;;
                *)
                    echo "    Download Docker Desktop: https://docker.com/products/docker-desktop"
                    ;;
            esac
            ;;
    esac
}

# Check kubectl
check_kubectl() {
    log_header "kubectl (Kubernetes CLI)"
    
    if command_exists kubectl; then
        local version=$(kubectl version --client --short 2>/dev/null | cut -d' ' -f3 | tr -d 'v')
        log_success "kubectl installed: $version"
        if version_ge "$version" "1.21.0"; then
            log_success "kubectl version meets requirements (≥1.21.0)"
        else
            log_warning "kubectl version $version is below required 1.21.0"
        fi
    else
        log_error "kubectl not installed"
        show_kubectl_install_instructions
    fi
    echo
}

show_kubectl_install_instructions() {
    local os=$(get_os_info)
    local pm=$(get_package_manager)
    
    echo "  📦 Installation Instructions:"
    case $os in
        linux)
            echo "    curl -LO \"https://dl.k8s.io/release/\$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl\""
            echo "    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl"
            ;;
        macos)
            if [[ "$pm" == "brew" ]]; then
                echo "    brew install kubectl"
            else
                echo "    curl -LO \"https://dl.k8s.io/release/\$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl\""
                echo "    chmod +x ./kubectl && sudo mv ./kubectl /usr/local/bin/kubectl"
            fi
            ;;
        windows)
            case $pm in
                choco)
                    echo "    choco install kubernetes-cli"
                    ;;
                winget)
                    echo "    winget install Kubernetes.kubectl"
                    ;;
                *)
                    echo "    Visit: https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/"
                    ;;
            esac
            ;;
    esac
}

# Check KIND
check_kind() {
    log_header "KIND (Kubernetes in Docker)"
    
    if command_exists kind; then
        local version=$(kind version | cut -d' ' -f2)
        log_success "KIND installed: $version"
        if version_ge "$version" "0.17.0"; then
            log_success "KIND version meets requirements (≥0.17.0)"
        else
            log_warning "KIND version $version is below recommended 0.17.0"
        fi
    else
        log_error "KIND not installed"
        show_kind_install_instructions
    fi
    echo
}

show_kind_install_instructions() {
    local os=$(get_os_info)
    local pm=$(get_package_manager)
    
    echo "  📦 Installation Instructions:"
    case $os in
        linux)
            echo "    curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64"
            echo "    chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind"
            ;;
        macos)
            if [[ "$pm" == "brew" ]]; then
                echo "    brew install kind"
            else
                echo "    curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-darwin-amd64"
                echo "    chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind"
            fi
            ;;
        windows)
            case $pm in
                choco)
                    echo "    choco install kind"
                    ;;
                winget)
                    echo "    winget install Kubernetes.kind"
                    ;;
                *)
                    echo "    curl.exe -Lo kind-windows-amd64.exe https://kind.sigs.k8s.io/dl/latest/kind-windows-amd64"
                    echo "    Move kind-windows-amd64.exe to your PATH as kind.exe"
                    ;;
            esac
            ;;
    esac
}

# Check Helm
check_helm() {
    log_header "Helm (Package Manager for Kubernetes)"
    
    if command_exists helm; then
        local version=$(helm version --short | cut -d'+' -f1 | tr -d 'v')
        log_success "Helm installed: $version"
        if version_ge "$version" "3.8.0"; then
            log_success "Helm version meets requirements (≥3.8.0)"
        else
            log_warning "Helm version $version is below recommended 3.8.0"
        fi
    else
        log_error "Helm not installed"
        show_helm_install_instructions
    fi
    echo
}

show_helm_install_instructions() {
    local os=$(get_os_info)
    local pm=$(get_package_manager)
    
    echo "  📦 Installation Instructions:"
    case $os in
        linux)
            echo "    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
            ;;
        macos)
            if [[ "$pm" == "brew" ]]; then
                echo "    brew install helm"
            else
                echo "    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
            fi
            ;;
        windows)
            case $pm in
                choco)
                    echo "    choco install kubernetes-helm"
                    ;;
                winget)
                    echo "    winget install Helm.Helm"
                    ;;
                *)
                    echo "    Visit: https://helm.sh/docs/intro/install/"
                    ;;
            esac
            ;;
    esac
}

# Check Go
check_go() {
    log_header "Go Programming Language"
    
    if command_exists go; then
        local version=$(go version | awk '{print $3}' | tr -d 'go')
        log_success "Go installed: $version"
        if version_ge "$version" "1.17.0"; then
            log_success "Go version meets requirements (≥1.17.0)"
        else
            log_warning "Go version $version is below required 1.17.0"
        fi
    else
        log_error "Go not installed"
        show_go_install_instructions
    fi
    echo
}

show_go_install_instructions() {
    local os=$(get_os_info)
    local pm=$(get_package_manager)
    
    echo "  📦 Installation Instructions:"
    case $os in
        linux)
            echo "    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz"
            echo "    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz"
            echo "    echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
            ;;
        macos)
            if [[ "$pm" == "brew" ]]; then
                echo "    brew install go"
            else
                echo "    Download from: https://go.dev/dl/ (macOS installer)"
            fi
            ;;
        windows)
            case $pm in
                choco)
                    echo "    choco install golang"
                    ;;
                winget)
                    echo "    winget install GoLang.Go"
                    ;;
                *)
                    echo "    Download from: https://go.dev/dl/ (Windows installer)"
                    ;;
            esac
            ;;
    esac
}

# Check Node.js
check_node() {
    log_header "Node.js Runtime"
    
    if command_exists node; then
        local version=$(node --version | tr -d 'v')
        log_success "Node.js installed: $version"
        if version_ge "$version" "16.14.0"; then
            log_success "Node.js version meets requirements (≥16.14.0)"
        else
            log_warning "Node.js version $version is below required 16.14.0"
        fi
    else
        log_error "Node.js not installed"
        show_node_install_instructions
    fi
    echo
}

show_node_install_instructions() {
    local os=$(get_os_info)
    local pm=$(get_package_manager)
    
    echo "  📦 Installation Instructions:"
    echo "  🎯 Recommended: Use Node Version Manager (nvm)"
    case $os in
        linux|macos)
            echo "    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"
            echo "    source ~/.bashrc"
            echo "    nvm install node"
            echo "    nvm use node"
            ;;
        windows)
            echo "    Download nvm-windows: https://github.com/coreybutler/nvm-windows"
            echo "    Or use package manager:"
            case $pm in
                choco)
                    echo "    choco install nodejs"
                    ;;
                winget)
                    echo "    winget install OpenJS.NodeJS"
                    ;;
                *)
                    echo "    Download from: https://nodejs.org/en/download/"
                    ;;
            esac
            ;;
    esac
}

# Check npm
check_npm() {
    log_header "npm (Node Package Manager)"
    
    if command_exists npm; then
        local version=$(npm --version)
        log_success "npm installed: $version"
        if version_ge "$version" "8.5.0"; then
            log_success "npm version meets requirements (≥8.5.0)"
        else
            log_warning "npm version $version is below recommended 8.5.0"
            echo "  🔄 Update npm: npm install -g npm@latest"
        fi
    else
        log_error "npm not installed (usually comes with Node.js)"
        echo "  🔄 npm is typically installed with Node.js"
    fi
    echo
}

# Check additional tools
check_additional_tools() {
    log_header "Additional Development Tools"
    
    # Check curl
    if command_exists curl; then
        log_success "curl available"
    else
        log_warning "curl not found (recommended for downloads)"
    fi
    
    # Check git
    if command_exists git; then
        local version=$(git --version | cut -d' ' -f3)
        log_success "git installed: $version"
    else
        log_warning "git not found (required for development)"
    fi
    
    # Check make
    if command_exists make; then
        log_success "make available"
    else
        log_warning "make not found (required for building)"
    fi
    
    # Check envsubst
    if command_exists envsubst; then
        log_success "envsubst available"
    else
        log_warning "envsubst not found (part of gettext, used for templating)"
    fi
    
    echo
}

# Check system resources
check_system_resources() {
    log_header "System Resources"
    
    local os=$(get_os_info)
    
    case $os in
        linux)
            # Check memory (minimum 8GB recommended)
            local mem_kb=$(grep MemTotal /proc/meminfo | awk '{print $2}')
            local mem_gb=$((mem_kb / 1024 / 1024))
            
            if [[ $mem_gb -ge 8 ]]; then
                log_success "Memory: ${mem_gb}GB (sufficient)"
            elif [[ $mem_gb -ge 4 ]]; then
                log_warning "Memory: ${mem_gb}GB (minimum, 8GB+ recommended)"
            else
                log_error "Memory: ${mem_gb}GB (insufficient, minimum 4GB required)"
            fi
            
            # Check CPU cores
            local cpu_cores=$(nproc)
            if [[ $cpu_cores -ge 4 ]]; then
                log_success "CPU cores: $cpu_cores (sufficient)"
            elif [[ $cpu_cores -ge 2 ]]; then
                log_warning "CPU cores: $cpu_cores (minimum, 4+ recommended)"
            else
                log_error "CPU cores: $cpu_cores (insufficient, minimum 2 required)"
            fi
            
            # Check disk space
            local disk_space=$(df -BG . | tail -1 | awk '{print $4}' | tr -d 'G')
            if [[ $disk_space -ge 50 ]]; then
                log_success "Available disk space: ${disk_space}GB (sufficient)"
            elif [[ $disk_space -ge 20 ]]; then
                log_warning "Available disk space: ${disk_space}GB (minimum, 50GB+ recommended)"
            else
                log_error "Available disk space: ${disk_space}GB (insufficient, minimum 20GB required)"
            fi
            ;;
        macos)
            # Check memory
            local mem_bytes=$(sysctl -n hw.memsize)
            local mem_gb=$((mem_bytes / 1024 / 1024 / 1024))
            
            if [[ $mem_gb -ge 8 ]]; then
                log_success "Memory: ${mem_gb}GB (sufficient)"
            elif [[ $mem_gb -ge 4 ]]; then
                log_warning "Memory: ${mem_gb}GB (minimum, 8GB+ recommended)"
            else
                log_error "Memory: ${mem_gb}GB (insufficient, minimum 4GB required)"
            fi
            
            # Check CPU cores
            local cpu_cores=$(sysctl -n hw.ncpu)
            if [[ $cpu_cores -ge 4 ]]; then
                log_success "CPU cores: $cpu_cores (sufficient)"
            elif [[ $cpu_cores -ge 2 ]]; then
                log_warning "CPU cores: $cpu_cores (minimum, 4+ recommended)"
            else
                log_error "CPU cores: $cpu_cores (insufficient, minimum 2 required)"
            fi
            ;;
        *)
            log_info "System resource check not available for this OS"
            ;;
    esac
    
    echo
}

# Main function
main() {
    log_info "🔍 O-RAN Near-RT RIC Platform Prerequisites Check"
    echo ""
    
    # Check all prerequisites
    check_docker
    check_kubectl
    check_kind
    check_helm
    check_go
    check_node
    check_npm
    check_additional_tools
    check_system_resources
    
    # Summary
    log_header "Summary"
    log_info "Prerequisites check completed!"
    echo ""
    log_info "📚 Next Steps:"
    echo "  1. Install any missing dependencies listed above"
    echo "  2. Restart your terminal to reload PATH"
    echo "  3. Run './scripts/setup.sh' to deploy the platform"
    echo ""
    log_info "📖 Resources:"
    echo "  - O-RAN Documentation: docs/README.md"
    echo "  - Developer Guide: dashboard-master/dashboard-master/docs/developer/getting-started.md"
    echo "  - Troubleshooting: docs/operations/troubleshooting.md"
}

# Run main function
main