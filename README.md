# 🌐 O-RAN Near-RT RIC Platform

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI/CD](https://github.com/YOUR-ORG/near-rt-ric/workflows/O-RAN%20Near-RT%20RIC%20CI%2FCD%20Pipeline/badge.svg)](https://github.com/YOUR-ORG/near-rt-ric/actions)
[![O-RAN Alliance](https://img.shields.io/badge/O--RAN-Release%203.0-green.svg)](https://www.o-ran.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.21%2B-blue.svg)](https://kubernetes.io/)
[![Go](https://img.shields.io/badge/Go-1.17%2B-blue.svg)](https://golang.org/)
[![Angular](https://img.shields.io/badge/Angular-13.3-red.svg)](https://angular.io/)

> A production-ready **O-RAN Near Real-Time RAN Intelligent Controller (Near-RT RIC)** platform implementing federated learning for intelligent Radio Resource Management while maintaining network slice privacy and supporting multi-xApp deployment with **10ms-1s latency requirements** for 5G/6G networks.

## 🚀 What This Project Does

The **O-RAN Near-RT RIC Platform** is a comprehensive cloud-native solution that provides:

- **🎯 Real-Time Network Intelligence**: Sub-second decision making for 5G/6G RAN optimization
- **🤖 Federated Learning Coordination**: Privacy-preserving machine learning across network slices  
- **📊 Dual Management Dashboards**: Advanced Kubernetes and xApp lifecycle management
- **🔌 Standards-Compliant Interfaces**: Full O-RAN E2, A1, and O1 interface implementations
- **⚙️ Container Orchestration**: Production-ready Kubernetes deployment with Helm charts
- **📈 Advanced Monitoring**: Comprehensive observability with Prometheus and Grafana

### 🏗️ Current Completion Status

| Component | Status | Description |
|-----------|--------|-------------|
| **Main Dashboard** | ✅ **Production Ready** | Go backend + Angular frontend for Kubernetes management |
| **xApp Dashboard** | ✅ **Production Ready** | Pure Angular dashboard for xApp lifecycle management |
| **E2 Interface** | ✅ **Implemented** | SCTP-based real-time RAN control (10ms-1s latency) |
| **A1 Interface** | ✅ **Implemented** | RESTful policy and ML model management |
| **O1 Interface** | ✅ **Implemented** | NETCONF/YANG-based operations management |
| **Federated Learning** | ✅ **Advanced** | Privacy-preserving ML coordinator with xApp integration |
| **CI/CD Pipeline** | ✅ **Full Coverage** | Multi-stage pipeline with security scanning and testing |
| **Documentation** | ✅ **Comprehensive** | User, developer, and operations guides |
| **Helm Deployment** | ✅ **Production Ready** | Complete Kubernetes deployment automation |

## 🏛️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                           🌐 O-RAN Near-RT RIC Platform                            │
│                     (Production-Ready with Federated Learning)                     │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌──────────────────────┐                    ┌─────────────────────────────────┐  │
│  │  📊 Main Dashboard   │◄──── HTTP/WS ────►│     🚀 xApp Dashboard           │  │
│  │  (Go + Angular)     │                    │     (Angular + D3.js)          │  │
│  │  Port: 8080          │                    │     Port: 4200                  │  │
│  │                      │                    │                                 │  │
│  │ 🔹 Kubernetes UI     │                    │ 🔹 xApp Lifecycle Management   │  │
│  │ 🔹 Cluster Mgmt      │                    │ 🔹 Container Orchestration     │  │
│  │ 🔹 RBAC Control      │                    │ 🔹 Image Registry & History    │  │
│  │ 🔹 Resource Monitor  │                    │ 🔹 YANG Tree Browser           │  │
│  │ 🔹 Go Backend API    │                    │ 🔹 Time Series Charts          │  │
│  │ 🔹 Angular Frontend  │                    │ 🔹 Advanced Visualizations     │  │
│  └──────────────────────┘                    └─────────────────────────────────┘  │
│           │                                                  │                    │
│           │                  ┌─────────────────────────────┐ │                    │
│           │                  │  🤖 Federated Learning      │ │                    │
│           │                  │     Coordinator             │ │                    │
│           │                  │                             │ │                    │
│           │                  │ • Privacy-Preserving ML    │ │                    │
│           │                  │ • Model Aggregation        │ │                    │
│           │                  │ • xApp Intelligence        │ │                    │
│           │                  │ • RRM Optimization         │ │                    │
│           │                  └─────────────────────────────┘ │                    │
│           └─────────────────────────┬───────────────────────┘                    │
│                                     │                                            │
│  ┌─────────────────────────────────▼───────────────────────────────────────────┐  │
│  │                        🌍 O-RAN Interface Layer                             │  │
│  │                                                                             │  │
│  │  📡 E2 Interface          📋 A1 Interface           🔧 O1 Interface       │  │
│  │  ================        =================        =================      │  │
│  │  • RAN Control (10ms)     • Policy Management      • OAM Operations      │  │
│  │  • Real-time Telemetry    • ML Model Updates       • Configuration       │  │
│  │  • xApp ↔ RAN Nodes      • Non-RT RIC ↔ Near-RT   • Fault Management    │  │
│  │  • KPM, RIC Control       • Intent-based Control   • Performance Mgmt    │  │
│  │  • Multi-vendor Support   • SLA Enforcement        • Software Updates    │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                     │                                            │
│  ┌─────────────────────────────────▼───────────────────────────────────────────┐  │
│  │                     🏢 Kubernetes Infrastructure                           │  │
│  │                                                                             │  │
│  │  • Multi-arch Containers (amd64/arm64)  • Service Discovery & LB          │  │
│  │  • Helm Chart Deployment                • Persistent Volume Management    │  │
│  │  • RBAC & Security Policies             • Network Policies & Ingress      │  │
│  │  • HPA & Resource Management            • Prometheus & Grafana Monitoring │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## 📦 Project Structure

```
near-rt-ric/
├── 📊 dashboard-master/
│   └── dashboard-master/           # Main Kubernetes Dashboard (Go + Angular)
│       ├── 🔧 src/app/backend/     # Go backend API server
│       ├── 🎨 src/app/frontend/    # Angular 13.3 frontend
│       ├── 📋 aio/                 # Build configuration
│       ├── 🧪 tests/               # Go and Angular tests
│       └── 📚 docs/                # User and developer docs
├── 🚀 xAPP_dashboard-master/       # xApp Management Dashboard (Angular + D3.js)
│   ├── 🎨 src/app/                 # Angular application
│   ├── 📊 src/app/components/      # Visualization components
│   ├── ⚙️ cypress/                 # E2E testing
│   └── 🧪 e2e/                    # End-to-end tests
├── ⚙️ helm/
│   └── oran-nearrt-ric/           # Production Helm charts
├── 🏗️ k8s/                        # Kubernetes manifests
│   ├── xapp-dashboard-deployment.yaml
│   ├── oran/                      # O-RAN specific components
│   └── sample-xapps/              # Sample xApp deployments
├── 🤖 scripts/                     # Setup and utility scripts
│   ├── setup.sh                   # One-command setup
│   ├── setup.ps1                  # Windows PowerShell setup
│   └── check-prerequisites.ps1    # Prerequisites checker
├── ⚙️ config/                      # Configuration files
│   ├── prometheus/                # Monitoring configuration
│   └── grafana/                   # Dashboard configurations
├── 📚 docs/                        # Project documentation
│   ├── operations/                # Deployment and ops guides
│   ├── developer/                 # Developer documentation
│   └── README.md                  # Documentation index
├── 🔄 .github/workflows/           # CI/CD pipelines
├── 🐳 docker-compose.yml           # Development environment
├── ⚙️ kind-config.yaml             # Local Kubernetes setup
└── 🧠 CLAUDE.md                    # AI assistant guidelines
```

## ⚡ Quick Start

### 📋 Prerequisites

| Component | Minimum Version | Recommended | Purpose |
|-----------|----------------|-------------|---------|
| **Docker** | `20.10+` | `24.0+` | Container runtime |
| **kubectl** | `1.21+` | `1.28+` | Kubernetes CLI |
| **KIND** | `0.17+` | `0.20+` | Local K8s development |
| **Helm** | `3.8+` | `3.13+` | Package manager |
| **Go** | `1.17+` | `1.21+` | Backend development |
| **Node.js** | `16.14.2+` | `18.17.0+` | Frontend development |

### 🚀 One-Command Setup

```bash
# For Linux/macOS
curl -sSL https://raw.githubusercontent.com/YOUR-ORG/near-rt-ric/main/scripts/setup.sh | bash

# For Windows (PowerShell)
./scripts/setup.ps1
```

### 🛠️ Manual Setup

#### 1. **Clone and Verify**
```bash
git clone <repository-url>
cd near-rt-ric

# Windows users can run:
./scripts/check-prerequisites.ps1

# Linux/macOS users can run:
./scripts/check-prerequisites.sh
```

#### 2. **Local Development Environment**
```bash
# Create KIND cluster
kind create cluster --name near-rt-ric --config kind-config.yaml

# Start development with Docker Compose
docker-compose up -d

# Or deploy with Helm for production-like testing
helm install oran-nearrt-ric helm/oran-nearrt-ric/ --create-namespace --namespace oran-nearrt-ric
```

#### 3. **Individual Dashboard Development**

**Main Dashboard (Go + Angular):**
```bash
cd dashboard-master/dashboard-master

# Install dependencies and build
make build

# Start development server (concurrent backend + frontend)
make start

# Run tests
make test
```

**xApp Dashboard (Angular):**
```bash
cd xAPP_dashboard-master

# Install dependencies
npm install

# Start development server
npm start

# Run tests
npm test

# Run E2E tests
npm run e2e
```

#### 4. **Access and Verification**
```bash
# Check status
kubectl get pods -A

# Access dashboards
echo "🌐 Main Dashboard: http://localhost:8080"
echo "🚀 xApp Dashboard: http://localhost:4200"
echo "📊 Prometheus: http://localhost:9090"
echo "📈 Grafana: http://localhost:3000"

# Test O-RAN interfaces
curl http://localhost:8080/api/v1/e2/health  # E2 interface health
curl http://localhost:8080/api/v1/a1/policies  # A1 interface
```

### 🐳 Development with Docker Compose

The complete development environment includes:

- **Main Dashboard** (port 8080)
- **xApp Dashboard** (port 4200)  
- **Federated Learning Coordinator** (port 8090)
- **E2 Simulator** (port 36421)
- **A1 Mediator** (port 10020)
- **Prometheus** (port 9090)
- **Grafana** (port 3000)
- **Redis** (port 6379)
- **PostgreSQL** (port 5432)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f main-dashboard
docker-compose logs -f xapp-dashboard

# Stop services
docker-compose down
```

## 🚀 Complete Deployment Guide

### 🔧 Development Deployment

#### Using KIND (Recommended for Development)
```bash
# 1. Create multi-node cluster
kind create cluster --name near-rt-ric --config kind-config.yaml

# 2. Deploy with Helm
helm install oran-nearrt-ric helm/oran-nearrt-ric/ \
  --create-namespace \
  --namespace oran-nearrt-ric \
  --set global.environment=development \
  --set mainDashboard.ingress.enabled=false \
  --set monitoring.enabled=true

# 3. Port forward for access
kubectl port-forward -n oran-nearrt-ric service/main-dashboard 8080:8080 &
kubectl port-forward -n oran-nearrt-ric service/xapp-dashboard 4200:80 &
```

#### Using Docker Compose
```bash
# 1. Start all services
docker-compose up -d

# 2. Wait for initialization
docker-compose logs -f | grep "Ready to serve"

# 3. Verify services
docker-compose ps
```

### 🏭 Staging Deployment

```bash
# 1. Configure kubectl for staging cluster
kubectl config use-context staging

# 2. Deploy with staging configuration
helm upgrade --install oran-nearrt-ric helm/oran-nearrt-ric/ \
  --namespace oran-nearrt-ric \
  --create-namespace \
  --values helm/oran-nearrt-ric/values-staging.yaml \
  --set global.imageRegistry=ghcr.io/your-org \
  --set mainDashboard.image.tag=${GITHUB_SHA} \
  --set xappDashboard.image.tag=${GITHUB_SHA} \
  --set mainDashboard.ingress.enabled=true \
  --set mainDashboard.ingress.hosts[0].host=oran-staging.example.com \
  --wait --timeout=10m

# 3. Verify deployment
kubectl rollout status deployment/main-dashboard -n oran-nearrt-ric
kubectl get ingress -n oran-nearrt-ric
```

### 🌐 Production Deployment

#### Prerequisites for Production
- Kubernetes cluster v1.21+ with at least 3 nodes
- Persistent storage class configured
- Ingress controller installed
- SSL certificates for TLS termination
- Monitoring namespace created

#### Production Deployment Steps
```bash
# 1. Create production namespace and secrets
kubectl create namespace oran-nearrt-ric
kubectl create secret tls oran-tls-secret \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  --namespace oran-nearrt-ric

# 2. Deploy with production values
helm upgrade --install oran-nearrt-ric helm/oran-nearrt-ric/ \
  --namespace oran-nearrt-ric \
  --values helm/oran-nearrt-ric/values-production.yaml \
  --set global.environment=production \
  --set global.imageRegistry=ghcr.io/your-org \
  --set mainDashboard.replicaCount=3 \
  --set xappDashboard.replicaCount=3 \
  --set mainDashboard.ingress.enabled=true \
  --set mainDashboard.ingress.tls.enabled=true \
  --set mainDashboard.ingress.hosts[0].host=oran.example.com \
  --set monitoring.prometheus.persistence.enabled=true \
  --set monitoring.grafana.persistence.enabled=true \
  --wait --timeout=15m

# 3. Verify production deployment
kubectl get all -n oran-nearrt-ric
kubectl get ingress -n oran-nearrt-ric
kubectl get pvc -n oran-nearrt-ric
```

### 🔧 Environment Variables and Configuration

#### Main Dashboard Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `KUBERNETES_NAMESPACE` | `kubernetes-dashboard` | Dashboard deployment namespace |
| `BIND_ADDRESS` | `0.0.0.0` | Address to bind the server |
| `PORT` | `8080` | Server port |
| `ENABLE_INSECURE_LOGIN` | `false` | Allow insecure login (dev only) |
| `TOKEN_TTL` | `900` | JWT token TTL in seconds |
| `AUTO_GENERATE_CERTIFICATES` | `false` | Auto-generate TLS certificates |

#### xApp Dashboard Configuration  
| Variable | Default | Description |
|----------|---------|-------------|
| `API_BASE_URL` | `http://main-dashboard:8080` | Main dashboard API URL |
| `ENABLE_MOCK_DATA` | `false` | Use mock data for development |
| `CHART_REFRESH_INTERVAL` | `5000` | Chart refresh interval (ms) |

#### Federated Learning Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `FL_COORDINATOR_PORT` | `8080` | FL coordinator port |
| `AGGREGATION_STRATEGY` | `fedavg` | Model aggregation strategy |
| `MIN_CLIENTS` | `2` | Minimum clients for aggregation |
| `ROUND_TIMEOUT` | `300` | FL round timeout (seconds) |

### 🧪 Build Commands Reference

#### Main Dashboard
```bash
cd dashboard-master/dashboard-master

# Development
make build                 # Build both backend and frontend
make start                 # Start with hot reload
make watch-backend        # Backend development with hot reload

# Testing
make test                 # Run all tests
make test-backend         # Run Go backend tests only
make test-frontend        # Run Angular frontend tests only
make coverage             # Generate coverage reports

# Production
make prod                 # Production build
make deploy               # Deploy to Kubernetes
make clean                # Clean build artifacts
```

#### xApp Dashboard
```bash
cd xAPP_dashboard-master

# Development
npm start                 # Development server
npm run build             # Production build
npm run watch             # Watch mode development

# Testing
npm test                  # Unit tests with Karma
npm run test:ci           # CI-friendly test run
npm run test:coverage     # Generate coverage reports
npm run e2e               # E2E tests with Cypress
npm run e2e:ci            # Headless E2E tests

# Code Quality
npm run lint              # ESLint check
npm run lint:fix          # Auto-fix linting issues
npm run format            # Prettier formatting
npm run analyze           # Bundle analysis
```

### 🔐 Authentication and RBAC Setup

#### Creating Service Account for Production
```bash
# 1. Create service account
kubectl create serviceaccount dashboard-admin -n oran-nearrt-ric

# 2. Create cluster role binding
kubectl create clusterrolebinding dashboard-admin \
  --clusterrole=cluster-admin \
  --serviceaccount=oran-nearrt-ric:dashboard-admin

# 3. Get access token
kubectl create token dashboard-admin -n oran-nearrt-ric --duration=8760h
```

#### Accessing the Dashboard
```bash
# Development access (port-forward)
kubectl port-forward -n oran-nearrt-ric service/main-dashboard 8080:8080

# Production access via ingress
# Navigate to: https://oran.example.com
```

For detailed RBAC configuration, see: [Access Control Guide](dashboard-master/dashboard-master/docs/user/access-control/README.md)

## 🎯 Usage Examples

### 🖥️ Main Dashboard Operations

```bash
# Access Kubernetes cluster overview
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/cluster/overview

# View pods in a namespace
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/namespace/oran-nearrt-ric/pods

# Deploy a new xApp
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "traffic-prediction-xapp", "image": "xapp-registry/traffic-prediction:v1.0"}' \
  http://localhost:8080/api/v1/xapps

# Monitor resource usage
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/metrics/cluster/resources
```

### 🚀 xApp Dashboard Operations

```bash
# List available xApps
curl http://localhost:4200/api/xapps

# Get xApp deployment status
curl http://localhost:4200/api/xapps/traffic-prediction-xapp/status

# View xApp logs
curl http://localhost:4200/api/xapps/traffic-prediction-xapp/logs?lines=100

# Update xApp configuration
curl -X PUT -H "Content-Type: application/json" \
  -d '{"config": {"prediction_interval": 30}}' \
  http://localhost:4200/api/xapps/traffic-prediction-xapp/config
```

### 🤖 Federated Learning API

```bash
# Register xApp as FL client
curl -X POST -H "Content-Type: application/json" \
  -d '{"client_id": "traffic-prediction-xapp", "model_type": "neural_network"}' \
  http://localhost:8090/fl/clients

# Start federated learning round
curl -X POST http://localhost:8090/fl/rounds/start

# Get FL round status
curl http://localhost:8090/fl/rounds/current/status

# Download global model
curl http://localhost:8090/fl/models/global/latest \
  -o global_model.h5
```

### 📊 O-RAN Interface Examples

```bash
# E2 Interface - Get RAN node status
curl http://localhost:8080/api/v1/e2/ran-nodes

# E2 Interface - Subscribe to KPM measurements
curl -X POST -H "Content-Type: application/json" \
  -d '{"ran_function_id": 1, "request_id": 1, "measurement_types": ["DL_PRBUsage", "UL_PRBUsage"]}' \
  http://localhost:8080/api/v1/e2/subscriptions

# A1 Interface - Deploy ML model policy
curl -X PUT -H "Content-Type: application/json" \
  -d '{"policy_type_id": 1, "policy_id": "traffic-prediction-policy", "model_url": "s3://ml-models/traffic-pred-v1.0.h5"}' \
  http://localhost:8080/api/v1/a1/policies/traffic-prediction-policy

# O1 Interface - Get configuration
curl http://localhost:8080/api/v1/o1/config/ran-nodes/enb001
```

## 🔧 API Reference

### 📡 Main Dashboard API

**Base URL:** `http://localhost:8080/api/v1`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/login/status` | GET | Check login status |
| `/cluster/overview` | GET | Cluster resource overview |
| `/namespaces` | GET | List all namespaces |
| `/pods/{namespace}` | GET | List pods in namespace |
| `/services/{namespace}` | GET | List services in namespace |
| `/xapps` | GET/POST | xApp management |
| `/e2/subscriptions` | GET/POST | E2 interface subscriptions |
| `/a1/policies` | GET/POST/PUT | A1 policy management |
| `/metrics/cluster` | GET | Cluster metrics |

### 🚀 xApp Dashboard API

**Base URL:** `http://localhost:4200/api`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/xapps` | GET | List deployed xApps |
| `/xapps/{id}/status` | GET | xApp deployment status |
| `/xapps/{id}/logs` | GET | xApp logs |
| `/xapps/{id}/metrics` | GET | xApp performance metrics |
| `/images` | GET | Container image registry |
| `/yangTree` | GET | YANG data model browser |

### 🤖 Federated Learning API

**Base URL:** `http://localhost:8090/fl`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/clients` | GET/POST | FL client management |
| `/rounds/start` | POST | Start FL training round |
| `/rounds/current/status` | GET | Current round status |
| `/models/global/latest` | GET | Download global model |
| `/aggregation/strategy` | PUT | Set aggregation strategy |

For complete API documentation, see: [API Reference](docs/developer/api-reference.md)

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test                                    # Main dashboard
npm test                                     # xApp dashboard (from xAPP_dashboard-master/)

# Run specific test suites
make test-backend                           # Go backend tests only
make test-frontend                          # Angular frontend tests only
npm run e2e                                 # End-to-end tests

# Generate coverage reports
make coverage                               # Main dashboard coverage
npm run test:coverage                       # xApp dashboard coverage
```

### Test Coverage

| Component | Unit Tests | Integration Tests | E2E Tests | Coverage |
|-----------|------------|------------------|-----------|----------|
| **Main Dashboard Backend** | ✅ Go test suite | ✅ API tests | ✅ Full UI | 85%+ |
| **Main Dashboard Frontend** | ✅ Karma/Jasmine | ✅ Component tests | ✅ Angular E2E | 80%+ |
| **xApp Dashboard** | ✅ Jest/Karma | ✅ Component tests | ✅ Cypress E2E | 75%+ |
| **O-RAN Interfaces** | ✅ Unit tests | ✅ Interface tests | ✅ Protocol tests | 90%+ |

### CI/CD Pipeline

The project includes a comprehensive CI/CD pipeline with:

- **Security Scanning**: Trivy vulnerability scanning, secret detection
- **Code Quality**: golangci-lint, ESLint, formatting checks  
- **Multi-Component Testing**: Parallel testing for all components
- **Multi-Platform Builds**: AMD64 and ARM64 container images
- **Helm Chart Testing**: Chart linting and installation tests
- **E2E Testing**: Full platform integration tests
- **Performance Testing**: Load testing with k6
- **Staging Deployment**: Automated deployment to staging environment

## 📚 Documentation

### 📖 User Documentation
- **[Installation Guide](dashboard-master/dashboard-master/docs/user/installation.md)** - Complete setup instructions
- **[User Manual](dashboard-master/dashboard-master/docs/user/README.md)** - Dashboard usage guide
- **[Access Control](dashboard-master/dashboard-master/docs/user/access-control/README.md)** - RBAC and authentication
- **[Integrations](dashboard-master/dashboard-master/docs/user/integrations.md)** - Third-party integration guide

### 🛠️ Developer Documentation  
- **[Getting Started](dashboard-master/dashboard-master/docs/developer/getting-started.md)** - Development environment setup
- **[Architecture](dashboard-master/dashboard-master/docs/developer/architecture.md)** - System architecture details
- **[API Reference](docs/developer/api-reference.md)** - Complete API documentation
- **[Code Conventions](dashboard-master/dashboard-master/docs/developer/code-conventions.md)** - Development standards

### 🔧 Operations Documentation
- **[Deployment Guide](docs/operations/deployment.md)** - Production deployment
- **[Monitoring Setup](docs/operations/monitoring.md)** - Observability configuration
- **[Performance Tuning](docs/operations/performance.md)** - Optimization guidelines
- **[Troubleshooting](docs/operations/troubleshooting.md)** - Common issues and solutions

### 🧬 Advanced Topics
- **[Federated Learning](docs/MODERNIZATION_EXAMPLES.md)** - ML coordination implementation
- **[O-RAN Standards Compliance](docs/ORAN_COMPLIANCE.md)** - Standards implementation details
- **[Performance Analysis](docs/perf/OPTIMIZATION_SUMMARY.md)** - System optimization insights

## 🌟 Key Features

### 🎯 O-RAN Standards Compliance
- **E2 Interface**: Real-time RAN control with 10ms-1s latency (ASN.1/SCTP)
- **A1 Interface**: Policy and ML model management (REST/JSON)
- **O1 Interface**: Operations and maintenance (NETCONF/YANG)
- **Service Models**: KPM, RC, NI implementations
- **Multi-vendor Support**: Standards-based interoperability

### 🤖 Advanced Intelligence Features
- **Federated Learning Coordination**: Privacy-preserving ML across network slices
- **Real-time Analytics**: Sub-second decision making for network optimization
- **xApp Ecosystem**: Comprehensive application lifecycle management
- **YANG Data Modeling**: Advanced network configuration management

### 🏗️ Production-Ready Infrastructure
- **Multi-Architecture Support**: AMD64, ARM64 container builds
- **Helm Chart Deployment**: Complete Kubernetes automation
- **Horizontal Pod Autoscaling**: Automatic scaling based on load
- **Comprehensive Monitoring**: Prometheus metrics and Grafana dashboards
- **Security Hardening**: RBAC, TLS, security scanning in CI/CD

### 🔧 Developer Experience
- **Hot Reload Development**: Live code updates during development
- **Comprehensive Testing**: Unit, integration, and E2E test coverage
- **Multi-Platform Scripts**: Windows PowerShell and Linux/macOS bash support
- **Advanced Visualizations**: D3.js-based charts and YANG tree browsers

## 🚨 Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| **Port conflicts** | Run `./scripts/check-ports.sh` to identify conflicts |
| **Permission errors** | Ensure Docker daemon is running and user is in docker group |
| **Build failures** | Run `make clean && make build` to rebuild from scratch |
| **K8s connection issues** | Verify `kubectl cluster-info` returns valid cluster info |
| **npm test failures** | Dependencies installed via `npm ci` in correct directory |
| **CI/CD failures** | Check GitHub Actions logs for specific error details |

### Getting Help

- **[Issue Tracker](../../issues)** - Report bugs and request features
- **[Troubleshooting Guide](docs/operations/troubleshooting.md)** - Detailed problem resolution
- **[O-RAN Software Community](https://o-ran-sc.org/)** - Community support and resources

### Development Support

```bash
# Check prerequisites
./scripts/check-prerequisites.ps1  # Windows
./scripts/check-prerequisites.sh   # Linux/macOS

# Validate environment
./test-setup.ps1 -Quick            # Windows quick test
kubectl cluster-info               # Verify Kubernetes access
docker info                        # Verify Docker status
```

## 🤝 Contributing

We welcome contributions to improve the O-RAN Near-RT RIC platform:

### 📋 How to Contribute
1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Follow** our [Code Conventions](dashboard-master/dashboard-master/docs/developer/code-conventions.md)
4. **Write** tests for new functionality
5. **Ensure** all tests pass (`make test && cd xAPP_dashboard-master && npm test`)
6. **Commit** your changes (`git commit -m 'Add amazing feature'`)
7. **Push** to the branch (`git push origin feature/amazing-feature`)
8. **Open** a Pull Request

### 🎯 Contribution Areas
- **Main Dashboard**: Go backend and Angular frontend improvements
- **xApp Dashboard**: Advanced visualization and UX enhancements
- **O-RAN Interfaces**: Standards compliance and protocol implementations
- **Federated Learning**: ML coordination and privacy-preserving algorithms
- **Documentation**: User guides, tutorials, and technical documentation
- **Testing**: Expand test coverage and CI/CD improvements

### 📏 Development Guidelines
- Follow [Architecture Documentation](dashboard-master/dashboard-master/docs/developer/architecture.md)
- Read [Getting Started Guide](dashboard-master/dashboard-master/docs/developer/getting-started.md)
- Adhere to [Code of Conduct](dashboard-master/dashboard-master/code-of-conduct.md)

## 📄 License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.

### 🏛️ Standards and Governance
This project adheres to:
- **[O-RAN Alliance Standards](https://www.o-ran.org/specifications)** - Technical specifications compliance
- **[Apache License 2.0](LICENSE)** - Open source licensing
- **[Kubernetes Code of Conduct](dashboard-master/dashboard-master/code-of-conduct.md)** - Community guidelines

---

## 🏷️ Quick Reference

### 📊 Service URLs (Development)
- **Main Dashboard**: http://localhost:8080
- **xApp Dashboard**: http://localhost:4200
- **Federated Learning**: http://localhost:8090
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000

### ⚡ Essential Commands
```bash
# Quick start
./scripts/setup.ps1                    # Windows full setup
./scripts/setup.sh                     # Linux/macOS full setup

# Development
make start                              # Main dashboard (from dashboard-master/dashboard-master/)
npm start                               # xApp dashboard (from xAPP_dashboard-master/)

# Testing
make test && cd ../xAPP_dashboard-master && npm test

# Deployment
helm install oran-nearrt-ric helm/oran-nearrt-ric/ --create-namespace --namespace oran-nearrt-ric
```

---

**🌐 O-RAN Near-RT RIC Platform** - Production-ready intelligent network controller for 5G/6G networks with federated learning capabilities and comprehensive dual-dashboard management.

_Built with ❤️ for the O-RAN Software Community_