# KubeMedic - Intelligent Kubernetes Auto-Remediation

KubeMedic is an advanced Kubernetes operator that goes beyond traditional autoscaling to provide comprehensive, intelligent, and safe automated remediation for your clusters. Unlike standard autoscalers, KubeMedic can handle complex scenarios through multi-dimensional analysis and varied remediation actions.

## What Makes KubeMedic Different?

### 🎯 Multi-Dimensional Remediation
Unlike traditional autoscalers that only scale based on single metrics, KubeMedic can:
- Combine multiple metrics (CPU, Memory, Error Rates, Pod Restarts)
- Perform varied remediation actions (scaling, restarts, rollbacks)
- Adjust resource limits and HPA settings dynamically
- Execute staged responses to escalating issues

### 🧠 Intelligent Policy Engine
- Combines multiple conditions for smarter decisions
- Supports duration-based triggers (e.g., high CPU for >5 minutes)
- Implements cooldown periods to prevent oscillation
- Provides conflict resolution with other controllers

### 🛡️ Advanced Safety Mechanisms
- Automatic state preservation before actions
- Gradual or immediate reversion strategies
- Resource quota awareness
- Protected namespaces and resources
- Audit trail of all actions

### 🔌 Rich Integration Capabilities
- Native Grafana integration for visualization
- Pre/post action webhooks for notifications
- Prometheus metrics integration
- State backup system

## Prerequisites

### Required
- Kubernetes cluster (v1.16+)
- [Metrics Server](https://github.com/kubernetes-sigs/metrics-server)
- [cert-manager](https://cert-manager.io/docs/installation/) (optional, recommended)

### Optional
- Prometheus & Grafana for advanced visualization
- Webhook endpoints for notifications

## Quick Start

1. **Install Metrics Server** (if not already installed)
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

2. **Install KubeMedic**
```bash
kubectl apply -f deploy/components.yaml
```

3. **Create an Intelligent Remediation Policy**
```yaml
apiVersion: remediation.kubemedic.io/v1alpha1
kind: SelfRemediationPolicy
metadata:
  name: advanced-remediation
  namespace: my-app
spec:
  rules:
    # CPU-based scaling with gradual approach
    - name: cpu-scaling
      conditions:
        - type: CPUUsage
          threshold: "80"
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
    
    # Memory leak detection and handling
    - name: memory-leak-handler
      conditions:
        - type: MemoryUsage
          threshold: "90"
          duration: "10m"
      actions:
        - type: RestartPod
          target:
            kind: Pod
            name: my-service
    
    # Error rate-based remediation
    - name: error-handler
      conditions:
        - type: ErrorRate
          threshold: "100"
          duration: "2m"
      actions:
        - type: RollbackDeployment
          target:
            kind: Deployment
            name: my-service

  cooldownPeriod: "5m"
  grafanaIntegration:
    enabled: true
    webhookUrl: "https://grafana.example.com/webhook"
```

## Key Features

### 🔄 Intelligent Scaling
- Dynamic HPA limit adjustments
- Temporary resource overrides
- Gradual scaling with automatic revert
- Multi-metric based decisions

### 🛠️ Advanced Remediation
- Automatic pod restarts for memory leaks
- Deployment rollbacks for error spikes
- Resource limit adjustments
- Custom webhook integrations

### 📊 Comprehensive Monitoring
- Real-time metrics analysis
- Historical data tracking
- Grafana dashboard integration
- Prometheus metrics export

### 🔒 Enterprise-Grade Safety
- Protected system namespaces
- Resource quotas and limits
- State preservation and rollback
- Action audit trail

## Documentation

For detailed documentation, visit our [Documentation](docs/README.md) section:
- [Getting Started Guide](docs/getting-started/README.md)
- [Core Concepts](docs/concepts/README.md)
- [Advanced Usage](docs/advanced-usage/README.md)
- [API Reference](docs/reference/api.md)

## Examples

Explore our comprehensive examples:
- [CPU Scaling](examples/cpu-scaling-with-test.yaml)
- [Memory Management](examples/memory-scaling-with-test.yaml)
- [Error Rate Handling](examples/pod-restart-with-test.yaml)

Each example includes both the policy configuration and a test application to demonstrate the functionality.

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.