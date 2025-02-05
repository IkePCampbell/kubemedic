# Remediation Policies

Remediation policies are the core concept in KubeMedic that define how your cluster should automatically respond to various conditions.

## Policy Structure

A remediation policy consists of:
- Rules: Sets of conditions and actions
- Cooldown periods: Prevent action storms
- Integration settings: Optional external system connections

### Basic Policy Example

```yaml
apiVersion: remediation.kubemedic.io/v1alpha1
kind: SelfRemediationPolicy
metadata:
  name: basic-scaling-policy
  namespace: default
spec:
  rules:
    - name: high-cpu-scaling
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

## Policy Components

### Conditions

Conditions define when actions should be triggered:

```yaml
conditions:
  - type: CPUUsage    # What to monitor
    threshold: "80%"   # When to trigger
    duration: "5m"     # How long condition must be true
```

Available condition types:
- `CPUUsage`: CPU utilization percentage
- `MemoryUsage`: Memory utilization percentage
- `ErrorRate`: Error count per second
- `PodRestarts`: Number of pod restarts

### Actions

Actions define what remediation to perform:

```yaml
actions:
  - type: ScaleUp
    target:
      kind: Deployment
      name: my-app
    scalingParams:
      temporaryMaxReplicas: 5
      scalingDuration: "30m"
      revertStrategy: "Gradual"
```

Available action types:
- `ScaleUp`: Increase replicas
- `ScaleDown`: Decrease replicas
- `RestartPod`: Restart problematic pods
- `RollbackDeployment`: Revert to previous version
- `AdjustHPALimits`: Modify HPA settings
- `UpdateResources`: Change resource requests/limits

### Scaling Parameters

Advanced scaling configuration:

```yaml
scalingParams:
  temporaryMaxReplicas: 5        # Maximum replicas during scaling
  scalingDuration: "30m"         # How long to maintain scaling
  revertStrategy: "Gradual"      # How to scale back down
  notificationWebhook: "..."     # Where to send notifications
```

### Integration Settings

Optional external system integration:

```yaml
grafanaIntegration:
  enabled: true
  webhookUrl: "https://grafana.example.com/webhook"
```

## Policy Scope

Policies can be:
- Namespace-scoped: Apply to resources in specific namespace
- Cluster-scoped: Apply to resources across the cluster

## Best Practices

1. **Start Conservative**
   ```yaml
   conditions:
     - type: CPUUsage
       threshold: "85%"    # Start high
       duration: "5m"      # Give time to stabilize
   ```

2. **Use Gradual Scaling**
   ```yaml
   scalingParams:
    revertStrategy: "Gradual"
    scalingDuration: "30m"
   ```

3. **Implement Safeguards**
   ```yaml
   cooldownPeriod: "15m"    # Prevent rapid oscillation
   ```

4. **Add Monitoring**
   ```yaml
   preActionHook: "https://monitoring.example.com/webhook"
   postActionHook: "https://monitoring.example.com/webhook"
   ```

## Advanced Usage

### Multiple Conditions

```yaml
conditions:
  - type: CPUUsage
    threshold: "80%"
    duration: "5m"
  - type: ErrorRate
    threshold: "10"
    duration: "2m"
```

### Chained Actions

```yaml
actions:
  - type: ScaleUp
    target:
      kind: Deployment
      name: my-app
  - type: AdjustHPALimits
    target:
      kind: HorizontalPodAutoscaler
      name: my-app-hpa
```

### Temporary Overrides

```yaml
actions:
  - type: AdjustHPALimits
    scalingParams:
      temporaryMaxReplicas: 10
      scalingDuration: "1h"
```

## Policy Validation

KubeMedic validates policies for:
- Syntax correctness
- Resource existence
- Permission availability
- Configuration conflicts

## Troubleshooting

Common policy issues:
1. Invalid condition thresholds
2. Missing target resources
3. RBAC permission issues
4. Conflicting actions

Check policy status:
```bash
kubectl get srp my-policy -o yaml
```

## Next Steps

- [Conditions and Triggers](conditions.md)
- [Actions and Effects](actions.md)
- [Advanced Usage Examples](../advanced-usage/README.md) 