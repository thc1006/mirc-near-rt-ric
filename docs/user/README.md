# User Guide

Welcome to the O-RAN Near-RT RIC platform user guide. This guide provides comprehensive information for operators and administrators.

## Getting Started

### Quick Start

1. **Access the Platform**
   - Main Dashboard: `http://localhost:8080`
   - xApp Dashboard: `http://localhost:4200`

2. **First Login**
   - See [Access Control Guide](access-control/README.md) for authentication setup
   - See [Creating Sample User](access-control/creating-sample-user.md) for initial user creation

3. **Basic Navigation**
   - TODO: Add navigation guide based on actual Angular frontend structure

## Dashboard Features

### Main Dashboard (Kubernetes Management)

Located at `dashboard-master/dashboard-master/`, the main dashboard provides:

#### Cluster Overview
- Kubernetes cluster status and metrics
- Node information and resource utilization
- TODO: Add specific feature documentation based on actual frontend components

#### Workload Management  
- Pod lifecycle management
- Deployment configuration
- Service management
- TODO: Extract actual workload management features

#### Resource Monitoring
- Real-time metrics and graphs
- Resource usage tracking
- Performance monitoring
- TODO: Document actual monitoring features

### xApp Dashboard (Application Management)

Located at `xAPP_dashboard-master/`, the xApp dashboard provides:

#### xApp Deployment
- Container-based application deployment
- Image registry management
- TODO: Extract actual xApp deployment features from Angular components

#### Lifecycle Management
- Start, stop, restart xApps
- Health monitoring
- Status tracking
- TODO: Document actual lifecycle management features

#### Configuration Management
- xApp configuration updates
- Environment variable management
- TODO: Extract configuration management features

## xApp Management

### xApp Deployment

TODO: Document xApp deployment procedures based on actual implementation in xAPP_dashboard-master/

### Lifecycle Management

TODO: Document lifecycle management procedures.

### Configuration

TODO: Document configuration management.

## Federated Learning

Based on the federated learning implementation in `dashboard-master/dashboard-master/src/app/backend/federatedlearning/`:

### Client Registration
- Register xApps as federated learning clients
- Configure RRM task capabilities
- Set compute resource limits

### Model Management
- Create and manage global models
- Configure training parameters
- Monitor model performance

### Training Coordination
- Start federated learning jobs
- Monitor training progress
- Review training metrics

TODO: Expand with actual federated learning user workflows.

## Security & Access Control

### RBAC Configuration

Based on existing documentation in `access-control/README.md`:

- Role-based access control setup
- User and service account management
- Permission configuration

### User Management

- Creating and managing users
- Assigning roles and permissions
- Authentication configuration

### Authentication Methods

Available authentication methods:
- Token-based authentication
- Kubeconfig authentication
- Basic authentication (if configured)

TODO: Document actual authentication methods based on auth implementation.

## Network Slice Management

### O-RAN Network Slices

TODO: Document network slice management based on O-RAN implementation.

### Slice Configuration

TODO: Add slice configuration procedures.

### Performance Monitoring

TODO: Document slice performance monitoring.

## Monitoring and Observability

### Metrics Dashboard

TODO: Document metrics dashboard features based on actual frontend implementation.

### Logging

TODO: Document logging features.

### Alerting

TODO: Document alerting configuration.

## Troubleshooting

### Common Issues

For detailed troubleshooting, see [Operations Troubleshooting Guide](../operations/troubleshooting.md).

#### Authentication Issues
- Check RBAC configuration
- Verify service account permissions
- Review authentication logs

#### Performance Issues  
- Check resource utilization
- Review cluster health
- Monitor network connectivity

### FAQ

See [Common FAQ](../common/faq.md) for frequently asked questions.

### Support Resources

- [Issue Tracker](https://github.com/kubernetes/dashboard/issues)
- [Community Slack](https://kubernetes.slack.com)
- [User Documentation](README.md)

## Integration with External Systems

See [Integrations Guide](integrations.md) for information about:
- Prometheus integration
- External authentication providers
- Third-party monitoring systems

---

**Note**: This user guide is based on the actual codebase structure. Specific features and workflows will be documented based on the real Angular frontend implementations found in both dashboard components.