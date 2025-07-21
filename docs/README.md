# O-RAN Near-RT RIC Documentation

Welcome to the comprehensive documentation for the O-RAN Near Real-Time RAN Intelligent Controller (Near-RT RIC) platform.

## 🚀 Quick Start

New to the platform? Start here:
- [Installation Guide](../dashboard-master/dashboard-master/docs/user/installation.md) - Get up and running quickly
- [User Guide](user/README.md) - Essential user workflows
- [Developer Guide](../dashboard-master/dashboard-master/docs/developer/README.md) - Development environment setup

## 📚 Documentation Structure

### 👥 User Documentation
For operators and administrators using the platform:

- **[User Guide](user/README.md)** - Complete user manual
- **[Installation Guide](../dashboard-master/dashboard-master/docs/user/installation.md)** - Step-by-step setup
- **[Access Control](../dashboard-master/dashboard-master/docs/user/access-control/README.md)** - RBAC and security
- **[Integrations](../dashboard-master/dashboard-master/docs/user/integrations.md)** - Third-party integrations

### 🛠️ Developer Documentation
For developers extending and customizing the platform:

- **[Developer Guide](../dashboard-master/dashboard-master/docs/developer/README.md)** - Development workflows
- **[Getting Started](../dashboard-master/dashboard-master/docs/developer/getting-started.md)** - Environment setup
- **[API Reference](developer/api-reference.md)** - REST API documentation
- **[Architecture](../dashboard-master/dashboard-master/docs/developer/architecture.md)** - System design
- **[Dependency Management](../dashboard-master/dashboard-master/docs/developer/dependency-management.md)** - Package management

### 🔧 Operations Documentation
For production deployment and maintenance:

- **[Deployment Guide](operations/deployment.md)** - Production deployment
- **[Monitoring & Logging](operations/monitoring.md)** - Observability setup
- **[Troubleshooting](operations/troubleshooting.md)** - Issue resolution
- **[Performance Tuning](operations/performance.md)** - Optimization

### 📋 Common Resources
Shared information and concepts:

- **[Common Guide](../dashboard-master/dashboard-master/docs/common/README.md)** - Overview and concepts
- **[FAQ](../dashboard-master/dashboard-master/docs/common/faq.md)** - Frequently asked questions
- **[Dashboard Arguments](../dashboard-master/dashboard-master/docs/common/dashboard-arguments.md)** - Command-line options

## 🏗️ Platform Architecture

The Near-RT RIC platform consists of two main components:

### Main Dashboard (Kubernetes Management)
- **Location**: `dashboard-master/dashboard-master/`
- **Backend**: Go-based Kubernetes Dashboard API
- **Frontend**: Angular 13.3+ application
- **Purpose**: Kubernetes cluster management and platform oversight

### xApp Dashboard (Application Management)  
- **Location**: `xAPP_dashboard-master/`
- **Technology**: Angular 13.3+ application
- **Purpose**: xApp lifecycle management and monitoring

## 🌐 O-RAN Standards Compliance

This platform implements O-RAN Alliance specifications:

- **E2 Interface**: Real-time RAN control (10ms-1s latency)
- **A1 Interface**: Policy and ML model management  
- **O1 Interface**: Operations and maintenance
- **Federated Learning**: Privacy-preserving intelligent RRM

## 📖 Additional Resources

- [Release Procedures](../dashboard-master/dashboard-master/docs/developer/release-procedures.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
- [Code of Conduct](../dashboard-master/dashboard-master/code-of-conduct.md)
- [License](../dashboard-master/dashboard-master/LICENSE)

## 🆘 Getting Help

- [Troubleshooting Guide](operations/troubleshooting.md)
- [O-RAN SC Documentation](https://docs.o-ran-sc.org/)
- [O-RAN SC Projects](https://docs.o-ran-sc.org/en/latest/projects.html)
- [Issue Tracker](https://github.com/hctsai1006/near-rt-ric/issues)

---

**Note**: This documentation reflects the actual codebase structure. Links point to existing or newly created documentation files based on the real implementation.