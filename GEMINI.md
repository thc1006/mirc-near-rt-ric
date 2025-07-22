# O-RAN Near-RT RIC Project Context for Gemini CLI

## Project Overview
This project transforms a prototype O-RAN Near-RT RIC into a production-ready, standards-compliant system. The current codebase has significant gaps between claimed functionality and actual implementation.

**Repository**: https://github.com/hctsai1006/near-rt-ric  
**Objective**: Build genuine O-RAN compliant Near Real-Time RAN Intelligent Controller

## Critical Problems Identified

### Missing Core Implementations
- **E2 Interface**: No actual E2AP protocol implementation, only metric definitions
- **A1 Interface**: Missing policy management and ML model services  
- **O1 Interface**: No FCAPS management via NETCONF/YANG
- **Federated Learning**: Only stub interfaces, no real algorithms

### Technical Debt Issues
- Excessive copying from Kubernetes Dashboard without proper customization
- Hardcoded credentials in docker-compose.yml and configs
- Broken CI/CD pipeline with build failures
- Documentation claims functionality that doesn't exist

## Architecture & Technology Stack

### Primary Languages
- **Go 1.19+**: Backend services, E2/A1/O1 interfaces
- **Python 3.9+**: ML components, federated learning
- **TypeScript**: Frontend Angular 15+ applications

### Key Protocols & Standards
- **E2 Interface**: E2AP over SCTP, E2SM-KPM, E2SM-RC service models
- **A1 Interface**: HTTP/2 + JSON, OAuth 2.0 authentication  
- **O1 Interface**: NETCONF/YANG for configuration management
- **Security**: TLS 1.3, RBAC, certificate-based auth

### Infrastructure Components
- **Messaging**: gRPC inter-service, Kafka event streaming
- **Storage**: PostgreSQL (persistent), InfluxDB (metrics), Redis (cache)
- **Orchestration**: Kubernetes with Helm charts
- **Monitoring**: Prometheus metrics, Grafana dashboards

## Development Standards

### Code Quality Requirements
- **Go**: Use gofmt, follow standard project layout, ≥80% test coverage
- **Error Handling**: Comprehensive context propagation, structured logging
- **Documentation**: OpenAPI 3.0 specs, ADRs for design decisions
- **Testing**: TDD approach, testify framework, integration tests

### Project Structure
```
near-rt-ric/
├── pkg/
│   ├── e2/          # E2 interface implementation
│   ├── a1/          # A1 interface implementation  
│   ├── o1/          # O1 interface implementation
│   ├── xapp/        # xApp framework
│   └── common/      # Shared utilities
├── cmd/
│   ├── ricserver/   # Main RIC server
│   └── xappmgr/     # xApp manager
├── deployments/
│   ├── helm/        # Helm charts
│   └── k8s/         # Kubernetes manifests
└── tests/
    ├── unit/        # Unit tests
    ├── integration/ # Integration tests  
    └── e2e/         # End-to-end tests
```

## Performance Requirements
- **E2 Interface**: <10ms P99 latency for control functions
- **A1 Interface**: <1s policy deployment time  
- **System**: Support 100+ concurrent E2 node connections
- **Availability**: >99.9% uptime requirement

## Security Constraints
- **Never commit secrets**: Use Kubernetes Secrets or environment variables
- **TLS everywhere**: All inter-component communication must use TLS 1.3
- **Authentication**: Certificate-based for E2 nodes, OAuth 2.0 for APIs
- **Authorization**: Implement RBAC for all user and xApp permissions
- **Audit**: Log all administrative and configuration changes

## Build & Test Commands
```bash
# Build the project
make build

# Run unit tests  
make test

# Run integration tests
make integration-test

# Lint and format
make lint
make fmt

# Build Docker images
make docker-build

# Deploy to local Kubernetes
make deploy-local
```

## Common Patterns to Follow

### Error Handling
```go
if err != nil {
    return fmt.Errorf("failed to process E2 setup: %w", err)
}
```

### Logging  
```go
log.WithFields(logrus.Fields{
    "nodeID": nodeID,
    "requestID": requestID,
}).Info("E2 setup procedure completed")
```

### Configuration Management
- Use viper for configuration loading
- Support both file and environment variable configuration
- Validate all configuration on startup

## What NOT to Do
- **Never use hardcoded credentials or secrets**
- **Don't copy-paste code without understanding**  
- **Avoid assuming libraries are available - always verify imports**
- **Don't make changes without running tests**
- **Never commit debug prints or temporary files**

## xApp Development Guidelines
- All xApps must use the provided SDK
- Follow the lifecycle management protocols
- Implement proper resource cleanup
- Use structured configuration files
- Include comprehensive health checks

## Monitoring & Observability
- **Metrics**: All components must export Prometheus metrics
- **Tracing**: Use OpenTelemetry for distributed tracing
- **Logging**: Structured JSON logs with correlation IDs
- **Health**: Implement /health and /ready endpoints

## Deployment Notes
- Use Helm for Kubernetes deployments
- Support multiple environments (dev, staging, prod)
- Implement proper secret management
- Include resource limits and requests
- Configure appropriate security contexts

## Integration Testing Requirements  
- Test with simulated RAN nodes
- Verify E2/A1/O1 protocol compliance
- Performance testing under load
- Security penetration testing
- End-to-end workflow validation

Remember: This must be a genuine O-RAN implementation that meets real-world production requirements, not a proof-of-concept or demonstration system.