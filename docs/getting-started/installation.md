# Installation Guide

This guide walks you through the process of installing KubeMedic in your Kubernetes cluster.

## Prerequisites

### Required Components
- Kubernetes cluster (v1.16+)
- kubectl configured with admin access
- Prometheus installed for metrics collection

### Optional Components
- Grafana (v8.0+) for visualization
- Helm (v3+) for package management
- Argo CD for GitOps workflows

## Installation Methods

### Method 1: Direct Installation (Recommended for Testing)

```bash
# Apply the CRDs
kubectl apply -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/config/crd/bases/remediation.kubemedic.io_selfremediationpolicies.yaml

# Install the operator
kubectl apply -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/config/deploy/kubemedic.yaml
```

### Method 2: Helm Installation (Recommended for Production)

```bash
# Add the KubeMedic Helm repository
helm repo add kubemedic https://ikepcampbell.github.io/kubemedic/charts

# Update your repositories
helm repo update

# Install KubeMedic
helm install kubemedic kubemedic/kubemedic \
  --namespace kubemedic \
  --create-namespace
```

### Method 3: Manual Installation

```bash
# Clone the repository
git clone https://github.com/ikepcampbell/kubemedic.git
cd kubemedic

# Install CRDs
make install

# Deploy the operator
make deploy IMG=ghcr.io/ikepcampbell/kubemedic:latest
```

## Verifying the Installation

1. Check if the operator pod is running:
```bash
kubectl get pods -n kubemedic
```

Expected output:
```
NAME                                  READY   STATUS    RESTARTS   AGE
kubemedic-controller-manager-xxxxx    1/1     Running   0          1m
```

2. Verify CRD installation:
```bash
kubectl get crds | grep kubemedic
```

Expected output:
```
selfremediationpolicies.remediation.kubemedic.io   2024-02-05T12:00:00Z
```

3. Check operator logs:
```bash
kubectl logs -n kubemedic deployment/kubemedic-controller-manager
```

## Configuration

### Basic Configuration
Create a basic configuration file (`values.yaml`):

```yaml
prometheus:
  url: http://prometheus-server:9090

grafana:
  enabled: true
  url: http://grafana:3000

operator:
  logLevel: info
  metricsPort: 8080
```

### Applying Configuration

If using Helm:
```bash
helm upgrade kubemedic kubemedic/kubemedic \
  -f values.yaml \
  --namespace kubemedic
```

## Next Steps

1. Create your first remediation policy
2. Set up monitoring integrations
3. Configure webhooks (if needed)
4. Review the [Quickstart Guide](quickstart.md)

## Troubleshooting Installation

### Common Issues

1. CRDs not installing:
```bash
kubectl apply -f config/crd/bases/ --force
```

2. Operator pod not starting:
```bash
kubectl describe pod -n kubemedic kubemedic-controller-manager-xxxxx
```

3. RBAC issues:
```bash
kubectl describe clusterrole kubemedic-manager-role
kubectl describe clusterrolebinding kubemedic-manager-rolebinding
```

### Getting Help

If you encounter any issues:
1. Check our [Troubleshooting Guide](../reference/troubleshooting.md)
2. Search existing [GitHub Issues](https://github.com/ikepcampbell/kubemedic/issues)
3. Join our [Community Forum](https://github.com/ikepcampbell/kubemedic/discussions)

## Uninstallation

### Using kubectl:
```bash
kubectl delete -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/config/deploy/kubemedic.yaml
kubectl delete -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/config/crd/bases/remediation.kubemedic.io_selfremediationpolicies.yaml
```

### Using Helm:
```bash
helm uninstall kubemedic -n kubemedic
``` 