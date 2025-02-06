# HPA Integration

KubeMedic can work alongside Kubernetes Horizontal Pod Autoscaler (HPA) by temporarily adjusting its settings during high-load situations.

## Overview

KubeMedic enhances HPA functionality by:
1. Temporarily overriding HPA limits during spikes
2. Preserving original configurations
3. Automatically reverting changes
4. Coordinating with other scaling mechanisms

## Configuration Examples

### Basic HPA Override

```yaml
apiVersion: remediation.kubemedic.io/v1alpha1
kind: SelfRemediationPolicy
metadata:
  name: hpa-override-policy
spec:
  rules:
    - name: high-load-scaling
      conditions:
        - type: CPUUsage
          threshold: "85%"
          duration: "3m"
      actions:
        - type: AdjustHPALimits
          target:
            kind: HorizontalPodAutoscaler
            name: my-app-hpa
          scalingParams:
            temporaryMaxReplicas: 10
            scalingDuration: "30m"
            revertStrategy: "Gradual"
```

### Multiple Metric Scaling

```yaml
spec:
  rules:
    - name: multi-metric-scaling
      conditions:
        - type: CPUUsage
          threshold: "80%"
          duration: "3m"
        - type: MemoryUsage
          threshold: "75%"
          duration: "3m"
      actions:
        - type: AdjustHPALimits
          target:
            kind: HorizontalPodAutoscaler
            name: my-app-hpa
          scalingParams:
            temporaryMaxReplicas: 8
            scalingDuration: "15m"
```

## How It Works

1. **Monitoring Phase**
   - KubeMedic monitors specified metrics
   - Evaluates against thresholds
   - Checks duration requirements

2. **Action Phase**
   ```yaml
   actions:
     - type: AdjustHPALimits
       scalingParams:
         temporaryMaxReplicas: 5    # New maximum
         scalingDuration: "30m"     # Override duration
         revertStrategy: "Gradual"  # How to scale back
   ```

3. **State Preservation**
   - Original HPA settings stored in annotations
   ```yaml
   metadata:
     annotations:
       kubemedic.io/original-max-replicas: "3"
       kubemedic.io/scaling-expiry: "2024-02-05T15:04:05Z"
   ```

4. **Reversion Phase**
   - Automatic reversion after duration
   - Gradual or immediate scaling back
   - Annotation cleanup

## Integration with Existing HPAs

### Example HPA Configuration

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 2
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 80
```

### KubeMedic Override

```yaml
actions:
  - type: AdjustHPALimits
    target:
      kind: HorizontalPodAutoscaler
      name: my-app-hpa
    scalingParams:
      temporaryMaxReplicas: 8    # Temporary override
      scalingDuration: "1h"      # Duration of override
```

## Best Practices

1. **Conservative Overrides**
   ```yaml
   scalingParams:
     temporaryMaxReplicas: 7    # Not too high
     scalingDuration: "30m"     # Not too long
   ```

2. **Gradual Changes**
   ```yaml
   scalingParams:
     revertStrategy: "Gradual"
     notificationWebhook: "https://notify.example.com"
   ```

3. **Multiple Conditions**
   ```yaml
   conditions:
     - type: CPUUsage
       threshold: "85%"
       duration: "3m"
     - type: ErrorRate
       threshold: "10"
       duration: "2m"
   ```

## Troubleshooting

### Common Issues

1. **HPA Not Updating**
   ```bash
   kubectl describe hpa my-app-hpa
   kubectl get events --field-selector involvedObject.kind=HorizontalPodAutoscaler
   ```

2. **Scaling Conflicts**
   ```bash
   kubectl get hpa my-app-hpa -o yaml | grep kubemedic.io
   ```

3. **Permission Issues**
   ```bash
   kubectl auth can-i update horizontalpodautoscalers
   ```

### Debugging

1. Check KubeMedic logs:
```bash
kubectl logs -n kubemedic deployment/kubemedic-controller-manager
```

2. Verify annotations:
```bash
kubectl get hpa my-app-hpa -o jsonpath='{.metadata.annotations}'
```

3. Check scaling status:
```bash
kubectl get events --field-selector involvedObject.kind=HorizontalPodAutoscaler,involvedObject.name=my-app-hpa
```

## Advanced Configurations

### Staged Scaling

```yaml
spec:
  rules:
    - name: initial-response
      conditions:
        - type: CPUUsage
          threshold: "80%"
          duration: "3m"
      actions:
        - type: AdjustHPALimits
          scalingParams:
            temporaryMaxReplicas: 6
            scalingDuration: "15m"
    - name: emergency-response
      conditions:
        - type: CPUUsage
          threshold: "90%"
          duration: "2m"
      actions:
        - type: AdjustHPALimits
          scalingParams:
            temporaryMaxReplicas: 10
            scalingDuration: "10m"
```

### Complex Metrics

```yaml
spec:
  rules:
    - name: complex-scaling
      conditions:
        - type: CPUUsage
          threshold: "85%"
          duration: "3m"
        - type: ErrorRate
          threshold: "5"
          duration: "1m"
      actions:
        - type: AdjustHPALimits
          scalingParams:
            temporaryMaxReplicas: 8
            scalingDuration: "20m"
            revertStrategy: "Gradual"
```

## Next Steps

- [Scaling Strategies](scaling-strategies.md)
- [Conflict Resolution](conflict-resolution.md)
- [Webhook Integration](webhooks.md) 