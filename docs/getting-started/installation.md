# Installation Guide

This guide walks you through the process of installing KubeMedic in your Kubernetes cluster.

## Prerequisites

### Required Components
- Kubernetes cluster (v1.16+)
- kubectl configured with admin access
- Metrics Server installed for metrics collection

### Optional Components
- cert-manager (v1.14+) for automated certificate management
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

## Certificate Management

KubeMedic's webhook uses a unified certificate configuration that supports both cert-manager and custom certificates in a single configuration file.

### Option 1: Using cert-manager (Recommended)

1. Install cert-manager if not already installed:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.3/cert-manager.yaml
```

2. Apply the webhook configuration without modifications:
```bash
kubectl apply -f deploy/webhook-cert.yaml
```

The configuration includes:
- A Certificate resource for cert-manager
- A self-signed ClusterIssuer
- Webhook configuration with cert-manager annotations

### Option 2: Using Custom Certificates

1. Edit `deploy/webhook-cert.yaml`:
   - Comment out or remove the cert-manager Certificate and ClusterIssuer sections
   - Uncomment the Secret section and replace placeholders:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubemedic-webhook-cert
  namespace: kubemedic
type: kubernetes.io/tls
data:
  tls.crt: <your-base64-encoded-cert>
  tls.key: <your-base64-encoded-key>
  ca.crt: <your-base64-encoded-ca>
```

2. In the ValidatingWebhookConfiguration section:
   - Remove the cert-manager annotation
   - Uncomment and set the caBundle field

3. Apply your modified configuration:
```bash
kubectl apply -f deploy/webhook-cert.yaml
```

### Automatic Certificate Management

KubeMedic includes two layers of certificate management:

1. **Automatic Renewal**:
   - With cert-manager: Certificates are automatically renewed
   - With custom certificates: You manage renewal externally

2. **Certificate Expiry Monitoring**:
   - Automatic monitoring of certificate expiration
   - Webhook pod restarts when certificates are within 7 days of expiry
   - Configurable cooldown period (default: 1 hour) prevents restart storms

## Verifying the Installation

1. Check if the operator pod is running:
```bash
kubectl get pods -n kubemedic
```

Expected output:
```
NAME                                  READY   STATUS    RESTARTS   AGE
kubemedic-controller-manager-xxxxx    1/1     Running   0          1m
kubemedic-webhook-yyyyy              1/1     Running   0          1m
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

4. Verify webhook configuration:
```bash
kubectl get validatingwebhookconfigurations | grep kubemedic
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

webhook:
  certManager:
    enabled: true  # Set to false if using custom certificates
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

1. Certificate Issues:
```bash
# Check webhook pod logs
kubectl logs -n kubemedic -l app.kubernetes.io/component=webhook

# Check certificate secret
kubectl get secret -n kubemedic kubemedic-webhook-cert
```

2. Webhook Configuration Issues:
```bash
# Check webhook configuration
kubectl get validatingwebhookconfigurations kubemedic-validating-webhook -o yaml

# Check webhook service
kubectl get svc -n kubemedic kubemedic-webhook-service
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