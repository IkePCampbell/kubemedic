# KubeMedic Examples

This directory contains complete, tested examples for KubeMedic use cases. Each example includes both the policy configuration and a test application to demonstrate the functionality.

## Examples Overview

### CPU Scaling (`cpu-scaling-with-test.yaml`)
- Test application that generates controlled CPU load
- Policy that scales based on CPU usage
- Complete testing instructions
- Automatic revert after load decreases

### Memory Scaling (`memory-scaling-with-test.yaml`)
- Test application with configurable memory consumption
- Two-tier policy (warning and critical thresholds)
- Gradual scaling approach
- Memory protection mechanisms

### Pod Restart (`pod-restart-with-test.yaml`)
- Error-generating test application
- Error rate monitoring and automatic restarts
- Deployment rollback on repeated failures
- Error rate control and monitoring

## Using the Examples

Each example file is self-contained and includes:
1. A test application deployment
2. The corresponding remediation policy
3. Supporting resources (Services, etc.)
4. Step-by-step testing instructions

### General Usage Pattern

1. Apply an example:
   ```bash
   kubectl apply -f <example-name>.yaml
   ```

2. Follow the testing instructions in the file comments

3. Monitor the results using:
   ```bash
   kubectl get pods,selfremediationpolicy
   kubectl top pods
   ```

4. Clean up when done:
   ```bash
   kubectl delete -f <example-name>.yaml
   ```

## Prerequisites

- Kubernetes cluster (v1.16+)
- Metrics Server installed
- KubeMedic operator installed
- `kubectl` configured with cluster access

## Best Practices

1. **Testing in Dev First**
   - Always test in a development environment first
   - Use resource quotas to prevent runaway scaling
   - Monitor closely during initial deployment

2. **Resource Management**
   - Start with conservative thresholds
   - Use appropriate cooldown periods
   - Set reasonable scaling limits

3. **Monitoring**
   - Watch pod metrics during tests
   - Monitor policy status
   - Check logs for remediation actions

## Example Structure

Each example follows this structure:
```yaml
---
# Test Application
apiVersion: apps/v1
kind: Deployment
# ... test application configuration ...

---
# Supporting Services
apiVersion: v1
kind: Service
# ... service configuration ...

---
# Remediation Policy
apiVersion: remediation.kubemedic.io/v1alpha1
kind: SelfRemediationPolicy
# ... policy configuration ...

---
# Testing Instructions
# Step-by-step instructions in comments
```

## Contributing New Examples

When contributing new examples:
1. Include both test application and policy
2. Add clear testing instructions
3. Follow the existing format
4. Test thoroughly before submitting
5. Document any prerequisites or special requirements 