# KubeMedic - Kubernetes just got its CNA 🏥

KubeMedic is a powerful, proactive Kubernetes operator that acts as your cluster's autonomous medical system. Instead of waiting for disasters to strike, KubeMedic continuously monitors your services and takes preemptive actions to maintain optimal cluster health.

## 🌟 Why KubeMedic?

### Proactive, Not Reactive
Traditional monitoring tools alert you after problems occur. KubeMedic takes action before small issues become major incidents:
- 🔄 Automatically scales resources when usage trends indicate impending bottlenecks
- 🚫 Prevents cascading failures by identifying and addressing early warning signs
- 🎯 Takes precise, targeted actions based on customizable conditions

### Self-Sufficient Yet Integrated
- 🤖 Operates autonomously without requiring constant human intervention
- 📊 Optional Grafana integration for enhanced visibility (but not required!)
- 🔌 Webhook support for seamless integration with your existing monitoring stack

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
- Kubernetes cluster v1.16+ (AKS, EKS, GKE, or any other flavor - we don't discriminate! 😉)
- `kubectl` installed and configured
- Prometheus installed in your cluster (we need those sweet, sweet metrics!)

#### Optional (but recommended)
- Grafana v8.0+ (for beautiful visualizations)
- Helm v3+ (for the smoothest installation experience)
- A cup of coffee ☕ (because everything's better with coffee)

### Installation

#### Quick Start (I'm Feeling Lucky 🎲)
```bash
# The one-liner for the brave
kubectl apply -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/config/deploy/kubemedic.yaml
```

#### The "I Like Helm" Way 🎯
```bash
# Add our shiny Helm repository
helm repo add kubemedic https://kubemedic.github.io/charts

# Update your repos (always good practice!)
helm repo update

# Install KubeMedic with default settings
helm install kubemedic kubemedic/kubemedic

# Or, if you're feeling fancy, with custom values
helm install kubemedic kubemedic/kubemedic -f my-values.yaml
```

#### The "I Want Control" Way 🎮
```bash
# Clone the repo
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

# Should see something like:
# NAME                        READY   STATUS    RESTARTS   AGE
# kubemedic-controller-xxx    1/1     Running   0          1m

# Check if the CRDs are installed
kubectl get crds | grep kubemedic
```

### Your First Policy 🎉

Let's create a simple policy to auto-scale a deployment when CPU gets hangry:

```yaml
apiVersion: remediation.kubemedic.io/v1alpha1
kind: SelfRemediationPolicy
metadata:
  name: my-first-policy
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

Save this as `first-policy.yaml` and apply:
```bash
kubectl apply -f first-policy.yaml
```

## 📚 Documentation

### Architecture
KubeMedic follows a simple but powerful architecture:
1. **Watch** - Continuously monitors your resources
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

Check out our [examples](./examples) directory for more configurations and use cases!

## 🔧 Troubleshooting

### Common Issues

#### "The operator isn't doing anything!"
1. Check if the operator is running:
   ```bash
   kubectl logs -n kubemedic-system deployment/kubemedic-controller
   ```
2. Verify your policy syntax:
   ```bash
   kubectl get srp -o yaml
   ```

#### "I'm getting Prometheus errors"
1. Check Prometheus connectivity:
   ```bash
   kubectl exec -it -n kubemedic-system deploy/kubemedic-controller -- curl -f prometheus:9090/-/healthy
   ```

## 🤝 Contributing

We love contributions! Whether it's:
- 🐛 Bug fixes
- ✨ New features
- 📚 Documentation improvements
- 🎨 UI enhancements

Check out our [Contributing Guide](CONTRIBUTING.md) to get started!

## 📜 License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## 🌟 Star Us!

If KubeMedic has helped your cluster stay healthy, consider giving us a star! It helps others discover the project and keeps us motivated! ⭐

## 🙋‍♂️ Need Help?

- 📖 [Documentation](./docs)
- 💬 [Discord Community](https://discord.gg/kubemedic)
- 🐤 [Twitter](https://twitter.com/kubemedic)
- 📧 [Email Support](mailto:support@kubemedic.io)

Remember: A healthy cluster is a happy cluster! 🎉

