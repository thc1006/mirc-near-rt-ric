# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an O-RAN Near Real-Time RAN Intelligent Controller (Near-RT RIC) platform consisting of two main Angular dashboards and supporting Kubernetes infrastructure for xApp deployment and management.

## Build Commands

### Main Dashboard (dashboard-master/dashboard-master)
```bash
# Build the project
make build

# Start development server with backend
npm start

# Start with HTTPS
npm run start:https

# Run tests
npm test
make test

# Code quality checks
make check
npm run fix

# Deploy to Kubernetes
make deploy
```

### xApp Dashboard (xAPP_dashboard-master)
```bash
# Install dependencies
npm install

# Start development server
npm start

# Build for production
npm run build

# Run tests
npm test
```

## Architecture

The codebase consists of two primary components:

### 1. Main Dashboard (dashboard-master/dashboard-master)
- **Backend**: Go-based Kubernetes Dashboard API server (`src/app/backend/`)
- **Frontend**: Angular 13.3.x application (`src/app/frontend/`)
- **Purpose**: Kubernetes cluster management and Near-RT RIC platform oversight
- **Build System**: Make-based with npm script integration

### 2. xApp Dashboard (xAPP_dashboard-master)
- **Technology**: Pure Angular 13.3.x application
- **Purpose**: xApp lifecycle management and monitoring
- **Components**: Front-page, Home, Tags, Image History management

## Key Development Workflows

### Local Development Setup
1. **Prerequisites**: Docker, kubectl, Go 1.17+, Node.js 16.14.2+, Angular CLI 13.3.3, KIND
2. **Kubernetes Cluster**: Use KIND for local development
3. **Backend Development**: Uses Make commands with live reload via `make watch-backend`
4. **Frontend Development**: Standard Angular CLI with proxy configuration

### Testing
- **Backend**: Go tests via `make test`
- **Frontend**: Jest/Karma via `npm test`
- **E2E**: Cypress tests via `npm run e2e`

### Code Quality
- **Linting**: ESLint, Stylelint, golangci-lint
- **Formatting**: gts (Google TypeScript Style), Go formatting
- **Pre-commit**: Husky with lint-staged

## O-RAN Domain Context

This platform implements O-RAN Alliance specifications for Near-RT RIC functionality:
- **Latency Requirements**: 10ms-1s response times
- **Standards Compliance**: E2, A1, O1 interface support
- **xApp Management**: Container-based application lifecycle
- **Multi-vendor Interoperability**: Standards-based telecommunications protocols

## Development Environment

### Backend Configuration
- RBAC-enabled Kubernetes API access
- Service account authentication
- Configurable endpoints via environment variables

### Frontend Configuration
- Proxy configuration for backend API (`aio/proxy.conf.json`)
- HTTPS support with auto-generated certificates
- Material Design UI framework
- Chart.js for metrics visualization

## Common Issues

### RBAC Errors
Create appropriate ClusterRoleBindings for ServiceAccounts when encountering permission errors.

### Image Pull Issues
Configure KIND with local registry trust when using custom container images.

### Build Failures
Ensure Node.js >= 16.14.2 and Go >= 1.17 are installed and properly configured.