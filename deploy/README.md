# KubeMedic Deployment

This directory contains the Kubernetes manifest needed to deploy KubeMedic in your cluster.

## Quick Install

```bash
# Install KubeMedic
kubectl apply -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/deploy/all-in-one.yaml

# Verify installation
kubectl get pods -n kubemedic
```

## What Gets Installed

- Namespace: `kubemedic`
- CRD: `SelfRemediationPolicy`
- RBAC: ServiceAccount, ClusterRole, and ClusterRoleBinding
- Deployment: KubeMedic controller

## Configuration

The deployment uses sensible defaults:
- Memory limit: 256Mi
- CPU limit: 200m
- Single replica
- Latest stable image

## Uninstall

```bash
kubectl delete -f https://raw.githubusercontent.com/ikepcampbell/kubemedic/main/deploy/all-in-one.yaml
``` 