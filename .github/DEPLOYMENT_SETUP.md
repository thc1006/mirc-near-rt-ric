# GitHub Actions Deployment Setup Guide

This guide explains how to configure the required GitHub Actions secrets and deployment settings for the O-RAN Near-RT RIC CI/CD pipeline.

## Required GitHub Secrets

Navigate to your repository Settings → Secrets and variables → Actions to add these secrets:

### 1. Code Coverage
```
CODECOV_TOKEN
```
- **Purpose**: Upload code coverage reports to Codecov
- **How to get**: 
  1. Visit [codecov.io](https://codecov.io)
  2. Sign in with GitHub and add your repository
  3. Copy the repository token from the settings page
- **Format**: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

### 2. Kubernetes Deployment
```
KUBE_CONFIG
```
- **Purpose**: Kubectl configuration for production deployment
- **How to get**: 
  1. Get your Kubernetes cluster config file (`~/.kube/config`)
  2. Base64 encode it: `cat ~/.kube/config | base64 -w 0`
- **Format**: Base64-encoded kubeconfig file
- **Security**: Ensure the service account has only necessary permissions

### 3. Slack Notifications (Optional)
```
SLACK_WEBHOOK
```
- **Purpose**: Send deployment notifications to Slack
- **How to get**:
  1. Create a Slack app and incoming webhook
  2. Copy the webhook URL
- **Format**: `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`

## Security Best Practices

### 1. Kubernetes RBAC Configuration

Create a dedicated service account for CI/CD deployments:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: github-actions-deployer
  namespace: oran-nearrt-ric
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: github-actions-deployer
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "patch", "update"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: github-actions-deployer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: github-actions-deployer
subjects:
- kind: ServiceAccount
  name: github-actions-deployer
  namespace: oran-nearrt-ric
```

### 2. Environment Protection Rules

Configure environment protection rules in GitHub:

1. Go to Settings → Environments
2. Create a "production" environment
3. Add protection rules:
   - Required reviewers
   - Wait timer
   - Deployment branches (main only)

### 3. Secret Rotation

- Rotate `CODECOV_TOKEN` every 90 days
- Rotate `KUBE_CONFIG` when team members change
- Monitor secret usage in GitHub Actions logs

## Container Registry Configuration

The pipeline uses GitHub Container Registry (ghcr.io) by default:

1. **Permissions**: Ensure GitHub Actions has permission to push to packages
2. **Visibility**: Set package visibility to private for production
3. **Cleanup**: Configure automatic cleanup of old images

## Monitoring and Alerts

### 1. Pipeline Monitoring

Monitor pipeline health with:
- Failed build notifications
- Security scan alerts
- Performance regression alerts

### 2. Application Monitoring

Set up monitoring for:
- Container health checks
- Resource usage
- Application logs
- Security vulnerabilities

## Troubleshooting

### Common Issues

1. **RBAC Permission Denied**
   - Check service account permissions
   - Verify cluster role bindings
   - Ensure namespace exists

2. **Image Pull Errors**
   - Verify registry credentials
   - Check image tags and availability
   - Confirm network connectivity

3. **Codecov Upload Failures**
   - Verify token validity
   - Check coverage file format
   - Ensure repository is configured in Codecov

### Debug Commands

```bash
# Check service account permissions
kubectl auth can-i --list --as=system:serviceaccount:oran-nearrt-ric:github-actions-deployer

# Verify deployment status
kubectl get deployments -n oran-nearrt-ric
kubectl describe deployment dashboard -n oran-nearrt-ric

# Check pod logs
kubectl logs -l app=dashboard -n oran-nearrt-ric --tail=100
```

## Additional Security Considerations

1. **Network Policies**: Implement Kubernetes network policies
2. **Pod Security Standards**: Use restricted pod security standards
3. **Image Scanning**: Enable automatic vulnerability scanning
4. **Secret Management**: Consider using external secret management (e.g., HashiCorp Vault)
5. **Audit Logging**: Enable Kubernetes audit logging for compliance

## Support

For issues with the CI/CD pipeline:
1. Check GitHub Actions logs
2. Review this documentation
3. Check Kubernetes cluster status
4. Contact the DevOps team