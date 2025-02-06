# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with KubeMedic.

## Certificate Issues

### Certificate Management Architecture

KubeMedic uses a two-layer approach to certificate management:

1. **Certificate Provisioning**:
   - cert-manager integration (recommended)
   - Custom certificate support
   
2. **Certificate Monitoring**:
   - Automatic expiry monitoring
   - Proactive pod restarts
   - Configurable thresholds and cooldown

### Webhook Certificate Problems

1. **Certificate Renewal Failures**

If the webhook is failing with TLS errors:

```bash
# Check webhook pod logs
kubectl logs -n kubemedic -l app.kubernetes.io/component=webhook

# Verify certificate secret exists and is valid
kubectl get secret -n kubemedic kubemedic-webhook-cert

# Check certificate expiration
kubectl get secret -n kubemedic kubemedic-webhook-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates

# Check certificate renewal policy
kubectl get selfremediationpolicy webhook-cert-renewal -n kubemedic -o yaml
```

2. **cert-manager Issues**

If using cert-manager:

```bash
# Check Certificate resource status
kubectl get certificate -n kubemedic

# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager

# Verify ClusterIssuer is ready
kubectl get clusterissuer selfsigned-issuer -o yaml

# Check Certificate events
kubectl get events -n kubemedic --field-selector involvedObject.kind=Certificate
```

3. **Custom Certificate Issues**

If using custom certificates:

```bash
# Verify CA bundle in webhook configuration
kubectl get validatingwebhookconfigurations kubemedic-validating-webhook -o yaml

# Check webhook service endpoints
kubectl get endpoints -n kubemedic kubemedic-webhook-service

# Test TLS connection (from inside cluster)
kubectl run -it --rm --restart=Never test-connection --image=busybox -- wget --no-check-certificate https://kubemedic-webhook-service.kubemedic.svc:443

# Verify certificate chain
kubectl get secret kubemedic-webhook-cert -n kubemedic -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -text
```

### Certificate Expiry Monitoring

KubeMedic includes automatic certificate expiry monitoring:

1. **Monitoring Configuration**
```yaml
# Check the monitoring policy
kubectl get selfremediationpolicy webhook-cert-renewal -n kubemedic -o yaml
```

Key settings:
- Expiry threshold: 7 days
- Check duration: 1 minute
- Cooldown period: 1 hour

2. **Monitoring Status**
```bash
# Check if monitoring policy is active
kubectl get selfremediationpolicy webhook-cert-renewal -n kubemedic

# View recent pod restarts
kubectl get pods -n kubemedic -l app.kubernetes.io/component=webhook --sort-by=.status.startTime

# Check events related to certificate renewal
kubectl get events -n kubemedic --field-selector reason=CertificateRenewal
```

### Automatic Recovery

KubeMedic includes automatic recovery mechanisms:

1. **Certificate Renewal**
- Webhook pod automatically restarts when certificates are renewed
- Cooldown period prevents restart storms
- Configurable thresholds for proactive renewal

2. **Manual Recovery**

If automatic recovery fails:

```bash
# Restart webhook pod
kubectl rollout restart deployment -n kubemedic kubemedic-webhook

# Force certificate renewal (if using cert-manager)
kubectl delete secret -n kubemedic kubemedic-webhook-cert

# Reset certificate monitoring
kubectl delete selfremediationpolicy webhook-cert-renewal -n kubemedic
kubectl apply -f deploy/webhook-restart-policy.yaml
```

## Common Issues

### 1. Webhook Validation Failures

```bash
# Check webhook logs
kubectl logs -n kubemedic -l app.kubernetes.io/component=webhook

# Verify webhook configuration
kubectl get validatingwebhookconfigurations kubemedic-validating-webhook -o yaml

# Test webhook connectivity
kubectl run test-policy --dry-run=server -o yaml --image=nginx
```

### 2. Policy Not Taking Effect

```bash
# Check policy status
kubectl get selfremediationpolicy -A

# Check controller logs
kubectl logs -n kubemedic deployment/kubemedic-controller-manager

# Verify RBAC permissions
kubectl auth can-i --list --as system:serviceaccount:kubemedic:kubemedic-controller
```

### 3. Metrics Collection Issues

```bash
# Check metrics server is running
kubectl get deployment metrics-server -n kube-system

# Verify metrics are available
kubectl top pods -n kubemedic

# Check controller metrics access
kubectl logs -n kubemedic deployment/kubemedic-controller-manager | grep metrics
```

## Debugging Tools

### 1. Diagnostic Commands

```bash
# Get all KubeMedic resources
kubectl get all -n kubemedic

# Check events
kubectl get events -n kubemedic --sort-by='.lastTimestamp'

# View webhook configuration
kubectl get validatingwebhookconfigurations -o yaml
```

### 2. Log Collection

```bash
# Collect all KubeMedic logs
kubectl logs -n kubemedic -l app.kubernetes.io/name=kubemedic --all-containers

# Get webhook logs
kubectl logs -n kubemedic -l app.kubernetes.io/component=webhook

# Get controller logs
kubectl logs -n kubemedic deployment/kubemedic-controller-manager
```

### 3. Configuration Verification

```bash
# Check CRD installation
kubectl get crd | grep kubemedic

# Verify RBAC setup
kubectl get clusterrole,clusterrolebinding -l app.kubernetes.io/name=kubemedic

# Check service endpoints
kubectl get endpoints -n kubemedic
```

## Getting Help

If you can't resolve the issue:

1. Check existing [GitHub Issues](https://github.com/ikepcampbell/kubemedic/issues)
2. Join our [Community Forum](https://github.com/ikepcampbell/kubemedic/discussions)
3. Create a new issue with:
   - Full error messages
   - Relevant logs
   - Kubernetes version
   - KubeMedic version
   - Steps to reproduce 