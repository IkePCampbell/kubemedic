# KubeMedic - Safe Kubernetes Auto-Remediation

KubeMedic is a Kubernetes operator that safely automates common remediation tasks while protecting your cluster from unintended consequences.

## Prerequisites

### Required
- Kubernetes cluster (v1.16+)
- [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) installed in your cluster
  ```bash
  kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
  ```
- [cert-manager](https://cert-manager.io/docs/installation/) for managing TLS certificates (optional, but recommended for webhook)
  ```bash
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
  ```

### Optional
- Prometheus & Grafana for advanced metrics visualization
  - KubeMedic works with the native Kubernetes metrics API by default
  - Can be integrated with Prometheus for historical data and advanced querying
  - Grafana dashboards available for visualization

## Webhook Configuration

### Using cert-manager
1. Install cert-manager in your cluster.
2. Configure a `Certificate` resource for the webhook service.
3. Update the `ValidatingWebhookConfiguration` to use the cert-manager managed CA.

### Bring Your Own CA Bundle
1. Generate a CA certificate and key using OpenSSL:
   ```bash
   openssl req -x509 -newkey rsa:4096 -keyout tls.key -out tls.crt -days 365 -nodes -subj "/CN=kubemedic-webhook.kubemedic.svc"
   ```

2. Create a Kubernetes secret to store the certificate and key:
   ```bash
   kubectl create secret tls kubemedic-webhook-cert --cert=tls.crt --key=tls.key -n kubemedic
   ```

3. Encode the CA certificate in base64:
   ```bash
   cat tls.crt | base64 | tr -d '\n'
   ```

4. Update the `ValidatingWebhookConfiguration` with the base64-encoded CA bundle:
   ```yaml
   apiVersion: admissionregistration.k8s.io/v1
   kind: ValidatingWebhookConfiguration
   metadata:
     name: kubemedic-validating-webhook
   webhooks:
   - name: validate.remediation.kubemedic.io
     clientConfig:
       service:
         name: kubemedic-webhook
         namespace: kubemedic
         path: "/validate"
       caBundle: "<BASE64_ENCODED_CERTIFICATE>"
     rules:
     - operations: ["CREATE", "UPDATE"]
       apiGroups: ["remediation.kubemedic.io"]
       apiVersions: ["v1alpha1"]
       resources: ["selfremediationpolicies"]
     failurePolicy: Fail
     sideEffects: None
     admissionReviewVersions: ["v1"]
   ```

### Automating caBundle Update from a Secret

You can automate the process of updating the `caBundle` in the `ValidatingWebhookConfiguration` using the following script:

```bash
#!/bin/bash

# Retrieve the CA certificate from the Secret
CA_CERT=$(kubectl get secret kubemedic-webhook-cert -n kubemedic -o jsonpath='{.data.tls\.crt}' | base64 --decode)

# Base64 encode the CA certificate
CA_BUNDLE=$(echo "$CA_CERT" | base64 | tr -d '\n')

# Update the ValidatingWebhookConfiguration with the CA bundle
kubectl patch validatingwebhookconfiguration kubemedic-validating-webhook --type='json' -p="[{\"op\": \"replace\", \"path\": \"/webhooks/0/clientConfig/caBundle\", \"value\":\"$CA_BUNDLE\"}]"
```

Run this script after deploying your webhook and creating the Secret to update the `caBundle` automatically.

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

- Built with Go, Kubernetes APIs, and controller-runtime using operator-sdk

## What Others Do

In the Kubernetes ecosystem, several tools are commonly used to manage and optimize cluster operations:

- **Cluster Autoscaler:** Automatically adjusts the number of nodes in a cluster based on resource usage. It focuses on infrastructure-level scaling, ensuring that there are enough nodes to handle the workload.

- **Helm:** A package manager for Kubernetes that simplifies the deployment and management of applications. It helps in managing application lifecycles and dependencies.

- **Argo CD:** A declarative, GitOps continuous delivery tool for Kubernetes. It automates the deployment of applications and ensures that the live state of the cluster matches the desired state defined in Git.

While these tools provide essential functionalities for managing Kubernetes clusters, KubeMedic goes a step further by focusing on application-level remediation. It automates responses to specific conditions within the cluster, such as high CPU usage or error rates, by executing predefined actions like scaling or restarting pods.

### Combining KubeMedic with Other Tools

KubeMedic can be effectively combined with other Kubernetes tools to enhance cluster management:

- **With Cluster Autoscaler:** Use KubeMedic to handle application-level scaling and remediation, while the Cluster Autoscaler manages node-level scaling. This ensures that both application performance and infrastructure efficiency are optimized.

- **With Helm:** Deploy KubeMedic and its policies using Helm charts for easy management and versioning. Helm can also be used to manage the lifecycle of applications that KubeMedic monitors.

- **With Argo CD:** Use Argo CD to manage the GitOps workflow for KubeMedic policies. This ensures that any changes to remediation strategies are tracked and versioned in Git, providing a clear audit trail and facilitating rollbacks if necessary.

By integrating KubeMedic with these tools, you can achieve a comprehensive and automated approach to managing both application and infrastructure health in your Kubernetes clusters.

## Components and Images

KubeMedic is composed of several components, each packaged as a separate Docker image. Here's a breakdown of these components and their roles:

1. **Controller Manager:**
   - **Image:** `kubemedic-controller-manager`
   - **Role:** This is the core component of the operator. It runs the reconciliation loops that manage the state of the cluster based on the custom resources (CRDs) defined by KubeMedic. It interacts with the Kubernetes API to monitor resources and apply remediation actions as specified in the policies.

2. **Webhook Server:**
   - **Image:** `kubemedic-webhook`
   - **Role:** This component handles admission webhooks, which are used to validate or mutate Kubernetes resources during their creation or update. The webhook server ensures that the custom resources conform to the expected schema and business logic before they are persisted in the cluster.

3. **Metrics or Auxiliary Services:**
   - **Image:** This could be another image if there are additional services like a metrics server or a separate component for handling specific tasks.
   - **Role:** These services might be responsible for collecting metrics, providing dashboards, or handling specific integrations (e.g., with Prometheus or Grafana).

### Why Multiple Images?

- **Separation of Concerns:** Each component has a distinct responsibility, allowing for better organization and maintainability.
- **Scalability:** Components can be scaled independently based on their resource requirements and load.
- **Security and Stability:** Isolating components reduces the risk of a single point of failure and allows for more granular security policies.
- **Flexibility:** Different components can be updated or replaced independently, facilitating continuous integration and deployment.

By understanding the roles of each component and managing them effectively, you can ensure that your Kubernetes operator functions smoothly and efficiently.