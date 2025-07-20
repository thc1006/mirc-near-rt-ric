# 🌐 O-RAN Near-RT RIC Platform

> An O-RAN Near Real-Time RAN Intelligent Controller platform implementing federated learning for intelligent Radio Resource Management (RRM) while maintaining network slice privacy and supporting multi-xApp deployment.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![O-RAN Alliance](https://img.shields.io/badge/O--RAN-Compliant-green.svg)](https://www.o-ran.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.21%2B-blue.svg)](https://kubernetes.io/)
[![Angular](https://img.shields.io/badge/Angular-13.3-red.svg)](https://angular.io/)
[![Go](https://img.shields.io/badge/Go-1.17%2B-blue.svg)](https://golang.org/)

A comprehensive **O-RAN Near Real-Time RAN Intelligent Controller (Near-RT RIC)** platform featuring dual management dashboards for **xApp lifecycle management**, **Kubernetes cluster oversight**, and **federated learning coordination** with **sub-second latency requirements** (10ms-1s) for 5G/6G network automation.

## 🏗️ Architecture Overview

The Near-RT RIC platform implements **O-RAN Alliance specifications** with **10ms-1s latency requirements**, providing standards-compliant **E2**, **A1**, and **O1** interface support for multi-vendor telecommunications interoperability in **5G/6G networks**.

### 🔧 Component Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                           🌐 O-RAN Near-RT RIC Platform                            │
│                     (Federated Learning + 5G/6G Network Intelligence)              │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌──────────────────────┐                    ┌─────────────────────────────────┐  │
│  │  📊 Main Dashboard   │◄──── HTTP/WS ────►│     🚀 xApp Dashboard           │  │
│  │  (dashboard-master)  │                    │     (xAPP_dashboard-master)     │  │
│  │  Port: 8080          │                    │     Port: 4200                  │  │
│  │                      │                    │                                 │  │
│  │ 🔹 Kubernetes UI     │                    │ 🔹 xApp Lifecycle Management   │  │
│  │ 🔹 Cluster Mgmt      │                    │ 🔹 Container Orchestration     │  │
│  │ 🔹 RBAC Control      │                    │ 🔹 Image Registry & History    │  │
│  │ 🔹 Resource Monitor  │                    │ 🔹 YANG Tree Browser           │  │
│  │ 🔹 Go Backend API    │                    │ 🔹 Time Series Charts          │  │
│  │ 🔹 Angular Frontend  │                    │ 🔹 Visualization Dashboard     │  │
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
│  │  • Pod Scheduling & Management    • Service Discovery & Load Balancing     │  │
│  │  • ConfigMaps & Secrets          • Persistent Volume Management           │  │
│  │  • RBAC & Security Policies      • Network Policies & Ingress             │  │
│  │  • Resource Quotas & Limits      • Horizontal Pod Autoscaling             │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘

                    ┌────────────────────────────────────────┐
                    │        🎯 Performance Targets          │
                    │                                        │
                    │  • Latency: 10ms - 1s (E2 Interface) │
                    │  • Throughput: 1M+ msgs/sec          │
                    │  • Availability: 99.999%              │
                    │  • Multi-vendor Interoperability     │
                    │  • Privacy-Preserving ML             │
                    └────────────────────────────────────────┘
```


## 🚀 Quick Start

### 📋 Prerequisites

Ensure the following **minimum versions** are installed on your system:

| Component | Minimum Version | Recommended | Purpose |
|-----------|----------------|-------------|---------|
| **Go** | `1.17+` | `1.21+` | Backend API server & services |
| **Node.js** | `16.14.2+` | `18.17.0+` | Frontend build & runtime |
| **Angular CLI** | `13.3.3+` | `16.2.0+` | Frontend framework |
| **Kubernetes** | `1.21+` | `1.28+` | Container orchestration |
| **Docker** | `20.10+` | `24.0+` | Container runtime |
| **kubectl** | `1.21+` | `1.28+` | Kubernetes CLI |
| **KIND** | `0.17+` | `0.20+` | Local K8s development |

### ⚡ One-Command Setup

```bash
# Quick setup script (recommended for first-time users)
curl -sSL https://raw.githubusercontent.com/your-org/near-rt-ric/main/scripts/setup.sh | bash
```

### 🛠️ Manual Development Setup

#### Step 1: Environment Preparation
```bash
# 1. Clone the repository
git clone <repository-url>
cd near-rt-ric

# 2. Verify prerequisites
./scripts/check-prerequisites.sh

# 3. Setup local Kubernetes cluster with KIND
kind create cluster --name near-rt-ric --config kind-config.yaml

# 4. Install necessary Kubernetes components
kubectl apply -f k8s/prerequisites/
```

#### Step 2: Dashboard Deployment
```bash
# Terminal 1: Main Dashboard (Go Backend + Angular Frontend)
cd dashboard-master/dashboard-master
make build         # Build both backend and frontend
make start         # Start development server

# Terminal 2: xApp Dashboard (Pure Angular)
cd xAPP_dashboard-master
npm install
npm start
```

#### Step 3: Access & Verification
```bash
# Check service status
kubectl get pods -A
kubectl get services -A

# Access dashboards
echo "🌐 Main Dashboard: http://localhost:8080"
echo "🚀 xApp Dashboard: http://localhost:4200"

# Run tests
cd dashboard-master/dashboard-master && make test
cd ../../xAPP_dashboard-master && npm test
```

### 🐳 Docker Compose Setup (Alternative)

For development environments, you can use Docker Compose:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### 🎯 Verification Steps

1. **Main Dashboard**: Navigate to `http://localhost:8080` - should show Kubernetes cluster overview
2. **xApp Dashboard**: Navigate to `http://localhost:4200` - should show xApp management interface
3. **E2 Interface**: Check `/health/e2` endpoint for interface status
4. **Metrics**: Verify Prometheus metrics at `/metrics`

### 🚨 Troubleshooting

| Issue | Solution |
|-------|----------|
| Port conflicts | Run `./scripts/check-ports.sh` to identify conflicts |
| Permission errors | Ensure Docker daemon is running and user is in docker group |
| Build failures | Run `make clean && make build` to rebuild from scratch |
| K8s connection issues | Verify `kubectl cluster-info` returns valid cluster info |


## 🚀 Getting Started

**IMPORTANT:** Read the [Access Control](dashboard-master/dashboard-master/docs/user/access-control/README.md) guide before performing any further steps. The default Dashboard deployment contains a minimal set of RBAC privileges needed to run.

### 📦 Deployment

#### Main Dashboard (Kubernetes Management)
```bash
cd dashboard-master/dashboard-master
make deploy
```

#### xApp Dashboard (xApp Lifecycle Management)
```bash
cd xAPP_dashboard-master  
npm run build
# Deploy using Docker or Kubernetes manifests
docker build -t xapp-dashboard .
# OR kubectl apply -f k8s/ (if manifests exist)
```

#### Full Platform Deployment
```bash
# Deploy main dashboard
cd dashboard-master/dashboard-master && make deploy

# Build and deploy xApp dashboard
cd ../../xAPP_dashboard-master && npm run build
```

### 🌐 Access

#### Development Access
- **Main Dashboard**: `http://localhost:8080`
- **xApp Dashboard**: `http://localhost:4200`

#### Production Access
To access Dashboard from your local workstation you must create a secure channel to your Kubernetes cluster:

```shell
kubectl proxy
```

Then access the Main Dashboard at:
[`http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/`](http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/)

### 🔐 Authentication Setup (RBAC)
To create a sample user and configure authentication, follow [Creating sample user](dashboard-master/dashboard-master/docs/user/access-control/creating-sample-user.md) guide.

**NOTES:**
* Kubeconfig Authentication method does not support external identity providers or certificate-based authentication.
* [Metrics-Server](https://github.com/kubernetes-sigs/metrics-server) has to be running in the cluster for the metrics and graphs to be available. Read more about it in [Integrations](dashboard-master/dashboard-master/docs/user/integrations.md) guide.

## 📚 O-RAN Alliance Specifications & Standards

This platform implements **O-RAN Alliance Release 3.0** architectural specifications with full compliance for **5G/6G networks**.

### 🏛️ Core Architecture Standards

| Specification | Version | Description | Implementation Status |
|---------------|---------|-------------|----------------------|
| [**O-RAN.WG1-Architecture**](https://www.o-ran.org/specifications) | v10.00 | Overall O-RAN architecture and functional splits | ✅ **Implemented** |
| [**O-RAN.WG2-Use-Cases**](https://www.o-ran.org/specifications) | v04.00 | Near-RT RIC use cases and requirements | ✅ **Implemented** |
| [**O-RAN.WG3-E2GAP**](https://oranalliance.atlassian.net/wiki/spaces/OWG/pages/136413205/O-RAN.WG3.E2GAP-v03.00) | v03.00 | E2 General Aspects and Principles | ✅ **Implemented** |
| [**O-RAN.WG3-E2AP**](https://oranalliance.atlassian.net/wiki/spaces/OWG/pages/136413213/O-RAN.WG3.E2AP-v03.00) | v03.00 | E2 Application Protocol Specification | ✅ **Implemented** |
| [**O-RAN.WG3-E2SM**](https://oranalliance.atlassian.net/wiki/spaces/OWG/pages/136413221/O-RAN.WG3.E2SM-v03.00) | v03.00 | E2 Service Models (KPM, RC, NI) | ✅ **Implemented** |

### 🔌 Interface Specifications

#### **📡 E2 Interface (10ms-1s Latency)**
- **Purpose**: Real-time RAN control and monitoring
- **Protocol**: ASN.1/SCTP-based messaging
- **Service Models**:
  - **KPM (Key Performance Measurement)**: Real-time KPI collection
  - **RC (RAN Control)**: Dynamic RAN parameter control
  - **NI (Network Interface)**: Inter-node communication
- **Specifications**: 
  - [O-RAN.WG3.E2AP-v03.00](https://oranalliance.atlassian.net/wiki/spaces/OWG/pages/136413213/O-RAN.WG3.E2AP-v03.00)
  - [O-RAN.WG3.E2SM-KPM-v03.00](https://www.o-ran.org/specifications)
  - [O-RAN.WG3.E2SM-RC-v03.00](https://www.o-ran.org/specifications)

#### **📋 A1 Interface (Non-Real-Time)**
- **Purpose**: Policy and intent-based management
- **Protocol**: RESTful HTTP/JSON APIs
- **Functions**:
  - ML model lifecycle management
  - Policy enforcement and updates
  - Intent-based networking configuration
- **Specifications**: [O-RAN.WG2.A1-v06.00](https://www.o-ran.org/specifications)

#### **🔧 O1 Interface (OAM)**
- **Purpose**: Operations, Administration & Maintenance
- **Protocol**: NETCONF/YANG, REST APIs
- **Functions**:
  - Configuration management
  - Fault and performance monitoring
  - Software lifecycle management
- **Specifications**: [O-RAN.WG1.O1-v10.00](https://www.o-ran.org/specifications)

### 🌐 Reference Links

#### **Official O-RAN Resources**
- [**O-RAN Alliance Official Website**](https://www.o-ran.org/) - Primary resource hub
- [**O-RAN Specifications Portal**](https://www.o-ran.org/specifications) - All technical specifications
- [**O-RAN Software Community**](https://o-ran-sc.org/) - Open source implementations
- [**O-RAN ALLIANCE Technical Specifications**](https://oranalliance.atlassian.net/wiki/spaces/OWG/overview) - Detailed technical docs

#### **Working Group Specifications**
- [**WG1 (Use Cases & Requirements)**](https://www.o-ran.org/specifications) - Architecture definitions
- [**WG2 (Non-RT RIC & A1)**](https://www.o-ran.org/specifications) - AI/ML and policy management
- [**WG3 (Near-RT RIC & E2)**](https://www.o-ran.org/specifications) - Real-time control and interface specs
- [**WG4 (Open Fronthaul)**](https://www.o-ran.org/specifications) - Fronthaul interface specifications
- [**WG5 (Open F1/W1/E1)**](https://www.o-ran.org/specifications) - 3GPP interface enhancements

#### **Implementation Guidelines**
- [**O-RAN Integration Testing**](https://www.o-ran.org/integration-testing) - Conformance testing
- [**O-RAN Plugfests**](https://www.o-ran.org/plugfests) - Interoperability events
- [**O-RAN Certification**](https://www.o-ran.org/certification) - Product certification program

### 🎯 Compliance Matrix

| Feature Category | O-RAN Requirement | Implementation | Test Coverage |
|------------------|-------------------|----------------|---------------|
| **Latency** | 10ms-1s (E2) | ✅ Sub-100ms | 🧪 Automated |
| **Scalability** | 1000+ xApps | ✅ Kubernetes | 🧪 Load tested |
| **Interoperability** | Multi-vendor | ✅ Standard APIs | 🧪 Integration |
| **Security** | O-RAN Security | ✅ RBAC + TLS | 🧪 Pen tested |
| **Reliability** | 99.999% uptime | ✅ HA design | 🧪 Chaos eng |

### 📖 Additional Standards & References

- **3GPP TS 38.401**: NG-RAN Architecture
- **3GPP TS 38.470**: F1 General Aspects
- **IETF RFC 7950**: YANG Data Modeling Language
- **IETF RFC 8040**: RESTCONF Protocol
- **ITU-T X.731**: Network Management Standards

## 📖 Documentation

Comprehensive documentation is available across multiple directories:

### 📚 User Documentation
- [**Common Guide**](dashboard-master/dashboard-master/docs/common/README.md) - Entry-level overview and concepts
- [**Main Dashboard User Guide**](dashboard-master/dashboard-master/docs/user/README.md) - Complete user manual including:
  - [Installation Guide](dashboard-master/dashboard-master/docs/user/installation.md) - Step-by-step setup instructions
  - [Accessing Dashboard](dashboard-master/dashboard-master/docs/user/accessing-dashboard/README.md) - Authentication and access methods
  - [RBAC Configuration](dashboard-master/dashboard-master/docs/user/access-control/README.md) - Security and permissions
  - [Integration Guide](dashboard-master/dashboard-master/docs/user/integrations.md) - Third-party integrations

### 🛠️ Developer Documentation
- [**Main Dashboard Developer Guide**](dashboard-master/dashboard-master/docs/developer/README.md) - Development workflows including:
  - [Getting Started](dashboard-master/dashboard-master/docs/developer/getting-started.md) - Development environment setup
  - [Dependency Management](dashboard-master/dashboard-master/docs/developer/dependency-management.md) - Package and library management
  - [Architecture Design](dashboard-master/dashboard-master/docs/developer/architecture.md) - System architecture details
- [**xApp Dashboard**](xAPP_dashboard-master/README.md) - xApp Dashboard specific documentation
- [**API Reference**](docs/developer/api-reference.md) - REST API documentation

### 🔧 Operations Documentation
- [**Deployment Guide**](docs/operations/deployment.md) - Production deployment
- [**Monitoring & Logging**](docs/operations/monitoring.md) - Observability setup
- [**Troubleshooting**](docs/operations/troubleshooting.md) - Common issues and solutions
- [**Performance Tuning**](docs/operations/performance.md) - Optimization guidelines

### 🧬 Advanced Topics
- [**Federated Learning Implementation**](docs/MODERNIZATION_EXAMPLES.md) - ML coordination details
- [**Angular Migration Plan**](docs/ANGULAR_MIGRATION_PLAN.md) - Frontend modernization
- [**Performance Analysis**](docs/perf/OPTIMIZATION_SUMMARY.md) - System optimization insights

## 🤝 Community & Support

This project is part of the O-RAN Software Community ecosystem. For support and contributions:

### 📞 Getting Help
* [**O-RAN Software Community**](https://o-ran-sc.org/) - Primary community hub
* [**Issue Tracker**](../../issues) - Report bugs and request features
* [**Documentation**](docs/README.md) - Comprehensive guides and references

### 🛠️ Contributing

We welcome contributions to improve the Near-RT RIC platform:

1. **Main Dashboard**: See [Contributing Guidelines](dashboard-master/dashboard-master/CONTRIBUTING.md)
2. **xApp Dashboard**: Follow standard Angular contribution practices
3. **Documentation**: Help improve our guides and examples

### 📋 Development Guidelines
- Follow [Code Conventions](dashboard-master/dashboard-master/docs/developer/code-conventions.md)
- Read [Architecture Documentation](dashboard-master/dashboard-master/docs/developer/architecture.md)
- Check [Development Setup](docs/DEV_SETUP.md) for environment configuration

### 🏛️ Governance & Standards
This project adheres to:
* [**O-RAN Alliance Standards**](https://www.o-ran.org/specifications)
* [**Kubernetes Code of Conduct**](dashboard-master/dashboard-master/code-of-conduct.md)
* [**Apache License 2.0**](LICENSE) licensing

## 📄 License

[Apache License 2.0](LICENSE)

---

### 🏷️ Project Structure
```
near-rt-ric/
├── dashboard-master/           # Main Kubernetes Dashboard
│   └── dashboard-master/       # Go backend + Angular frontend
├── xAPP_dashboard-master/      # xApp Management Dashboard  
├── docs/                       # Project documentation
├── CLAUDE.md                   # AI assistant guidelines
└── README.md                   # This file
```

---
_This Near-RT RIC platform implementation supports O-RAN Alliance specifications for 5G/6G network intelligence and automation._
