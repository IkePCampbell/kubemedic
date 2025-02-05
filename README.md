# KubeMedic - Safe Kubernetes Auto-Remediation

KubeMedic is a Kubernetes operator that safely automates common remediation tasks while protecting your cluster from unintended consequences.

## Prerequisites

### Required
- Kubernetes cluster (v1.16+)
- [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) installed in your cluster
  ```bash
  kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
  ```

### Optional
- Prometheus & Grafana for advanced metrics visualization
  - KubeMedic works with the native Kubernetes metrics API by default
  - Can be integrated with Prometheus for historical data and advanced querying
  - Grafana dashboards available for visualization

## Key Features

### 🛡️ Safe by Default
- Protected system namespaces (kube-system, etc.)
- Resource quotas and scaling limits
- Automatic state backups before actions
- Gradual scaling with automatic revert

### 🎯 Common Remediations
- CPU/Memory-based scaling
- Pod restart on high error rates
- HPA limit adjustments
- Temporary resource overrides

### 🔒 Built-in Safeguards
- Maximum scale factor (2x by default)
- Rate limiting and cooldown periods
- Resource quota validation
- Protected resources via labels

## Quick Start

1. **Install Metrics Server** (if not already installed)
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

2. **Install KubeMedic**
```bash
kubectl apply -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/config/deploy/kubemedic.yaml
```

3. **Create a Simple Policy**
```yaml
apiVersion: remediation.kubemedic.io/v1alpha1
kind: SelfRemediationPolicy
metadata:
  name: cpu-scaling
  namespace: my-app
spec:
  rules:
    - name: high-cpu-scale
      conditions:
        - type: PodCPUUsage    # Uses metrics-server directly
          threshold: "80"      # 80% CPU usage
          duration: "5m"
      actions:
        - type: ScaleUp
          target:
            kind: Deployment
            name: my-service
          scalingParams:
            temporaryMaxReplicas: 5
            scalingDuration: "30m"
            revertStrategy: "Gradual"
```

## Monitoring Options

### 1. Basic Monitoring (Default)
- Uses Kubernetes metrics API directly
- Real-time metrics without historical data
- View with kubectl:
  ```bash
  kubectl top pods
  kubectl get pods
  kubectl describe selfremediationpolicy
  ```

### 2. Advanced Monitoring (Optional)
#### With Prometheus
```yaml
# values.yaml
monitoring:
  prometheus:
    enabled: true
    serviceMonitor:
      enabled: true    # If using prometheus-operator
    rules:
      enabled: true    # Install default alerting rules
```

#### With Grafana
```yaml
monitoring:
  grafana:
    enabled: true
    dashboards:
      enabled: true    # Install default dashboards
```

## Testing KubeMedic

KubeMedic comes with comprehensive examples that include both policies and test applications. Each example is self-contained and includes step-by-step testing instructions.

### Available Examples

1. **CPU Scaling Test**
```bash
# Apply the CPU scaling example
kubectl apply -f examples/cpu-scaling-with-test.yaml

# Follow the testing instructions in the file comments
```

2. **Memory Scaling Test**
```bash
# Apply the memory scaling example
kubectl apply -f examples/memory-scaling-with-test.yaml

# Follow the testing instructions in the file comments
```

3. **Pod Restart Test**
```bash
# Apply the pod restart example
kubectl apply -f examples/pod-restart-with-test.yaml

# Follow the testing instructions in the file comments
```

### Monitoring Tests

Monitor your tests using standard Kubernetes tools:
```bash
# Watch pods and policies
kubectl get pods,selfremediationpolicy -w

# Monitor resource usage
kubectl top pods

# Check policy status
kubectl describe selfremediationpolicy
```

## Safety Features

### Protected Resources
```yaml
metadata:
  labels:
    kubemedic.io/protected: "true"  # Prevents any remediation
```

### Namespace Exclusion
```yaml
metadata:
  labels:
    kubemedic.io/exclude: "true"  # Excludes namespace from remediation
```

### Resource Limits
- Maximum 2x scaling factor
- Minimum 1 pod maintained
- Maximum 2-hour remediation duration
- Namespace quota validation

## Configuration

### values.yaml Highlights
```yaml
rbac:
  # Namespace restrictions
  namespaceRestrictions:
    enabled: true
    denied: ["kube-system", "kube-public"]

  # Resource protection
  resourceRestrictions:
    enabled: true
    allowed: ["deployments", "statefulsets"]

remediation:
  # Safety limits
  safetyLimits:
    maxScaleFactor: 2
    minPods: 1
    maxScalingDuration: "2h"

monitoring:
  # Metrics source
  metricsSource: "kubernetes"  # or "prometheus"
  # Optional Prometheus integration
  prometheus:
    enabled: false
  # Optional Grafana integration
  grafana:
    enabled: false
```

## Support

- 📖 [Documentation](./docs)
- 💬 [Discussions](https://github.com/ikepcampbell/kubemedic/discussions)
- 💼 [Follow Ike](https://linkedin.com/in/isaac-campbell)
- 📧 [Support](mailto:ike@isaacs.cloud)

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.
