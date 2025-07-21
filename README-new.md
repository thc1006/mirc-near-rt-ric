# O-RAN Near-RT RIC Platform

[![CI/CD Pipeline](https://github.com/your-org/near-rt-ric/actions/workflows/ci-integrated.yml/badge.svg)](https://github.com/your-org/near-rt-ric/actions)
[![Security Scan](https://github.com/your-org/near-rt-ric/actions/workflows/security-complete.yml/badge.svg)](https://github.com/your-org/near-rt-ric/actions)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![O-RAN SC](https://img.shields.io/badge/O--RAN-Software%20Community-orange.svg)](https://o-ran-sc.org/)

> **Modern O-RAN Near Real-Time RAN Intelligent Controller with interactive dashboard, production-grade SMO stack, and comprehensive observability pipeline.**

## 🚀 Quick Start

**One-command deployment:**

```bash
make demo
```

This will deploy the complete O-RAN platform including:
- ✅ Near-RT RIC with modern React dashboard
- ✅ SMO stack based on ONAP Frankfurt
- ✅ Full observability pipeline (Grafana, Prometheus, Elasticsearch, Kafka)
- ✅ Auto-discovery of network functions
- ✅ Real-time KPIs and alarms

## 📊 Dashboard Preview

### Interactive O-RAN Operations Console
![O-RAN Dashboard](docs/images/dashboard-overview.png)
*Modern React-based dashboard with real-time network function discovery and KPI monitoring*

### Observability Stack
![Grafana Dashboard](docs/images/grafana-dashboard.png)
*Production-grade observability with Grafana, Prometheus, and Elasticsearch*

## 🏗️ Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph "O-RAN Near-RT RIC Platform"
        subgraph "Frontend Layer"
            FE[React Dashboard]
            XApp[xApp Management UI]
        end
        
        subgraph "API Gateway & Backend"
            API[Go Backend API]
            WS[WebSocket Server]
        end
        
        subgraph "Near-RT RIC Components"
            E2T[E2 Termination]
            A1M[A1 Mediator]
            O1M[O1 Mediator]
            AppMgr[App Manager]
            SubMgr[Subscription Manager]
            RtMgr[Routing Manager]
        end
        
        subgraph "SMO (ONAP Frankfurt)"
            AAI[A&AI]
            SO[Service Orchestrator]
            SDNC[SDN Controller]
            Policy[Policy Framework]
            DMaaP[DMaaP]
        end
        
        subgraph "Observability Stack"
            Grafana[Grafana]
            Prometheus[Prometheus]
            Elasticsearch[Elasticsearch]
            Kafka[Apache Kafka]
            Loki[Loki]
        end
        
        subgraph "O-RAN Network Functions"
            ORU[O-RU Simulator]
            ODU[O-DU Simulator] 
            OCU[O-CU Simulator]
        end
    end
    
    FE --> API
    XApp --> API
    API --> E2T
    API --> A1M
    API --> O1M
    
    E2T <--> ORU
    E2T <--> ODU
    E2T <--> OCU
    
    A1M <--> Policy
    O1M <--> AAI
    
    AppMgr --> SubMgr
    SubMgr --> RtMgr
    
    Prometheus --> Grafana
    Elasticsearch --> Grafana
    Kafka --> Elasticsearch
    
    style FE fill:#e1f5fe
    style SMO fill:#fff3e0
    style Observability fill:#f3e5f5
```

### Interface Specifications

| Interface | Standard | Purpose | Implementation |
|-----------|----------|---------|----------------|
| **E2** | O-RAN WG3 E2AP v03.00 | Near-RT RIC ↔ RAN Node | SCTP/TCP, ASN.1 encoding |
| **A1** | O-RAN WG2 A1AP v04.00 | Non-RT RIC ↔ Near-RT RIC | HTTP/REST, JSON |
| **O1** | O-RAN WG6 O1 v07.00 | SMO ↔ Managed Elements | NETCONF/RESTCONF, YANG models |
| **O2** | O-RAN WG6 O2 v01.00 | SMO ↔ Cloud Infrastructure | HTTP/REST, OpenStack APIs |

### Network Function Auto-Discovery

The platform automatically discovers and monitors:

- **Kubernetes Services** in O-RAN namespaces (`oran-nearrt-ric`, `onap`, `observability`)
- **Near-RT RIC Components** via E2 interface health checks
- **xApps** through App Manager APIs
- **SMO Components** via ONAP service discovery
- **O-RAN Simulators** with real-time status monitoring

## 🛠️ Prerequisites

### Development Environment

- **Docker** >= 20.10.0
- **Kubernetes** >= 1.24 (KIND, K3s, or minikube for local development)
- **Helm** >= 3.10.0
- **kubectl** >= 1.24.0
- **Node.js** >= 18.0.0 LTS
- **Go** >= 1.19.0
- **Make** >= 4.0

### Production Environment

- **Kubernetes** cluster with 8+ GB RAM, 4+ CPU cores
- **Persistent storage** with ReadWriteMany support
- **LoadBalancer** service support (MetalLB, Cloud LB)
- **Ingress Controller** (NGINX, Traefik)

### Quick Installation (Ubuntu 24.04)

```bash
# Install prerequisites
curl -fsSL https://get.docker.com | sh
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update && sudo apt-get install -y kubectl

# Install KIND for local development
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind

# Install Helm
curl https://get.helm.sh/helm-v3.12.0-linux-amd64.tar.gz | tar xz
sudo mv linux-amd64/helm /usr/local/bin/helm
```

## 🚀 Deployment Options

### Option 1: Complete Demo (Recommended)

Deploy the full O-RAN platform with one command:

```bash
git clone <repository-url>
cd near-rt-ric
make demo
```

**What this deploys:**
- Modern React dashboard with auto-discovery
- Complete Near-RT RIC stack (E2T, A1M, O1M, AppMgr, etc.)
- SMO stack based on ONAP Frankfurt (A&AI, SO, SDNC, Policy)
- Observability pipeline (Grafana, Prometheus, Elasticsearch, Kafka)
- O-RAN network function simulators (O-RU, O-DU, O-CU)

### Option 2: Selective Deployment

Deploy individual components:

```bash
# Deploy only observability stack
make deploy-observability

# Deploy only SMO components
make deploy-smo

# Deploy only Near-RT RIC
make deploy-oran
```

### Option 3: Development Mode

For active development:

```bash
# Install dependencies
make install-deps

# Build all components
make build

# Run tests
make test

# Deploy to local KIND cluster
make deploy

# Setup port forwarding
make port-forward
```

## 🔗 Access Points

After deployment, access the platform via:

| Service | URL | Credentials |
|---------|-----|-------------|
| **O-RAN Dashboard** | http://localhost:3000 | None |
| **Backend API** | http://localhost:8080 | None |
| **Grafana** | http://localhost:3001 | admin/o-ran-admin |
| **Prometheus** | http://localhost:9090 | None |
| **Kibana** | http://localhost:5601 | None |
| **xApp Dashboard** | http://localhost:4200 | None |

## 📊 Monitoring & Observability

### Real-Time Metrics

The platform provides comprehensive real-time monitoring:

- **Network Function Health** - Auto-discovered components with live status
- **Interface Metrics** - E2, A1, O1 message rates and latencies  
- **xApp Performance** - Resource usage, throughput, error rates
- **SMO Operations** - Orchestration workflows, policy enforcement
- **Infrastructure Metrics** - Kubernetes cluster, container resources

### Pre-configured Dashboards

**Grafana Dashboards:**
- O-RAN Near-RT RIC Overview
- E2 Interface Monitoring  
- xApp Performance Analytics
- SMO Operations Dashboard
- Kubernetes Cluster Metrics
- Kafka Streaming Analytics

**Prometheus Alerts:**
- High E2 interface latency (>100ms)
- xApp deployment failures
- SMO component outages
- Resource exhaustion warnings

### Log Analytics

**Elasticsearch + Kibana:**
- Centralized log aggregation from all components
- Real-time log streaming via Kafka
- Advanced search and filtering
- Alerting on error patterns

## 🧪 Testing & Validation

### Automated Testing

```bash
# Run all tests
make test

# End-to-end interface testing
make e2e

# Test specific interfaces
make test-interfaces
make test-smo
make test-observability
```

### Manual Validation

**Health Checks:**
```bash
# Check deployment status
make status

# Test E2 interface
curl http://localhost:8080/api/v1/interfaces/e2/health

# Test A1 interface  
curl http://localhost:8080/api/v1/interfaces/a1/health

# Test O1 interface
curl http://localhost:8080/api/v1/interfaces/o1/health
```

## 🛡️ Security

### Security Features

- **RBAC** - Kubernetes Role-Based Access Control
- **Network Policies** - Micro-segmentation between components
- **Pod Security Standards** - Restricted security contexts
- **Image Scanning** - Trivy vulnerability scanning in CI/CD
- **Secret Management** - Kubernetes secrets for credentials
- **TLS Encryption** - HTTPS endpoints with auto-generated certificates

### Security Scanning

The platform includes automated security scanning:

```bash
# Run security scans
docker run --rm -v $(pwd):/workspace aquasec/trivy fs /workspace

# Check for known vulnerabilities
make security-scan
```

### Compliance

- **O-RAN Security Requirements** - Aligned with O-RAN WG11 specifications
- **NIST Cybersecurity Framework** - Implementation of security controls
- **3GPP Security** - Network function security requirements

## 🔧 Configuration

### Environment Variables

```bash
# Backend API Configuration
export ORAN_NAMESPACE=oran-nearrt-ric
export API_PORT=8080
export LOG_LEVEL=info

# Database Configuration  
export POSTGRES_HOST=postgresql
export POSTGRES_DB=ricplt
export REDIS_HOST=redis

# Observability Configuration
export PROMETHEUS_URL=http://prometheus:9090
export GRAFANA_URL=http://grafana:3000
export ELASTICSEARCH_URL=http://elasticsearch:9200
```

### Helm Value Overrides

```yaml
# values-production.yaml
global:
  registry: your-registry.com
  environment: production

frontend:
  replicas: 3
  resources:
    requests:
      cpu: 500m
      memory: 1Gi

observability:
  grafana:
    persistence:
      size: 50Gi
  prometheus:
    retention: 90d
    storage: 100Gi
```

Deploy with custom values:
```bash
helm upgrade --install oran-nearrt-ric ./helm/oran-nearrt-ric \
  -f values-production.yaml \
  --namespace oran-nearrt-ric
```

## 🚀 CI/CD Pipeline

### GitHub Actions Workflows

The platform includes comprehensive CI/CD automation:

- **Integrated CI** - Build, test, security scan, deploy
- **Multi-environment** - Dev, staging, production deployments  
- **Security Scanning** - Trivy, CodeQL, dependency checks
- **Performance Testing** - Load testing and regression detection
- **Helm Chart Testing** - Lint, validation, install testing

### Pipeline Status

| Workflow | Status | Description |
|----------|--------|-------------|
| **CI/CD Integration** | [![CI](https://github.com/your-org/near-rt-ric/actions/workflows/ci-integrated.yml/badge.svg)](https://github.com/your-org/near-rt-ric/actions) | Build, test, deploy |
| **Security Scanning** | [![Security](https://github.com/your-org/near-rt-ric/actions/workflows/security-complete.yml/badge.svg)](https://github.com/your-org/near-rt-ric/actions) | Vulnerability assessment |
| **Helm Charts** | [![Helm](https://github.com/your-org/near-rt-ric/actions/workflows/helm-test.yml/badge.svg)](https://github.com/your-org/near-rt-ric/actions) | Chart validation |

## 🛠️ Development

### Project Structure

```
near-rt-ric/
├── frontend-dashboard/          # Modern React dashboard
├── dashboard-master/            # Legacy Angular dashboard (deprecated)
├── xAPP_dashboard-master/       # xApp management interface
├── helm/                        # Helm charts
│   ├── oran-nearrt-ric/        # Main O-RAN chart
│   ├── smo-onap/               # SMO ONAP chart
│   └── observability-stack/    # Monitoring stack
├── k8s/                         # Kubernetes manifests
├── docs/                        # Documentation
├── scripts/                     # Deployment scripts
├── .github/workflows/           # CI/CD pipelines
└── Makefile                     # Build automation
```

### Development Workflow

1. **Setup Development Environment**
   ```bash
   make install-deps
   make dev-env
   ```

2. **Make Changes**
   - Frontend: `cd frontend-dashboard && npm run dev`
   - Backend: `cd dashboard-master/dashboard-master && go run src/app/backend/dashboard.go`

3. **Test Changes**
   ```bash
   make test
   make lint
   ```

4. **Deploy Locally**
   ```bash
   make deploy
   make port-forward
   ```

5. **Submit PR**
   - All tests must pass
   - Security scans must be clean
   - Documentation must be updated

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Guidelines

- **Code Style** - Follow language-specific style guides (gofmt, Prettier)
- **Testing** - Maintain >80% test coverage
- **Documentation** - Update docs for all changes
- **Security** - Follow security best practices
- **Performance** - Consider performance impact of changes

### Reporting Issues

Please use GitHub Issues for:
- 🐛 Bug reports with reproduction steps
- 💡 Feature requests with use cases  
- 📚 Documentation improvements
- 🔒 Security vulnerabilities (privately)

## 📚 Documentation

### Additional Resources

- **[Developer Guide](docs/developer/README.md)** - Detailed development instructions
- **[Operations Guide](docs/operations/README.md)** - Production deployment guide
- **[API Reference](docs/developer/api-reference.md)** - REST API documentation
- **[Architecture Deep Dive](docs/ARCHITECTURE.md)** - Technical architecture details
- **[Troubleshooting](docs/operations/troubleshooting.md)** - Common issues and solutions

### O-RAN Standards

- [O-RAN Alliance Specifications](https://www.o-ran.org/specifications)
- [O-RAN Software Community](https://o-ran-sc.org/)
- [3GPP RAN Specifications](https://www.3gpp.org/DynaReport/38-series.htm)

## 🆘 Troubleshooting

### Common Issues

**1. Pod Startup Failures**
```bash
# Check pod status
kubectl get pods -n oran-nearrt-ric

# View pod logs  
kubectl logs -f deployment/oran-nearrt-ric -n oran-nearrt-ric

# Describe pod for events
kubectl describe pod <pod-name> -n oran-nearrt-ric
```

**2. Service Discovery Issues**
```bash
# Test service connectivity
kubectl exec -it <pod-name> -n oran-nearrt-ric -- nslookup kubernetes.default

# Check service endpoints
kubectl get endpoints -n oran-nearrt-ric
```

**3. Interface Connection Problems**
```bash
# Test E2 interface
curl -f http://localhost:8080/api/v1/interfaces/e2/health

# Check routing table
kubectl logs -f deployment/rtmgr -n oran-nearrt-ric
```

**4. Observability Stack Issues**
```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Verify Grafana data sources
curl http://admin:o-ran-admin@localhost:3001/api/datasources
```

### Getting Help

- 📧 **Email**: [support@o-ran-platform.org](mailto:support@o-ran-platform.org)
- 💬 **Slack**: [O-RAN SC Slack](https://o-ran-sc.slack.com)
- 🐛 **Issues**: [GitHub Issues](https://github.com/your-org/near-rt-ric/issues)
- 📖 **Wiki**: [O-RAN SC Wiki](https://wiki.o-ran-sc.org)

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **O-RAN Software Community** for specifications and guidance
- **ONAP Community** for SMO framework components  
- **Cloud Native Computing Foundation** for Kubernetes ecosystem
- **Linux Foundation** for open source governance

---

**🌟 Star this repository if you find it useful!**

**🔄 Contributions, issues, and feature requests are welcome!**

**📢 For updates, follow [@ORanSC](https://twitter.com/ORanSC) on Twitter**