# O-RAN Near-RT RIC Platform

[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes/dashboard)](https://goreportcard.com/report/github.com/kubernetes/dashboard)
[![Coverage Status](https://codecov.io/github/kubernetes/dashboard/coverage.svg?branch=master)](https://codecov.io/github/kubernetes/dashboard?branch=master)
[![GitHub release](https://img.shields.io/github/release/kubernetes/dashboard.svg)](https://github.com/kubernetes/dashboard/releases/latest)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kubernetes/dashboard/blob/master/LICENSE)

A comprehensive O-RAN (Open Radio Access Network) Near Real-Time RAN Intelligent Controller (Near-RT RIC) platform featuring dual Angular dashboards for xApp deployment, lifecycle management, and Kubernetes cluster oversight with sub-second latency requirements.

## Architecture Overview

The Near-RT RIC platform implements O-RAN Alliance specifications with 10ms-1s latency requirements, providing standards-compliant E2, A1, and O1 interface support for multi-vendor telecommunications interoperability.

### Component Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                     O-RAN Near-RT RIC Platform                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────────┐              ┌─────────────────────────────┐  │
│  │   Main Dashboard │◄────────────►│      xApp Dashboard         │  │
│  │                  │              │                             │  │
│  │ • Kubernetes UI  │              │ • xApp Lifecycle Mgmt      │  │
│  │ • Cluster Mgmt   │              │ • Container Orchestration  │  │
│  │ • RBAC Control   │              │ • Image History            │  │
│  │ • Go Backend     │              │ • Angular Frontend         │  │
│  └──────────────────┘              └─────────────────────────────┘  │
│           │                                        │                │
│           └────────────────┬───────────────────────┘                │
│                            │                                        │
│  ┌─────────────────────────▼─────────────────────────────────────┐  │
│  │                    E2 Interface                              │  │
│  │                                                              │  │
│  │ • RAN Node Control & Configuration                          │  │
│  │ • Real-time Telemetry (10ms-1s latency)                    │  │
│  │ • O-RAN Alliance Compliant Messaging                       │  │
│  │ • Multi-vendor Interoperability                            │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

![Dashboard UI workloads page](https://github.com/user-attachments/assets/47a058da-63a8-4140-ae68-592e615c88df)


## Quick Start

### Prerequisites

Ensure the following versions are installed:

*   **Go:** `1.17+`
*   **Node.js:** `16.14.2+` 
*   **Angular CLI:** `13.3.3+`
*   **Kubernetes:** `1.21+`
*   **Docker:** Latest stable
*   **kubectl:** Compatible with your cluster
*   **KIND:** Latest (for local development)

### Local Development Setup

```bash
# 1. Clone the repository
git clone <repository-url>
cd near-rt-ric

# 2. Setup local Kubernetes cluster with KIND
kind create cluster --name near-rt-ric

# 3. Build and start main dashboard
cd dashboard-master/dashboard-master
make build
npm start

# 4. In another terminal, start xApp dashboard  
cd xAPP_dashboard-master
npm install
npm start
```

The main dashboard will be available at `http://localhost:8080` and the xApp dashboard at `http://localhost:4200`.


## Getting Started

**IMPORTANT:** Read the [Access Control](docs/user/access-control/README.md) guide before performing any further steps. The default Dashboard deployment contains a minimal set of RBAC privileges needed to run.

### Deployment

#### Main Dashboard
```bash
cd dashboard-master/dashboard-master
make deploy
```

#### xApp Dashboard
```bash
cd xAPP_dashboard-master  
npm run build
kubectl apply -f k8s/
```

#### Full Platform Deployment
```bash
# Deploy both dashboards
make deploy-all

# Or using individual Makefiles
cd dashboard-master/dashboard-master && make deploy
cd ../../xAPP_dashboard-master && kubectl apply -f k8s/
```

### Access

To access Dashboard from your local workstation you must create a secure channel to your Kubernetes cluster. Run the following command:

```shell
kubectl proxy
```
Now access Dashboard at:

[`http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/`](
http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/).

## Create An Authentication Token (RBAC)
To find out how to create sample user and log in follow [Creating sample user](docs/user/access-control/creating-sample-user.md) guide.

**NOTE:**
* Kubeconfig Authentication method does not support external identity providers or certificate-based authentication.
* [Metrics-Server](https://github.com/kubernetes-sigs/metrics-server) has to be running in the cluster for the metrics and graphs to be available. Read more about it in [Integrations](docs/user/integrations.md) guide.

## O-RAN Alliance Specifications

This platform implements key O-RAN Alliance architectural specifications:

### Architecture References
*   [O-RAN Alliance Official Website](https://www.o-ran.org/)
*   [O-RAN.WG1.O-RAN-Architecture-Description](https://www.o-ran.org/specifications) - Overall architecture
*   [O-RAN.WG3.E2GAP](https://oranalliance.atlassian.net/wiki/spaces/OWG/pages/136413205/O-RAN.WG3.E2GAP-v03.00) - E2 General Aspects and Principles
*   [O-RAN.WG3.E2AP](https://oranalliance.atlassian.net/wiki/spaces/OWG/pages/136413213/O-RAN.WG3.E2AP-v03.00) - E2 Application Protocol
*   [O-RAN.WG3.E2SM](https://oranalliance.atlassian.net/wiki/spaces/OWG/pages/136413221/O-RAN.WG3.E2SM-v03.00) - E2 Service Model
*   [O-RAN.WG2.Use-Cases](https://www.o-ran.org/specifications) - Near-RT RIC use cases and requirements

### Interface Compliance
*   **E2 Interface**: RAN node control and telemetry
*   **A1 Interface**: Policy management from Non-RT RIC
*   **O1 Interface**: Operation and maintenance

## Documentation

Dashboard documentation can be found on [docs](docs/README.md) directory which contains:

* [Common](docs/common/README.md): Entry-level overview.
* [User Guide](docs/user/README.md): [Installation](docs/user/installation.md), [Accessing Dashboard](docs/user/accessing-dashboard/README.md) and more for users.
* [Developer Guide](docs/developer/README.md): [Getting Started](docs/developer/getting-started.md), [Dependency Management](docs/developer/dependency-management.md) and more for anyone interested in contributing.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* [**#sig-ui on Kubernetes Slack**](https://kubernetes.slack.com)
* [**kubernetes-sig-ui mailing list** ](https://groups.google.com/forum/#!forum/kubernetes-sig-ui)
* [**Issue tracker**](https://github.com/kubernetes/dashboard/issues)
* [**SIG info**](https://github.com/kubernetes/community/tree/master/sig-ui)
* [**Roles**](ROLES.md)

### Contribution

Learn how to start contribution on the [Contributing Guideline](CONTRIBUTING.md).

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

## License

[Apache License 2.0](https://github.com/kubernetes/dashboard/blob/master/LICENSE)

----
_Copyright 2019 [The Kubernetes Dashboard Authors](https://github.com/kubernetes/dashboard/graphs/contributors)_
