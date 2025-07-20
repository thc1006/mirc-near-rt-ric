# Production Deployment Guide

This comprehensive guide covers deploying the O-RAN Near-RT RIC platform in production environments using cloud-native best practices.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Deployment Methods](#deployment-methods)
3. [Configuration](#configuration)
4. [High Availability Setup](#high-availability-setup)
5. [Security Configuration](#security-configuration)
6. [Monitoring & Observability](#monitoring--observability)
7. [Post-Deployment Verification](#post-deployment-verification)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

**Minimum Production Requirements:**
- **Kubernetes**: 1.29+ (with containerd runtime)
- **Nodes**: 3+ worker nodes, 3+ control plane nodes
- **CPU**: 8+ cores per worker node, 4+ cores per control plane node
- **Memory**: 32GB+ per worker node, 16GB+ per control plane node
- **Storage**: 500GB+ NVMe SSD per node, separate storage for persistent volumes
- **Network**: 10Gbps+ backbone, 1Gbps+ per node, low latency (<1ms RTT within cluster)

**O-RAN Specific Requirements:**
- **E2 Interface Latency**: <10ms between E2 nodes and Near-RT RIC
- **A1 Interface**: REST API support with TLS 1.3
- **O1 Interface**: NETCONF/YANG support with SSH key-based authentication
- **RRM Processing**: GPU acceleration for ML workloads (NVIDIA T4+ recommended)

### Infrastructure Tools

```bash
# Required versions (minimum)
kubectl >= 1.29.0
helm >= 3.13.0
docker >= 24.0.0
istio >= 1.20.0  # For service mesh
cert-manager >= 1.13.0  # For certificate management
prometheus-operator >= 0.70.0  # For monitoring
```

### Kubernetes Cluster Setup

#### Multi-zone Production Cluster

```yaml
# Example cluster configuration for cloud providers
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: oran-nearrt-ric-prod
spec:
  clusterNetwork:
    pods:
      cidrBlocks: ["10.244.0.0/16"]
    services:
      cidrBlocks: ["10.96.0.0/16"]
  infrastructure:
    kind: AWSCluster  # or GCPCluster, AzureCluster
    name: oran-prod-infra
  controlPlane:
    replicas: 3
    version: v1.29.0
  workers:
    replicas: 6  # Spread across 3 zones
    version: v1.29.0
```

#### Node Configuration

```bash
# Configure nodes with O-RAN specific labels and taints
kubectl label nodes worker-1 node-type=oran-worker
kubectl label nodes worker-2 node-type=monitoring-worker  
kubectl label nodes worker-3 node-type=storage-worker

# Taint GPU nodes for ML workloads
kubectl taint nodes gpu-worker-1 nvidia.com/gpu=true:NoSchedule
kubectl taint nodes gpu-worker-2 nvidia.com/gpu=true:NoSchedule
```

## Deployment Methods

### Method 1: Helm-based Production Deployment (Recommended)

```bash
# 1. Add required Helm repositories
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add jetstack https://charts.jetstack.io
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update

# 2. Install prerequisite components
# Install cert-manager for TLS certificate management
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.13.0 \
  --set installCRDs=true

# Install Istio service mesh
helm install istio-base istio/base \
  --namespace istio-system \
  --create-namespace

helm install istiod istio/istiod \
  --namespace istio-system \
  --wait

# Install Istio gateway
helm install istio-gateway istio/gateway \
  --namespace istio-gateway \
  --create-namespace

# 3. Deploy O-RAN Near-RT RIC platform
helm install oran-nearrt-ric ./helm/oran-nearrt-ric \
  --namespace oran-nearrt-ric \
  --create-namespace \
  --values ./helm/oran-nearrt-ric/values-production.yaml \
  --wait --timeout=30m

# 4. Verify deployment
kubectl get all -n oran-nearrt-ric
helm status oran-nearrt-ric -n oran-nearrt-ric
```

### Method 2: GitOps with ArgoCD

```bash
# 1. Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 2. Create ArgoCD Application
cat <<EOF | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: oran-nearrt-ric
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/near-rt-ric
    targetRevision: main
    path: helm/oran-nearrt-ric
    helm:
      valueFiles:
      - values-production.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: oran-nearrt-ric
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
EOF

# 3. Access ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

### Method 3: Kustomize-based Deployment

```bash
# 1. Deploy using Kustomize overlays
kubectl apply -k k8s/overlays/production

# 2. Verify deployment
kubectl get all -n oran-nearrt-ric
```

## Configuration

### Production Values Configuration

Create `values-production.yaml` for Helm deployment:

```yaml
# Production configuration for O-RAN Near-RT RIC
global:
  imageRegistry: "your-registry.com"
  environment: "production"
  oran:
    plmnId: "310410"  # Update with your PLMN ID
    domain: "oran.yourdomain.com"

# Main Dashboard configuration
mainDashboard:
  enabled: true
  replicaCount: 3
  image:
    tag: "v1.0.0"
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi
  ingress:
    enabled: true
    className: "istio"
    annotations:
      kubernetes.io/ingress.class: "istio"
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
    hosts:
    - host: dashboard.oran.yourdomain.com
      paths:
      - path: /
        pathType: Prefix
    tls:
    - secretName: dashboard-tls
      hosts:
      - dashboard.oran.yourdomain.com
  persistence:
    enabled: true
    storageClass: "fast-ssd"
    size: 10Gi

# xApp Dashboard configuration
xappDashboard:
  enabled: true
  replicaCount: 3
  ingress:
    enabled: true
    hosts:
    - host: xapps.oran.yourdomain.com

# Federated Learning configuration
federatedLearning:
  enabled: true
  replicaCount: 2
  resources:
    requests:
      cpu: 1000m
      memory: 4Gi
      nvidia.com/gpu: 1
    limits:
      cpu: 4000m
      memory: 16Gi
      nvidia.com/gpu: 2
  persistence:
    enabled: true
    storageClass: "fast-ssd"
    size: 100Gi

# O-RAN Components
e2Simulator:
  enabled: true
  replicaCount: 2
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi

# Monitoring configuration
monitoring:
  prometheus:
    enabled: true
    server:
      retention: "30d"
      persistentVolume:
        enabled: true
        size: 100Gi
        storageClass: "fast-ssd"
      resources:
        requests:
          cpu: 1000m
          memory: 4Gi
        limits:
          cpu: 4000m
          memory: 16Gi
  
  grafana:
    enabled: true
    adminPassword: "your-secure-password"
    persistence:
      enabled: true
      size: 10Gi
      storageClass: "fast-ssd"
    ingress:
      enabled: true
      hosts:
      - grafana.oran.yourdomain.com

# Database configuration
postgresql:
  enabled: true
  auth:
    database: "oran_nearrt_ric"
    username: "oran"
    password: "your-secure-db-password"
  primary:
    persistence:
      enabled: true
      size: 100Gi
      storageClass: "fast-ssd"
    resources:
      requests:
        cpu: 1000m
        memory: 4Gi
      limits:
        cpu: 4000m
        memory: 16Gi

# Redis configuration
redis:
  enabled: true
  auth:
    enabled: true
    password: "your-secure-redis-password"
  master:
    persistence:
      enabled: true
      size: 20Gi
      storageClass: "fast-ssd"
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        cpu: 2000m
        memory: 8Gi

# Security configuration
rbac:
  create: true

networkPolicies:
  enabled: true

podSecurityPolicy:
  enabled: true

# Autoscaling configuration
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

### Environment-Specific Configurations

#### SSL/TLS Configuration

```yaml
# ClusterIssuer for Let's Encrypt
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@yourdomain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: istio
```

#### Resource Quotas

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: oran-quota
  namespace: oran-nearrt-ric
spec:
  hard:
    requests.cpu: "50"
    requests.memory: 200Gi
    limits.cpu: "100"
    limits.memory: 400Gi
    persistentvolumeclaims: "20"
    pods: "100"
    services: "20"
```

## High Availability Setup

### Multi-zone Deployment

```yaml
# Pod Anti-affinity for high availability
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app.kubernetes.io/name
          operator: In
          values:
          - main-dashboard
      topologyKey: topology.kubernetes.io/zone
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - main-dashboard
        topologyKey: kubernetes.io/hostname
```

### Load Balancing

#### External Load Balancer

```yaml
apiVersion: v1
kind: Service
metadata:
  name: oran-ingress-lb
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
    service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: "true"
spec:
  type: LoadBalancer
  ports:
  - port: 443
    targetPort: 8443
    protocol: TCP
  - port: 80
    targetPort: 8080
    protocol: TCP
  selector:
    app: istio-gateway
```

### Backup and Recovery

#### Database Backup

```bash
# PostgreSQL backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
NAMESPACE="oran-nearrt-ric"
BACKUP_PATH="/backups/postgresql"

# Create backup
kubectl exec -n $NAMESPACE deployment/postgresql -- \
  pg_dump -U oran oran_nearrt_ric | \
  gzip > $BACKUP_PATH/oran_db_backup_$DATE.sql.gz

# Upload to cloud storage (example for AWS S3)
aws s3 cp $BACKUP_PATH/oran_db_backup_$DATE.sql.gz \
  s3://oran-backups/postgresql/

# Cleanup old backups (keep last 30 days)
find $BACKUP_PATH -name "oran_db_backup_*.sql.gz" -mtime +30 -delete
```

#### Persistent Volume Backup

```yaml
# VolumeSnapshot for backup
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: oran-pv-snapshot
  namespace: oran-nearrt-ric
spec:
  volumeSnapshotClassName: csi-hostpath-snapclass
  source:
    persistentVolumeClaimName: data-postgresql-0
```

## Security Configuration

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: oran-network-policy
  namespace: oran-nearrt-ric
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  # Allow traffic from istio-system
  - from:
    - namespaceSelector:
        matchLabels:
          name: istio-system
  # Allow internal namespace traffic
  - from:
    - namespaceSelector:
        matchLabels:
          name: oran-nearrt-ric
  # Allow monitoring
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 9090
  egress:
  # Allow DNS resolution
  - to: []
    ports:
    - protocol: UDP
      port: 53
  # Allow HTTPS outbound
  - to: []
    ports:
    - protocol: TCP
      port: 443
  # Allow internal traffic
  - to:
    - namespaceSelector:
        matchLabels:
          name: oran-nearrt-ric
```

### Pod Security Standards

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: oran-nearrt-ric
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### RBAC Configuration

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: oran-operator
rules:
# O-RAN specific resources
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "daemonsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["networkpolicies", "ingresses"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# Monitoring resources
- apiGroups: ["monitoring.coreos.com"]
  resources: ["servicemonitors", "prometheusrules"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

## Monitoring & Observability

### Prometheus Configuration

```yaml
# ServiceMonitor for scraping metrics
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: oran-platform
  namespace: oran-nearrt-ric
spec:
  selector:
    matchLabels:
      app.kubernetes.io/part-of: oran-nearrt-ric
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

### Grafana Dashboards

Import pre-built dashboards:
- O-RAN Platform Overview: [dashboard JSON](../config/grafana/dashboard-files/oran-overview.json)
- E2 Interface Monitoring: [dashboard JSON](../config/grafana/dashboard-files/e2-interface.json)
- xApp Performance: [dashboard JSON](../config/grafana/dashboard-files/xapp-performance.json)

### Logging with Loki

```yaml
# Loki configuration for log aggregation
apiVersion: v1
kind: ConfigMap
metadata:
  name: loki-config
data:
  loki.yaml: |
    auth_enabled: false
    server:
      http_listen_port: 3100
    ingester:
      lifecycler:
        address: 127.0.0.1
        ring:
          kvstore:
            store: inmemory
          replication_factor: 1
    schema_config:
      configs:
        - from: 2020-10-24
          store: boltdb-shipper
          object_store: filesystem
          schema: v11
          index:
            prefix: index_
            period: 24h
```

## Post-Deployment Verification

### Health Checks

```bash
#!/bin/bash
# Comprehensive health check script

echo "=== O-RAN Near-RT RIC Health Check ==="

# Check namespace
kubectl get ns oran-nearrt-ric || exit 1

# Check deployments
kubectl get deployments -n oran-nearrt-ric
kubectl wait --for=condition=available deployment --all -n oran-nearrt-ric --timeout=300s

# Check services
kubectl get services -n oran-nearrt-ric

# Check ingress
kubectl get ingress -n oran-nearrt-ric

# Test E2 Interface
echo "Testing E2 Interface..."
kubectl exec -n oran-nearrt-ric deployment/e2-simulator -- \
  nc -z e2-termination 38000 && echo "E2 Interface: OK" || echo "E2 Interface: FAILED"

# Test A1 Interface
echo "Testing A1 Interface..."
kubectl exec -n oran-nearrt-ric deployment/a1-mediator -- \
  curl -f http://localhost:10020/a1-p/healthcheck && echo "A1 Interface: OK" || echo "A1 Interface: FAILED"

# Test Dashboard endpoints
echo "Testing Dashboard endpoints..."
kubectl port-forward -n oran-nearrt-ric service/main-dashboard 8080:8080 &
PF_PID=$!
sleep 5
curl -k -f https://localhost:8080/api/v1/login/status && echo "Main Dashboard: OK" || echo "Main Dashboard: FAILED"
kill $PF_PID

# Test monitoring
echo "Testing monitoring stack..."
kubectl port-forward -n monitoring service/prometheus-server 9090:80 &
PF_PID=$!
sleep 5
curl -f http://localhost:9090/-/healthy && echo "Prometheus: OK" || echo "Prometheus: FAILED"
kill $PF_PID

echo "=== Health Check Complete ==="
```

### Performance Validation

```bash
# Load testing with k6
k6 run - <<EOF
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 10 },
    { duration: '5m', target: 50 },
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.1'],
  },
};

export default function () {
  let response = http.get('https://dashboard.oran.yourdomain.com/api/v1/login/status');
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
EOF
```

### Functional Testing

```bash
# E2E functional test script
#!/bin/bash

echo "=== Functional Testing ==="

# Test xApp deployment
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-xapp
  namespace: oran-nearrt-ric
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-xapp
  template:
    metadata:
      labels:
        app: test-xapp
    spec:
      containers:
      - name: test-xapp
        image: busybox:1.35
        command: ['sleep', '3600']
EOF

# Wait for xApp deployment
kubectl wait --for=condition=available deployment/test-xapp -n oran-nearrt-ric --timeout=60s

# Test xApp registration with FL Coordinator
kubectl exec -n oran-nearrt-ric deployment/test-xapp -- \
  wget -qO- http://fl-coordinator:8080/fl/clients

# Cleanup test xApp
kubectl delete deployment test-xapp -n oran-nearrt-ric

echo "=== Functional Testing Complete ==="
```

## Troubleshooting

### Common Issues

#### Pod Startup Failures
```bash
# Debug pod issues
kubectl describe pod <pod-name> -n oran-nearrt-ric
kubectl logs <pod-name> -n oran-nearrt-ric --previous
kubectl get events -n oran-nearrt-ric --sort-by='.lastTimestamp'
```

#### Network Connectivity Issues
```bash
# Test network connectivity
kubectl exec -it <pod-name> -n oran-nearrt-ric -- nslookup e2-termination
kubectl exec -it <pod-name> -n oran-nearrt-ric -- nc -zv a1-mediator 10020
```

#### Resource Constraints
```bash
# Check resource usage
kubectl top nodes
kubectl top pods -n oran-nearrt-ric
kubectl describe resourcequota -n oran-nearrt-ric
```

### Diagnostic Commands

```bash
# Collect comprehensive diagnostic information
kubectl cluster-info dump --output-directory=/tmp/cluster-dump
kubectl get all -A -o yaml > /tmp/all-resources.yaml
kubectl describe nodes > /tmp/node-status.txt
```

### Recovery Procedures

#### Rolling Back Deployment
```bash
# Rollback to previous version
helm rollback oran-nearrt-ric -n oran-nearrt-ric

# Check rollback status
helm history oran-nearrt-ric -n oran-nearrt-ric
```

#### Database Recovery
```bash
# Restore from backup
kubectl exec -n oran-nearrt-ric deployment/postgresql -- \
  dropdb -U oran oran_nearrt_ric
kubectl exec -n oran-nearrt-ric deployment/postgresql -- \
  createdb -U oran oran_nearrt_ric
gunzip -c /backups/postgresql/oran_db_backup_20240120_120000.sql.gz | \
  kubectl exec -i -n oran-nearrt-ric deployment/postgresql -- \
  psql -U oran oran_nearrt_ric
```

---

**Status**: Production-ready deployment guide with comprehensive configuration examples, security best practices, and operational procedures for O-RAN Near-RT RIC platform.