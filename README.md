# O-RAN Near Real-Time RAN Intelligent Controller (Near-RT RIC)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI/CD Pipeline](https://github.com/hctsai1006/near-rt-ric/workflows/O-RAN%20Near-RT%20RIC%20CI%2FCD%20Pipeline/badge.svg)](https://github.com/hctsai1006/near-rt-ric/actions)
[![O-RAN Alliance](https://img.shields.io/badge/O--RAN-Release%203.0-green.svg)](https://www.o-ran.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.21%2B-blue.svg)](https://kubernetes.io/)

> **Production-ready O-RAN Near-RT RIC platform** implementing federated learning for intelligent Radio Resource Management with dual Angular dashboards, comprehensive Kubernetes orchestration, and full O-RAN standards compliance for 5G/6G networks.

## 🎯 Project Overview

This is a comprehensive **O-RAN Near Real-Time RAN Intelligent Controller (Near-RT RIC)** platform that provides intelligent network management and optimization capabilities for 5G/6G radio access networks. The platform features:

- **🎯 Real-Time Intelligence**: Sub-second decision making (10ms-1s latency) for RAN optimization
- **🤖 Federated Learning**: Privacy-preserving ML coordination across network slices
- **📊 Dual Management Dashboards**: Advanced Kubernetes and xApp lifecycle management
- **🔌 O-RAN Standards Compliance**: Full E2, A1, and O1 interface implementations
- **⚙️ Production-Ready Deployment**: Complete Kubernetes automation with Helm charts

### Current Completion Status

| Component | Status | Description |
|-----------|--------|-------------|
| **Main Dashboard** | ✅ **Complete** | Go backend + Angular frontend for Kubernetes management |
| **xApp Dashboard** | ✅ **Complete** | Angular dashboard for xApp lifecycle management |
| **O-RAN Interfaces** | ✅ **Implemented** | E2, A1, O1 interface support with simulators |
| **Federated Learning** | ✅ **Advanced** | FL coordinator with privacy-preserving capabilities |
| **CI/CD Pipeline** | ✅ **Full Coverage** | Multi-stage pipeline with security scanning |
| **Helm Deployment** | ✅ **Production Ready** | Complete Kubernetes deployment automation |
| **Documentation** | ✅ **Comprehensive** | User, developer, and operations guides |

## 🏗️ Architecture

The platform consists of two primary Angular dashboards with supporting O-RAN infrastructure:

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                        O-RAN Near-RT RIC Platform                                  │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  📊 Main Dashboard           🚀 xApp Dashboard         🤖 FL Coordinator           │
│  (Go + Angular)              (Angular + D3.js)        (Privacy-Preserving ML)     │
│  Port: 8080                  Port: 4200               Port: 8090                   │
│                                                                                     │
│  • Kubernetes Management     • xApp Lifecycle Mgmt    • Model Aggregation         │
│  • Cluster Monitoring        • Container Registry     • xApp Intelligence         │
│  • RBAC Control             • Image History           • RRM Optimization          │
│  • Resource Management       • YANG Tree Browser      • Network Slice Privacy     │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                           O-RAN Interface Layer                                    │
│                                                                                     │
│  📡 E2 Interface           📋 A1 Interface           🔧 O1 Interface               │
│  • Real-time RAN Control   • Policy Management       • Operations & Maintenance   │
│  • 10ms-1s Latency        • ML Model Distribution    • Configuration Management   │
│  • KPM, RC, NI Services   • Intent-based Control     • Fault & Performance Mgmt  │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                        Kubernetes Infrastructure                                   │
│  • Multi-arch Containers  • Service Discovery        • Prometheus Monitoring      │
│  • Helm Chart Deployment  • Persistent Storage       • Grafana Dashboards        │
│  • RBAC & Security        • Network Policies         • Redis & PostgreSQL        │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## 📦 Project Structure

```
near-rt-ric/
├── 📊 dashboard-master/
│   └── dashboard-master/           # Main Kubernetes Dashboard (Go + Angular)
│       ├── 🔧 src/app/backend/     # Go backend API server
│       ├── 🎨 src/app/frontend/    # Angular 13.3 frontend
│       ├── 📋 aio/                 # Build configuration & Docker
│       ├── 🧪 cypress/             # E2E testing
│       └── 📚 docs/                # User and developer documentation
├── 🚀 xAPP_dashboard-master/       # xApp Management Dashboard (Angular)
│   ├── 🎨 src/app/                 # Angular application
│   ├── 📊 src/app/components/      # Visualization components (D3.js, ECharts)
│   └── 🧪 cypress/                 # E2E testing with Cypress
├── ⚙️ helm/
│   └── oran-nearrt-ric/           # Production Helm charts
├── 🏗️ k8s/                        # Kubernetes manifests
│   ├── oran/                      # O-RAN specific components
│   └── sample-xapps/              # Sample xApp deployments
├── 🤖 scripts/                     # Setup and utility scripts
│   ├── setup.sh                   # Linux/macOS setup
│   ├── setup.ps1                  # Windows PowerShell setup
│   └── check-prerequisites.*      # Prerequisites checker
├── ⚙️ config/                      # Configuration files
│   └── prometheus/                # Monitoring configuration
├── 📚 docs/                        # Project documentation
│   ├── operations/                # Deployment and ops guides
│   └── developer/                 # Developer documentation
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

**Linux/macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/hctsai1006/near-rt-ric/main/scripts/setup.sh | bash
```

**Windows (PowerShell):**
```powershell
.\scripts\setup.ps1
```

### 🛠️ Manual Setup

#### 1. Clone and Verify
```bash
git clone https://github.com/hctsai1006/near-rt-ric.git
cd near-rt-ric

# Check prerequisites
.\scripts\check-prerequisites.ps1   # Windows
./scripts/check-prerequisites.sh    # Linux/macOS
```

#### 2. Local Development Environment
```bash
# Create KIND cluster
kind create cluster --name near-rt-ric --config kind-config.yaml

# Start development with Docker Compose
docker-compose up -d

# Or deploy with Helm for production-like testing
helm install oran-nearrt-ric helm/oran-nearrt-ric/ \
  --create-namespace --namespace oran-nearrt-ric
```

#### 3. Individual Dashboard Development

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

#### 4. Access and Verification
```bash
# Check deployment status
kubectl get pods -A

# Access dashboards
echo "📊 Main Dashboard: http://localhost:8080"
echo "🚀 xApp Dashboard: http://localhost:4200"
echo "📈 Prometheus: http://localhost:9090"
echo "📊 Grafana: http://localhost:3000"

# Test O-RAN interfaces
curl http://localhost:8080/api/v1/e2/health     # E2 interface
curl http://localhost:8080/api/v1/a1/policies   # A1 interface
```

## 🚀 Complete Deployment Guide

### 🔧 Development Deployment

#### Using KIND (Recommended)
```bash
# 1. Create multi-node cluster
kind create cluster --name near-rt-ric --config kind-config.yaml

# 2. Deploy with Helm
helm install oran-nearrt-ric helm/oran-nearrt-ric/ \
  --create-namespace --namespace oran-nearrt-ric \
  --set global.environment=development \
  --set monitoring.enabled=true

# 3. Port forward for access
kubectl port-forward -n oran-nearrt-ric service/main-dashboard 8080:8080 &
kubectl port-forward -n oran-nearrt-ric service/xapp-dashboard 4200:80 &
```

#### Using Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f main-dashboard
docker-compose logs -f xapp-dashboard

# Stop services
docker-compose down
```

### 🏭 Production Deployment

#### Prerequisites
- Kubernetes cluster v1.21+ with at least 3 nodes
- Persistent storage class configured
- Ingress controller installed
- SSL certificates for TLS termination

#### Deployment Steps
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
  --set global.environment=production \
  --set mainDashboard.replicaCount=3 \
  --set xappDashboard.replicaCount=3 \
  --set mainDashboard.ingress.enabled=true \
  --set mainDashboard.ingress.hosts[0].host=oran.example.com \
  --set monitoring.prometheus.persistence.enabled=true \
  --wait --timeout=15m

# 3. Verify deployment
kubectl get all -n oran-nearrt-ric
kubectl get ingress -n oran-nearrt-ric
```

### 🔧 Environment Variables

#### Main Dashboard Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `BIND_ADDRESS` | `0.0.0.0` | Server bind address |
| `PORT` | `8080` | Server port |
| `TOKEN_TTL` | `900` | JWT token TTL (seconds) |
| `AUTO_GENERATE_CERTIFICATES` | `false` | Auto-generate TLS certs |
| `ENABLE_INSECURE_LOGIN` | `false` | Allow insecure login (dev only) |

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
| `FL_AGGREGATION_STRATEGY` | `fedavg` | Model aggregation strategy |
| `FL_MIN_CLIENTS` | `2` | Minimum clients for aggregation |
| `FL_ROUND_TIMEOUT` | `300` | FL round timeout (seconds) |

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

# Production
make prod                 # Production build
make deploy               # Deploy to Kubernetes
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
npm run e2e               # E2E tests with Cypress

# Code Quality
npm run lint              # ESLint check
npm run lint:fix          # Auto-fix linting issues
npm run format            # Prettier formatting
```

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
  -d '{"name": "traffic-prediction-xapp", "image": "oran/traffic-prediction:v1.0"}' \
  http://localhost:8080/api/v1/xapps
```

### 🚀 xApp Dashboard Operations

```bash
# List available xApps
curl http://localhost:4200/api/xapps

# Get xApp deployment status
curl http://localhost:4200/api/xapps/traffic-prediction-xapp/status

# View xApp logs
curl http://localhost:4200/api/xapps/traffic-prediction-xapp/logs?lines=100
```

### 🤖 Federated Learning API

```bash
# Register xApp as FL client
curl -X POST -H "Content-Type: application/json" \
  -d '{"client_id": "traffic-prediction-xapp", "model_type": "neural_network"}' \
  http://localhost:8090/fl/clients

# Start federated learning round
curl -X POST http://localhost:8090/fl/rounds/start

# Download global model
curl http://localhost:8090/fl/models/global/latest -o global_model.h5
```

### 📊 O-RAN Interface Examples

```bash
# E2 Interface - Get RAN node status
curl http://localhost:8080/api/v1/e2/ran-nodes

# A1 Interface - Deploy ML model policy
curl -X PUT -H "Content-Type: application/json" \
  -d '{"policy_type_id": 1, "policy_id": "traffic-prediction-policy"}' \
  http://localhost:8080/api/v1/a1/policies/traffic-prediction-policy

# O1 Interface - Get configuration
curl http://localhost:8080/api/v1/o1/config/ran-nodes/enb001
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test                    # Main dashboard (from dashboard-master/dashboard-master/)
npm test                     # xApp dashboard (from xAPP_dashboard-master/)

# Run specific test suites
make test-backend           # Go backend tests only
npm run test:ci             # Angular tests (CI mode)
npm run e2e                 # End-to-end tests with Cypress

# Generate coverage reports
make coverage               # Main dashboard coverage
npm run test:coverage       # xApp dashboard coverage
```

### Test Coverage

| Component | Unit Tests | Integration Tests | E2E Tests | Coverage |
|-----------|------------|------------------|-----------|----------|
| **Main Dashboard Backend** | ✅ Go test suite | ✅ API tests | ✅ Full UI flow | 85%+ |
| **Main Dashboard Frontend** | ✅ Karma/Jasmine | ✅ Component tests | ✅ Angular E2E | 80%+ |
| **xApp Dashboard** | ✅ Jest/Karma | ✅ Component tests | ✅ Cypress E2E | 75%+ |
| **O-RAN Interfaces** | ✅ Unit tests | ✅ Interface tests | ✅ Protocol tests | 90%+ |

### CI/CD Pipeline

The comprehensive CI/CD pipeline includes:

- **Security Scanning**: Trivy vulnerability scanning, secret detection with Gitleaks
- **Code Quality**: golangci-lint, ESLint, multi-component testing
- **Multi-Platform Builds**: AMD64 and ARM64 container images
- **Helm Chart Testing**: Chart linting and installation validation
- **E2E Testing**: Full platform integration tests with KIND
- **Performance Testing**: Load testing with k6
- **Staging Deployment**: Automated deployment to staging environment

## 🔧 API Reference

### 📡 Main Dashboard API (Port 8080)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/login/status` | GET | Authentication status |
| `/api/v1/cluster/overview` | GET | Cluster resource overview |
| `/api/v1/namespaces` | GET | List all namespaces |
| `/api/v1/pods/{namespace}` | GET | List pods in namespace |
| `/api/v1/xapps` | GET/POST | xApp management |
| `/api/v1/e2/subscriptions` | GET/POST | E2 interface subscriptions |
| `/api/v1/a1/policies` | GET/POST/PUT | A1 policy management |

### 🚀 xApp Dashboard API (Port 4200)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/xapps` | GET | List deployed xApps |
| `/api/xapps/{id}/status` | GET | xApp deployment status |
| `/api/xapps/{id}/logs` | GET | xApp logs and metrics |
| `/api/images` | GET | Container image registry |
| `/yangTree` | GET | YANG data model browser |

### 🤖 Federated Learning API (Port 8090)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/fl/clients` | GET/POST | FL client management |
| `/fl/rounds/start` | POST | Start FL training round |
| `/fl/models/global/latest` | GET | Download global model |
| `/fl/aggregation/strategy` | PUT | Set aggregation strategy |

## 🌟 Key Features

### 🎯 O-RAN Standards Compliance
- **E2 Interface**: Real-time RAN control with 10ms-1s latency (ASN.1/SCTP)
- **A1 Interface**: Policy and ML model management (REST/JSON)
- **O1 Interface**: Operations and maintenance (NETCONF/YANG)
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

## 🚨 Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| **Port conflicts** | Use `docker-compose down` and check `docker ps` |
| **Permission errors** | Ensure Docker daemon is running and user is in docker group |
| **Build failures** | Run `make clean && make build` to rebuild from scratch |
| **K8s connection issues** | Verify `kubectl cluster-info` returns valid cluster info |
| **npm test failures** | Ensure dependencies installed via `npm ci` |
| **CI/CD failures** | Check GitHub Actions logs for specific error details |

### Getting Help

- **[Issue Tracker](https://github.com/hctsai1006/near-rt-ric/issues)** - Report bugs and request features
- **[Troubleshooting Guide](docs/operations/troubleshooting.md)** - Detailed problem resolution
- **[O-RAN Software Community](https://o-ran-sc.org/)** - Community support and resources

### Development Support

```bash
# Check prerequisites
.\scripts\check-prerequisites.ps1  # Windows
./scripts/check-prerequisites.sh   # Linux/macOS

# Validate environment
kubectl cluster-info               # Verify Kubernetes access
docker info                        # Verify Docker status
```

## 📚 Documentation

### 📖 User Documentation
- **[Installation Guide](dashboard-master/dashboard-master/docs/user/installation.md)** - Complete setup instructions
- **[User Manual](dashboard-master/dashboard-master/docs/user/README.md)** - Dashboard usage guide
- **[Access Control](dashboard-master/dashboard-master/docs/user/access-control/README.md)** - RBAC and authentication

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
.\scripts\setup.ps1                    # Windows full setup
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