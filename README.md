# KubeMedic - Automated First Aid for Kubernetes Remediation

KubeMedic is a powerful, proactive Kubernetes operator that acts as your cluster's autonomous remediation system. Instead of waiting for incidents to occur, KubeMedic continuously monitors your services and takes preemptive actions to maintain optimal cluster health.

## 🌟 Why KubeMedic?

### Proactive, Not Reactive
Traditional monitoring tools alert you after problems occur. KubeMedic takes action before small issues become major incidents:
- 🔄 Automatically scales resources based on usage trends
- 🚫 Prevents cascading failures through early detection
- 🎯 Takes precise, targeted actions based on customizable conditions

### Self-Sufficient Yet Integrated
- 🤖 Operates autonomously with minimal human intervention
- 📊 Optional Grafana integration for enhanced visibility
- 🔌 Webhook support for existing monitoring stack integration

### Smart Remediation
- 🧠 Intelligent cooldown periods prevent remediation storms
- 🎛️ Fine-grained control over conditions and actions
- 🛡️ Built-in safeguards and manual override options

## 🚀 Features

### Comprehensive Monitoring
- CPU and Memory usage tracking
- Error rate monitoring
- Pod restart counting
- Extensible monitoring framework

### Automated Actions
- Dynamic scaling (up/down)
- Intelligent pod restarts
- Deployment rollbacks
- Custom action framework

### Enterprise Ready
- 🔒 Secure by design
- 📈 Prometheus metrics
- 📋 Detailed audit logging
- 🎯 Namespace-scoped policies

## 🛠️ Getting Started

### Prerequisites

#### Required
- Kubernetes cluster v1.16+ (AKS, EKS, GKE, or any other distribution)
- `kubectl` installed and configured
- Prometheus installed in your cluster for metrics collection

#### Optional (but recommended)
- Grafana v8.0+ for metrics visualization
- Helm v3+ for streamlined installation

### Installation

#### Quick Start
```bash
kubectl apply -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/config/deploy/kubemedic.yaml
```

#### Using Helm
```bash
# Add the repository
helm repo add kubemedic https://kubemedic.github.io/charts

# Update repositories
helm repo update

# Install KubeMedic with default settings
helm install kubemedic kubemedic/kubemedic

# Or with custom values
helm install kubemedic kubemedic/kubemedic -f my-values.yaml
```

#### Manual Installation
```bash
# Clone the repository
git clone https://github.com/ikepcampbell/kubemedic.git
cd kubemedic

# Install CRDs
make install

# Deploy the operator
make deploy IMG=ghcr.io/ikepcampbell/kubemedic:latest
```

### Verifying the Installation

```bash
# Check if the operator is running
kubectl get pods -n kubemedic-system

# Expected output:
# NAME                        READY   STATUS    RESTARTS   AGE
# kubemedic-controller-xxx    1/1     Running   0          1m

# Verify CRD installation
kubectl get crds | grep kubemedic
```

### Basic Configuration

Create a policy to auto-scale a deployment based on CPU usage:

```yaml
apiVersion: remediation.kubemedic.io/v1alpha1
kind: SelfRemediationPolicy
metadata:
  name: cpu-autoscale-policy
  namespace: default
spec:
  rules:
    - name: cpu-scaling
      conditions:
        - type: CPUUsage
          threshold: "80%"
          duration: "5m"
      actions:
        - type: ScaleUp
          target:
            kind: Deployment
            name: my-app
  cooldownPeriod: "10m"
```

Apply the configuration:
```bash
kubectl apply -f cpu-autoscale-policy.yaml
```

## 📚 Documentation

### Architecture
KubeMedic follows a streamlined architecture:
1. **Watch** - Monitors your resources
2. **Analyze** - Evaluates conditions against thresholds
3. **Act** - Takes remediation actions when needed
4. **Learn** - Adjusts based on action outcomes

### Configuration Options

#### Prometheus Integration
```yaml
# values.yaml
prometheus:
  url: http://prometheus.monitoring:9090
  scrapeInterval: 30s
```

#### Grafana Integration
```yaml
# in your SelfRemediationPolicy
spec:
  grafanaIntegration:
    enabled: true
    webhookUrl: "https://grafana.example.com/webhook"
```

See our [examples](./examples) directory for more configurations.

## 🔧 Troubleshooting

### Common Issues

#### Operator Status Check
1. Verify operator status:
   ```bash
   kubectl logs -n kubemedic-system deployment/kubemedic-controller
   ```
2. Check policy configuration:
   ```bash
   kubectl get srp -o yaml
   ```

#### Prometheus Connectivity
1. Verify Prometheus connection:
   ```bash
   kubectl exec -it -n kubemedic-system deploy/kubemedic-controller -- curl -f prometheus:9090/-/healthy
   ```

## 🤝 Contributing

We welcome contributions in various forms:
- 🐛 Bug fixes
- ✨ New features
- 📚 Documentation improvements
- 🎨 UI enhancements

See our [Contributing Guide](CONTRIBUTING.md) for details.

## 📜 License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## 🌟 Community

If you find KubeMedic useful, please star the repository. It helps others discover the project!

## 📞 Support

- 📖 [Documentation](./docs)
- 💬 [Community Forum](https://github.com/ikepcampbell/kubemedic/discussions)
- 🐤 [Twitter](https://twitter.com/kubemedic)
- 📧 [Support](mailto:support@kubemedic.io)

Remember: A healthy cluster is a happy cluster! 🎉

