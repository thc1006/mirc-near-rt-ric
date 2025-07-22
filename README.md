# O-RAN Near-RT RIC Platform

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI/CD Pipeline](https://github.com/hctsai1006/near-rt-ric/actions/workflows/ci-integrated.yml/badge.svg)](https://github.com/hctsai1006/near-rt-ric/actions)
[![Security Scan](https://github.com/hctsai1006/near-rt-ric/actions/workflows/security-complete.yml/badge.svg)](https://github.com/hctsai1006/near-rt-ric/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg)](https://golang.org/)

> **Modern O-RAN Near Real-Time RAN Intelligent Controller with interactive dashboard, production-grade SMO stack, and comprehensive observability pipeline.**

## 🎯 Executive Summary

This repository contains a **complete O-RAN Near Real-Time RAN Intelligent Controller (Near-RT RIC)** platform designed for 5G/6G network optimization. The platform provides intelligent network management, federated learning coordination, and comprehensive xApp lifecycle management through dual dashboards with production-ready Kubernetes deployment.

### 🏆 Key Achievements

- ✅ **Full O-RAN Compliance**: E2, A1, O1 interfaces with 10ms-1s latency requirements
- ✅ **Production-Ready Federated Learning**: Privacy-preserving ML coordination across network slices
- ✅ **Advanced Kubernetes and xApp Lifecycle Management**
- ✅ **One-Command Deployment**: Complete `make deploy` or `helm install` automation
- ✅ **Comprehensive CI/CD**: Multi-platform builds with security scanning and 3 optimized workflows
- ✅ **Security Hardened**: Container security contexts, network policies, and vulnerability scanning
- ✅ **Enterprise-Grade Security**: RBAC, TLS, container vulnerability scanning

## 🏗️ System Architecture

The O-RAN Near-RT RIC platform is composed of several microservices that work together to provide a comprehensive solution for managing and optimizing the Radio Access Network (RAN).

```
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                    O-RAN Near-RT RIC Platform Architecture                                 │
│                          (Production-Ready Implementation)                                 │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│                              Management & Control Layer                                    │
├─────────────────────────┬─────────────────────────┬────────────────────────────────────────┤
│  📊 Main Dashboard      │  🚀 xApp Dashboard      │  🤖 Federated Learning Coordinator    │
│  (Go + Angular)         │  (Angular + D3.js)      │  (Go + gRPC + Redis)                  │
│  Port: 8080/8443        │  Port: 4200             │  Port: 8090                           │
│                         │                         │                                       │
│  • K8s Cluster Mgmt    │  • xApp Lifecycle       │  • Privacy-Preserving ML             │
│  • Real-time Monitor   │  • Container Registry   │  • FedAvg/FedProx Aggregation        │
│  • RBAC & Security     │  • Image History Mgmt   │  • Byzantine Fault Tolerance         │
│  • Resource Scaling    │  • YANG Tree Browser    │  • Multi-Region Coordination         │
│  • E2/A1/O1 Control    │  • Performance Analytics│  • Dynamic Resource Management       │
├─────────────────────────┴─────────────────────────┴────────────────────────────────────────┤
│                              O-RAN Interface Layer                                        │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│  📡 E2 Interface            📋 A1 Interface             🔧 O1 Interface                    │
│  • ASN.1/SCTP Protocol     • REST/JSON API             • NETCONF/YANG Protocol          │
│  • 10ms-1s Latency SLA     • Policy Management         • Configuration Management       │
│  • KPM (v3.0) Metrics      • ML Model Distribution     • Fault Management               │
│  • RC (v3.0) Control       • Intent-based Control      • Performance Management         │
│  • NI (v1.0) Insertion     • xApp Orchestration        • Software Management            │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│                           Cloud-Native Infrastructure                                     │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│  🏗️ Kubernetes Orchestration              📊 Observability Stack                       │
│  • Multi-arch Deployments (AMD64/ARM64)   • Prometheus + Grafana                        │
│  • Helm Chart Automation                  • OpenTelemetry Tracing                       │
│  • HPA & VPA Scaling                      • Structured Logging (JSON)                   │
│  • Network Policies                       • Health Checks & Probes                      │
│  • Service Mesh Ready                     • SLI/SLO Monitoring                         │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│                              Data & Storage Layer                                         │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│  🗄️ Persistent Storage                    🔐 Security & Compliance                      │
│  • Redis Cluster (HA)                     • RBAC & Pod Security Standards               │
│  • PostgreSQL (Multi-AZ)                  • TLS 1.3 Encryption                         │
│  • Model Storage (S3/PVC)                 • Secret Management (Vault)                   │
│  • Metrics Retention                      • Vulnerability Scanning                     │
│  • Backup & Recovery                      • SOC 2 / ISO 27001 Ready                    │
└──────────────────────────────────────────────────────────────────────────────────────────┘
```

## ⚡ Quick Start

### 🛠️ Prerequisites

| Component | Minimum | Recommended | Purpose |
|-----------|---------|-------------|---------|
| **Docker** | 20.10+ | 24.0+ | Container runtime |
| **kubectl** | 1.21+ | 1.28+ | Kubernetes CLI |
| **Helm** | 3.8+ | 3.13+ | Package manager |
| **KIND** | 0.17+ | 0.20+ | Local development |
| **Go** | 1.17+ | 1.22+ | Backend development |

### 🚀 One-Command Production Deployment

```bash
# Clone repository
git clone https://github.com/hctsai1006/near-rt-ric.git
cd near-rt-ric

# Option 1: Fully automated deployment (Recommended)
./deploy.sh

# Option 2: Make-based deployment
make deploy

# Option 3: Manual Helm deployment
helm dependency build helm/oran-nearrt-ric/
helm install oran-nearrt-ric helm/oran-nearrt-ric/ \
  --create-namespace --namespace oran-nearrt-ric \
  --set oran.enabled=true \
  --set monitoring.prometheus.enabled=true \
  --set monitoring.grafana.enabled=true \
  --wait --timeout=10m
```

**📋 What gets deployed:**

- ✅ **Near-RT RIC Platform**: E2 Termination, A1 Policy, O1 Management
- ✅ **O-RU Simulator**: 2 Radio Units with 64 antennas @ 3.7GHz
- ✅ **O-DU Simulator**: Distributed Unit with E2/F1 interfaces
- ✅ **O-CU Simulator**: Central Unit (CP/UP) with E2/F1/NG interfaces
- ✅ **Management Dashboards**: Main K8s + xApp dashboards
- ✅ **Federated Learning**: Privacy-preserving ML coordination
- ✅ **Monitoring Stack**: Prometheus + Grafana with O-RAN metrics

### 🛠️ Development Environment Setup

```bash
# 1. Setup local development environment
./scripts/setup.sh                    # Linux/macOS
.\scripts\setup.ps1                   # Windows

# 2. Start with Docker Compose (fastest)
docker-compose up -d

# 3. Access dashboards
echo "🚀 xApp Dashboard: http://localhost:4200"
echo "🤖 FL Coordinator: http://localhost:8090"
echo "📈 Prometheus: http://localhost:9090"
echo "📊 Grafana: http://localhost:3000 (admin/admin123)"
```

## 📋 Project Structure

```
near-rt-ric/
├── api/
│   ├── a1/
│   ├── e2/
│   └── o1/
├── bin/
├── cmd/
│   ├── ric-a1/
│   ├── ric-control/
│   ├── ric-e2/
│   └── ric-o1/
├── config/
│   ├── grafana/
│   └── prometheus/
├── docker/
├── docs/
│   ├── developer/
│   ├── operations/
│   └── user/
├── helm/
│   ├── e2-simulator/
│   ├── observability-stack/
│   ├── oran-nearrt-ric/
│   ├── ric-a1/
│   ├── ric-e2/
│   ├── ric-o1/
│   └── smo-onap/
├── internal/
│   ├── config/
│   ├── control/
│   ├── o1/
│   └── tests/
├── k8s/
│   └── oran/
├── pkg/
│   ├── a1/
│   ├── a1ap/
│   ├── asn1/
│   ├── auth/
│   ├── config/
│   ├── e2/
│   ├── logging/
│   ├── messaging/
│   ├── o1/
│   ├── o1ap/
│   ├── storage/
│   ├── utils/
│   └── xapp/
├── scripts/
│   └── sql/
├── services/
│   ├── e2-simulator/
│   ├── ric-a1/
│   ├── ric-e2/
│   └── ric-o1/
└── web/
```

## 🧪 Testing and Validation

### 🔄 Running Complete Test Suite

```bash
# Run all backend tests (Go)
make test
```

## 🤝 Contributing

We welcome contributions to enhance the O-RAN Near-RT RIC platform!

### 📋 Contribution Guidelines

1. **Fork** and clone the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Write** comprehensive tests for new functionality
4. **Ensure** all tests pass (`make test`)
5. **Run** security and quality checks (`make lint && make security-scan`)
6. **Update** documentation for user-facing changes
7. **Commit** your changes with clear messages
8. **Push** to your fork and **create** a Pull Request

## 📄 License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.
