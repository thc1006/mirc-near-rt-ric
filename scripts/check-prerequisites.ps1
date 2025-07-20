# O-RAN Near-RT RIC Platform Prerequisites Checker for Windows
# PowerShell version of the prerequisites checker

# Color output functions
function Write-Success { param($Message) Write-Host "✓ $Message" -ForegroundColor Green }
function Write-Error { param($Message) Write-Host "✗ $Message" -ForegroundColor Red }
function Write-Warning { param($Message) Write-Host "⚠ $Message" -ForegroundColor Yellow }
function Write-Info { param($Message) Write-Host "ℹ $Message" -ForegroundColor Blue }
function Write-Header { param($Message) Write-Host "=== $Message ===" -ForegroundColor Cyan }

function Test-Command {
    param($CommandName)
    try {
        Get-Command $CommandName -ErrorAction Stop | Out-Null
        return $true
    } catch {
        return $false
    }
}

function Test-Version {
    param($Current, $Required)
    try {
        return [version]$Current -ge [version]$Required
    } catch {
        return $false
    }
}

function Get-DockerVersion {
    try {
        $version = docker version --format '{{.Server.Version}}' 2>$null
        return $version
    } catch {
        return $null
    }
}

function Get-KubectlVersion {
    try {
        $output = kubectl version --client -o json 2>$null | ConvertFrom-Json
        return $output.clientVersion.gitVersion -replace '^v'
    } catch {
        return $null
    }
}

function Get-KindVersion {
    try {
        $output = kind version 2>$null
        if ($output -match 'kind v([\d.]+)') {
            return $matches[1]
        }
        return $null
    } catch {
        return $null
    }
}

function Get-HelmVersion {
    try {
        $output = helm version --short 2>$null
        if ($output -match 'v([\d.]+)') {
            return $matches[1]
        }
        return $null
    } catch {
        return $null
    }
}

function Get-GoVersion {
    try {
        $output = go version 2>$null
        if ($output -match 'go([\d.]+)') {
            return $matches[1]
        }
        return $null
    } catch {
        return $null
    }
}

function Get-NodeVersion {
    try {
        $version = node --version 2>$null
        return $version -replace '^v'
    } catch {
        return $null
    }
}

function Get-NpmVersion {
    try {
        return npm --version 2>$null
    } catch {
        return $null
    }
}

function Show-DockerInstall {
    Write-Info "📦 Docker Installation Instructions:"
    Write-Host "  1. Download Docker Desktop for Windows:"
    Write-Host "     https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe"
    Write-Host "  2. Run the installer and follow the setup wizard"
    Write-Host "  3. Restart your computer when prompted"
    Write-Host "  4. Start Docker Desktop from the Start menu"
    Write-Host ""
    Write-Host "  Alternative (with Chocolatey):"
    Write-Host "    choco install docker-desktop"
    Write-Host ""
    Write-Host "  Alternative (with winget):"
    Write-Host "    winget install Docker.DockerDesktop"
}

function Show-KubectlInstall {
    Write-Info "📦 kubectl Installation Instructions:"
    Write-Host "  1. Download kubectl for Windows:"
    Write-Host "     https://dl.k8s.io/release/v1.29.0/bin/windows/amd64/kubectl.exe"
    Write-Host "  2. Add kubectl.exe to your PATH"
    Write-Host ""
    Write-Host "  Alternative (with Chocolatey):"
    Write-Host "    choco install kubernetes-cli"
    Write-Host ""
    Write-Host "  Alternative (with winget):"
    Write-Host "    winget install Kubernetes.kubectl"
    Write-Host ""
    Write-Host "  Alternative (with Scoop):"
    Write-Host "    scoop install kubectl"
}

function Show-KindInstall {
    Write-Info "📦 KIND Installation Instructions:"
    Write-Host "  1. Download KIND for Windows:"
    Write-Host "     https://kind.sigs.k8s.io/dl/v0.20.0/kind-windows-amd64"
    Write-Host "  2. Rename to kind.exe and add to your PATH"
    Write-Host ""
    Write-Host "  Alternative (with Chocolatey):"
    Write-Host "    choco install kind"
    Write-Host ""
    Write-Host "  Alternative (with winget):"
    Write-Host "    winget install Kubernetes.kind"
    Write-Host ""
    Write-Host "  Alternative (with Scoop):"
    Write-Host "    scoop install kind"
}

function Show-HelmInstall {
    Write-Info "📦 Helm Installation Instructions:"
    Write-Host "  1. Download Helm for Windows:"
    Write-Host "     https://get.helm.sh/helm-v3.13.0-windows-amd64.zip"
    Write-Host "  2. Extract and add helm.exe to your PATH"
    Write-Host ""
    Write-Host "  Alternative (with Chocolatey):"
    Write-Host "    choco install kubernetes-helm"
    Write-Host ""
    Write-Host "  Alternative (with winget):"
    Write-Host "    winget install Helm.Helm"
    Write-Host ""
    Write-Host "  Alternative (with Scoop):"
    Write-Host "    scoop install helm"
}

function Show-GoInstall {
    Write-Info "📦 Go Installation Instructions:"
    Write-Host "  1. Download Go for Windows:"
    Write-Host "     https://go.dev/dl/go1.21.5.windows-amd64.msi"
    Write-Host "  2. Run the installer"
    Write-Host "  3. Restart your command prompt/PowerShell"
    Write-Host ""
    Write-Host "  Alternative (with Chocolatey):"
    Write-Host "    choco install golang"
    Write-Host ""
    Write-Host "  Alternative (with winget):"
    Write-Host "    winget install GoLang.Go"
    Write-Host ""
    Write-Host "  Alternative (with Scoop):"
    Write-Host "    scoop install go"
}

function Show-NodeInstall {
    Write-Info "📦 Node.js Installation Instructions:"
    Write-Host "  🎯 Recommended: Use Node Version Manager (nvm-windows)"
    Write-Host "  1. Download nvm-windows:"
    Write-Host "     https://github.com/coreybutler/nvm-windows/releases/latest/download/nvm-setup.zip"
    Write-Host "  2. Extract and run nvm-setup.exe"
    Write-Host "  3. Install and use Node.js:"
    Write-Host "     nvm install lts"
    Write-Host "     nvm use lts"
    Write-Host ""
    Write-Host "  Alternative (direct download):"
    Write-Host "    https://nodejs.org/en/download/ (Windows Installer)"
    Write-Host ""
    Write-Host "  Alternative (with Chocolatey):"
    Write-Host "    choco install nodejs"
    Write-Host ""
    Write-Host "  Alternative (with winget):"
    Write-Host "    winget install OpenJS.NodeJS"
    Write-Host ""
    Write-Host "  Alternative (with Scoop):"
    Write-Host "    scoop install nodejs"
}

function Test-Docker {
    Write-Header "Docker Engine"
    
    if (Test-Command "docker") {
        $version = Get-DockerVersion
        if ($version) {
            Write-Success "Docker installed: $version"
            if (Test-Version $version "20.10.0") {
                Write-Success "Docker version meets requirements (≥20.10.0)"
            } else {
                Write-Warning "Docker version $version is below recommended 20.10.0"
            }
            
            # Test Docker daemon
            try {
                docker ps >$null 2>&1
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "Docker daemon is running"
                } else {
                    Write-Error "Docker daemon not running. Please start Docker Desktop."
                }
            } catch {
                Write-Error "Cannot connect to Docker daemon"
            }
        } else {
            Write-Error "Docker daemon not running"
        }
    } else {
        Write-Error "Docker not installed"
        Show-DockerInstall
    }
    Write-Host ""
}

function Test-Kubectl {
    Write-Header "kubectl (Kubernetes CLI)"
    
    if (Test-Command "kubectl") {
        $version = Get-KubectlVersion
        if ($version) {
            Write-Success "kubectl installed: $version"
            if (Test-Version $version "1.21.0") {
                Write-Success "kubectl version meets requirements (≥1.21.0)"
            } else {
                Write-Warning "kubectl version $version is below required 1.21.0"
            }
        } else {
            Write-Warning "Could not determine kubectl version"
        }
    } else {
        Write-Error "kubectl not installed"
        Show-KubectlInstall
    }
    Write-Host ""
}

function Test-Kind {
    Write-Header "KIND (Kubernetes in Docker)"
    
    if (Test-Command "kind") {
        $version = Get-KindVersion
        if ($version) {
            Write-Success "KIND installed: $version"
            if (Test-Version $version "0.17.0") {
                Write-Success "KIND version meets requirements (≥0.17.0)"
            } else {
                Write-Warning "KIND version $version is below recommended 0.17.0"
            }
        } else {
            Write-Warning "Could not determine KIND version"
        }
    } else {
        Write-Error "KIND not installed"
        Show-KindInstall
    }
    Write-Host ""
}

function Test-Helm {
    Write-Header "Helm (Package Manager for Kubernetes)"
    
    if (Test-Command "helm") {
        $version = Get-HelmVersion
        if ($version) {
            Write-Success "Helm installed: $version"
            if (Test-Version $version "3.8.0") {
                Write-Success "Helm version meets requirements (≥3.8.0)"
            } else {
                Write-Warning "Helm version $version is below recommended 3.8.0"
            }
        } else {
            Write-Warning "Could not determine Helm version"
        }
    } else {
        Write-Error "Helm not installed"
        Show-HelmInstall
    }
    Write-Host ""
}

function Test-Go {
    Write-Header "Go Programming Language"
    
    if (Test-Command "go") {
        $version = Get-GoVersion
        if ($version) {
            Write-Success "Go installed: $version"
            if (Test-Version $version "1.17.0") {
                Write-Success "Go version meets requirements (≥1.17.0)"
            } else {
                Write-Warning "Go version $version is below required 1.17.0"
            }
        } else {
            Write-Warning "Could not determine Go version"
        }
    } else {
        Write-Error "Go not installed"
        Show-GoInstall
    }
    Write-Host ""
}

function Test-Node {
    Write-Header "Node.js Runtime"
    
    if (Test-Command "node") {
        $version = Get-NodeVersion
        if ($version) {
            Write-Success "Node.js installed: $version"
            if (Test-Version $version "16.14.0") {
                Write-Success "Node.js version meets requirements (≥16.14.0)"
            } else {
                Write-Warning "Node.js version $version is below required 16.14.0"
            }
        } else {
            Write-Warning "Could not determine Node.js version"
        }
    } else {
        Write-Error "Node.js not installed"
        Show-NodeInstall
    }
    Write-Host ""
}

function Test-Npm {
    Write-Header "npm (Node Package Manager)"
    
    if (Test-Command "npm") {
        $version = Get-NpmVersion
        if ($version) {
            Write-Success "npm installed: $version"
            if (Test-Version $version "8.5.0") {
                Write-Success "npm version meets requirements (≥8.5.0)"
            } else {
                Write-Warning "npm version $version is below recommended 8.5.0"
                Write-Info "🔄 Update npm: npm install -g npm@latest"
            }
        } else {
            Write-Warning "Could not determine npm version"
        }
    } else {
        Write-Error "npm not installed (usually comes with Node.js)"
        Write-Info "🔄 npm is typically installed with Node.js"
    }
    Write-Host ""
}

function Test-AdditionalTools {
    Write-Header "Additional Development Tools"
    
    # Check git
    if (Test-Command "git") {
        try {
            $version = (git --version) -replace '.*?(\d+\.\d+\.\d+).*','$1'
            Write-Success "git installed: $version"
        } catch {
            Write-Success "git available"
        }
    } else {
        Write-Warning "git not found (required for development)"
        Write-Info "  Install from: https://git-scm.com/download/win"
    }
    
    # Check make (usually with Git for Windows or Visual Studio)
    if (Test-Command "make") {
        Write-Success "make available"
    } else {
        Write-Warning "make not found (required for building)"
        Write-Info "  Install with: choco install make"
        Write-Info "  Or install Visual Studio Build Tools"
    }
    
    # Check curl (usually available on Windows 10+)
    if (Test-Command "curl") {
        Write-Success "curl available"
    } else {
        Write-Warning "curl not found (recommended for downloads)"
    }
    
    Write-Host ""
}

function Test-SystemResources {
    Write-Header "System Resources"
    
    try {
        # Check memory
        $memory = Get-CimInstance -ClassName Win32_ComputerSystem
        $memoryGB = [math]::Round($memory.TotalPhysicalMemory / 1GB, 1)
        
        if ($memoryGB -ge 8) {
            Write-Success "Memory: ${memoryGB}GB (sufficient)"
        } elseif ($memoryGB -ge 4) {
            Write-Warning "Memory: ${memoryGB}GB (minimum, 8GB+ recommended)"
        } else {
            Write-Error "Memory: ${memoryGB}GB (insufficient, minimum 4GB required)"
        }
        
        # Check CPU cores
        $cpu = Get-CimInstance -ClassName Win32_Processor
        $cores = $cpu.NumberOfCores
        
        if ($cores -ge 4) {
            Write-Success "CPU cores: $cores (sufficient)"
        } elseif ($cores -ge 2) {
            Write-Warning "CPU cores: $cores (minimum, 4+ recommended)"
        } else {
            Write-Error "CPU cores: $cores (insufficient, minimum 2 required)"
        }
        
        # Check disk space
        $disk = Get-CimInstance -ClassName Win32_LogicalDisk | Where-Object { $_.DriveType -eq 3 -and $_.DeviceID -eq "C:" }
        $freeSpaceGB = [math]::Round($disk.FreeSpace / 1GB, 1)
        
        if ($freeSpaceGB -ge 50) {
            Write-Success "Available disk space: ${freeSpaceGB}GB (sufficient)"
        } elseif ($freeSpaceGB -ge 20) {
            Write-Warning "Available disk space: ${freeSpaceGB}GB (minimum, 50GB+ recommended)"
        } else {
            Write-Error "Available disk space: ${freeSpaceGB}GB (insufficient, minimum 20GB required)"
        }
        
    } catch {
        Write-Info "System resource check not available: $($_.Exception.Message)"
    }
    
    Write-Host ""
}

function Start-Check {
    Write-Info "🔍 O-RAN Near-RT RIC Platform Prerequisites Check for Windows"
    Write-Host ""
    
    # Check all prerequisites
    Test-Docker
    Test-Kubectl
    Test-Kind
    Test-Helm
    Test-Go
    Test-Node
    Test-Npm
    Test-AdditionalTools
    Test-SystemResources
    
    # Summary
    Write-Header "Summary"
    Write-Info "Prerequisites check completed!"
    Write-Host ""
    Write-Info "📚 Next Steps:"
    Write-Host "  1. Install any missing dependencies listed above"
    Write-Host "  2. Restart your PowerShell session to reload PATH"
    Write-Host "  3. Run '.\scripts\setup.ps1' to deploy the platform"
    Write-Host ""
    Write-Info "📖 Resources:"
    Write-Host "  - O-RAN Documentation: docs\README.md"
    Write-Host "  - Developer Guide: dashboard-master\dashboard-master\docs\developer\getting-started.md"
    Write-Host "  - Troubleshooting: docs\operations\troubleshooting.md"
}

# Run the check
Start-Check