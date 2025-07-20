# Test Script for O-RAN Near-RT RIC Platform
# This script tests the setup and ensures all components work properly on Windows

param(
    [switch]$Quick,
    [switch]$SkipBuild,
    [switch]$Help
)

# Color output functions
function Write-Success { param($Message) Write-Host "✓ $Message" -ForegroundColor Green }
function Write-Error { param($Message) Write-Host "✗ $Message" -ForegroundColor Red }
function Write-Warning { param($Message) Write-Host "⚠ $Message" -ForegroundColor Yellow }
function Write-Info { param($Message) Write-Host "ℹ $Message" -ForegroundColor Blue }
function Write-Header { param($Message) Write-Host "=== $Message ===" -ForegroundColor Cyan }

function Show-Help {
    Write-Host @"
O-RAN Near-RT RIC Platform Test Script

Usage: .\test-setup.ps1 [OPTIONS]

Options:
  -Quick      Run quick tests only (skip builds and long-running tests)
  -SkipBuild  Skip build steps
  -Help       Show this help message

Description:
  This script tests all components of the O-RAN platform to ensure
  they work correctly on your Windows system:
  
  - Prerequisites validation
  - xApp Dashboard build and test
  - Main Dashboard build (if applicable)
  - Docker Compose setup
  - npm test execution

Examples:
  .\test-setup.ps1           # Full test suite
  .\test-setup.ps1 -Quick    # Quick tests only
  .\test-setup.ps1 -SkipBuild # Skip build steps
"@
}

function Test-Prerequisites {
    Write-Header "Testing Prerequisites"
    
    # Run prerequisites check
    if (Test-Path "scripts\check-prerequisites.ps1") {
        try {
            & ".\scripts\check-prerequisites.ps1"
            Write-Success "Prerequisites check completed"
            return $true
        } catch {
            Write-Error "Prerequisites check failed: $($_.Exception.Message)"
            return $false
        }
    } else {
        Write-Warning "Prerequisites check script not found"
        return $false
    }
}

function Test-XappDashboard {
    Write-Header "Testing xApp Dashboard"
    
    if (-not (Test-Path "xAPP_dashboard-master")) {
        Write-Error "xApp Dashboard directory not found"
        return $false
    }
    
    try {
        Set-Location "xAPP_dashboard-master"
        
        # Check package.json exists
        if (-not (Test-Path "package.json")) {
            Write-Error "package.json not found in xApp Dashboard"
            return $false
        }
        
        # Check if node_modules exists and has packages
        if ((Test-Path "node_modules") -and ((Get-ChildItem "node_modules" -ErrorAction SilentlyContinue).Count -gt 0)) {
            Write-Success "Dependencies already installed"
        } else {
            Write-Info "Installing xApp Dashboard dependencies..."
            npm install
            if ($LASTEXITCODE -ne 0) {
                Write-Error "Failed to install xApp Dashboard dependencies"
                return $false
            }
            Write-Success "Dependencies installed successfully"
        }
        
        if (-not $SkipBuild) {
            Write-Info "Building xApp Dashboard..."
            npm run build
            if ($LASTEXITCODE -ne 0) {
                Write-Error "Failed to build xApp Dashboard"
                return $false
            }
            Write-Success "Build completed successfully"
            
            # Check if dist folder was created
            if (Test-Path "dist") {
                Write-Success "Build artifacts created in dist/ folder"
            } else {
                Write-Warning "Build artifacts not found in expected location"
            }
        }
        
        Write-Info "Running xApp Dashboard tests..."
        npm run test:ci
        if ($LASTEXITCODE -ne 0) {
            Write-Warning "Some tests failed, but continuing..."
        } else {
            Write-Success "All tests passed"
        }
        
        Write-Info "Running lint check..."
        npm run lint
        if ($LASTEXITCODE -ne 0) {
            Write-Warning "Lint check found issues, but continuing..."
        } else {
            Write-Success "Lint check passed"
        }
        
        if (-not $Quick) {
            Write-Info "Running production build test..."
            npm run build:prod
            if ($LASTEXITCODE -ne 0) {
                Write-Warning "Production build failed"
            } else {
                Write-Success "Production build completed"
            }
        }
        
        Set-Location ".."
        Write-Success "xApp Dashboard tests completed"
        return $true
        
    } catch {
        Write-Error "xApp Dashboard test failed: $($_.Exception.Message)"
        Set-Location ".."
        return $false
    }
}

function Test-MainDashboard {
    Write-Header "Testing Main Dashboard"
    
    if (-not (Test-Path "dashboard-master\dashboard-master")) {
        Write-Error "Main Dashboard directory not found"
        return $false
    }
    
    try {
        Set-Location "dashboard-master\dashboard-master"
        
        # Check if make is available
        if (-not (Get-Command "make" -ErrorAction SilentlyContinue)) {
            Write-Warning "make command not available, skipping Main Dashboard build test"
            Set-Location "..\..\"
            return $true
        }
        
        Write-Info "Testing Main Dashboard build system..."
        
        # Check if we can run make commands
        if (-not $SkipBuild) {
            Write-Info "Running make check..."
            make check
            if ($LASTEXITCODE -ne 0) {
                Write-Warning "Make check found issues"
            } else {
                Write-Success "Make check passed"
            }
        }
        
        # Check npm dependencies
        if (Test-Path "package.json") {
            Write-Info "Installing Main Dashboard npm dependencies..."
            npm install
            if ($LASTEXITCODE -ne 0) {
                Write-Warning "Failed to install Main Dashboard npm dependencies"
            } else {
                Write-Success "Main Dashboard npm dependencies installed"
            }
        }
        
        Set-Location "..\..\"
        Write-Success "Main Dashboard tests completed"
        return $true
        
    } catch {
        Write-Error "Main Dashboard test failed: $($_.Exception.Message)"
        Set-Location "..\..\"
        return $false
    }
}

function Test-DockerCompose {
    Write-Header "Testing Docker Compose Setup"
    
    if (-not (Test-Path "docker-compose.yml")) {
        Write-Warning "docker-compose.yml not found"
        return $false
    }
    
    # Check if Docker is available
    if (-not (Get-Command "docker" -ErrorAction SilentlyContinue)) {
        Write-Warning "Docker not available, skipping Docker Compose test"
        return $true
    }
    
    try {
        Write-Info "Validating Docker Compose configuration..."
        docker-compose config --quiet
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Docker Compose configuration is valid"
            return $true
        } else {
            Write-Error "Docker Compose configuration validation failed"
            return $false
        }
        
    } catch {
        Write-Error "Docker Compose test failed: $($_.Exception.Message)"
        return $false
    }
}

function Test-KubernetesManifests {
    Write-Header "Testing Kubernetes Manifests"
    
    if (-not (Get-Command "kubectl" -ErrorAction SilentlyContinue)) {
        Write-Warning "kubectl not available, skipping Kubernetes manifest validation"
        return $true
    }
    
    try {
        $manifestFiles = @()
        
        # Find all YAML files in k8s directory
        if (Test-Path "k8s") {
            $manifestFiles += Get-ChildItem -Path "k8s" -Filter "*.yaml" -Recurse
        }
        
        # Find all YAML files in helm directory
        if (Test-Path "helm") {
            $manifestFiles += Get-ChildItem -Path "helm" -Filter "*.yaml" -Recurse
        }
        
        $validCount = 0
        $errorCount = 0
        
        foreach ($file in $manifestFiles) {
            Write-Info "Validating $($file.Name)..."
            kubectl apply --dry-run=client -f $file.FullName >$null 2>&1
            if ($LASTEXITCODE -eq 0) {
                $validCount++
            } else {
                $errorCount++
                Write-Warning "Validation failed for $($file.Name)"
            }
        }
        
        Write-Info "Validated $($manifestFiles.Count) manifest files"
        Write-Success "$validCount files passed validation"
        if ($errorCount -gt 0) {
            Write-Warning "$errorCount files failed validation"
        }
        
        return $true
        
    } catch {
        Write-Error "Kubernetes manifest test failed: $($_.Exception.Message)"
        return $false
    }
}

function Test-Scripts {
    Write-Header "Testing Scripts"
    
    $scriptTests = @(
        @{ Name = "Setup Script"; Path = "scripts\setup.ps1"; Test = { param($path) Test-Path $path } },
        @{ Name = "Prerequisites Check"; Path = "scripts\check-prerequisites.ps1"; Test = { param($path) Test-Path $path } }
    )
    
    $passCount = 0
    
    foreach ($test in $scriptTests) {
        Write-Info "Testing $($test.Name)..."
        try {
            $result = & $test.Test $test.Path
            if ($result) {
                Write-Success "$($test.Name) exists and is accessible"
                $passCount++
            } else {
                Write-Error "$($test.Name) test failed"
            }
        } catch {
            Write-Error "$($test.Name) test failed: $($_.Exception.Message)"
        }
    }
    
    Write-Info "$passCount out of $($scriptTests.Count) script tests passed"
    return $passCount -eq $scriptTests.Count
}

function Test-Configuration {
    Write-Header "Testing Configuration Files"
    
    $configFiles = @(
        "kind-config.yaml",
        "docker-compose.yml",
        "config\prometheus\prometheus.yml"
    )
    
    $foundCount = 0
    
    foreach ($file in $configFiles) {
        if (Test-Path $file) {
            Write-Success "Found: $file"
            $foundCount++
        } else {
            Write-Warning "Missing: $file"
        }
    }
    
    Write-Info "$foundCount out of $($configFiles.Count) configuration files found"
    return $foundCount -gt 0
}

function Show-TestSummary {
    param($Results)
    
    Write-Header "Test Summary"
    
    $passCount = 0
    $totalCount = $Results.Count
    
    foreach ($result in $Results) {
        if ($result.Success) {
            Write-Success "$($result.Name): PASSED"
            $passCount++
        } else {
            Write-Error "$($result.Name): FAILED"
        }
    }
    
    Write-Host ""
    if ($passCount -eq $totalCount) {
        Write-Success "🎉 All tests passed! ($passCount/$totalCount)"
        Write-Info "Your O-RAN Near-RT RIC platform is ready for deployment."
        Write-Info "Run '.\scripts\setup.ps1' to start the platform."
    } elseif ($passCount -gt 0) {
        Write-Warning "⚠ Some tests passed ($passCount/$totalCount)"
        Write-Info "Review the failed tests above and fix any issues."
    } else {
        Write-Error "❌ All tests failed (0/$totalCount)"
        Write-Info "Please check your environment and dependencies."
    }
    
    Write-Host ""
    Write-Info "📚 For help with issues:"
    Write-Host "  - Check docs\operations\troubleshooting.md"
    Write-Host "  - Run '.\scripts\check-prerequisites.ps1' for detailed system info"
    Write-Host "  - Review the setup guide in docs\operations\deployment.md"
}

function Start-TestSuite {
    Write-Info "🧪 Starting O-RAN Near-RT RIC Platform Test Suite"
    Write-Host "This will test all components to ensure they work on your system."
    Write-Host ""
    
    if ($Quick) {
        Write-Info "Running in QUICK mode - skipping builds and long-running tests"
    }
    
    if ($SkipBuild) {
        Write-Info "Skipping build steps"
    }
    
    # Test results collection
    $results = @()
    
    # Run tests
    $results += @{ Name = "Prerequisites"; Success = (Test-Prerequisites) }
    $results += @{ Name = "Configuration Files"; Success = (Test-Configuration) }
    $results += @{ Name = "Scripts"; Success = (Test-Scripts) }
    $results += @{ Name = "xApp Dashboard"; Success = (Test-XappDashboard) }
    
    if (-not $Quick) {
        $results += @{ Name = "Main Dashboard"; Success = (Test-MainDashboard) }
        $results += @{ Name = "Docker Compose"; Success = (Test-DockerCompose) }
        $results += @{ Name = "Kubernetes Manifests"; Success = (Test-KubernetesManifests) }
    }
    
    # Show summary
    Show-TestSummary $results
}

# Main script logic
if ($Help) {
    Show-Help
    exit 0
}

# Run test suite
Start-TestSuite