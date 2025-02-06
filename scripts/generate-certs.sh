#!/bin/bash

# Create deploy directory if it doesn't exist
mkdir -p deploy

# Generate CA key and certificate
openssl genrsa -out deploy/ca.key 2048
openssl req -x509 -new -nodes -key deploy/ca.key -subj "/CN=kubemedic-webhook-ca" -days 365 -out deploy/ca.crt

# Generate webhook server key
openssl genrsa -out deploy/webhook-server.key 2048

# Generate CSR
cat > deploy/csr.conf << EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
[dn]
CN = kubemedic-webhook-service.kubemedic.svc
[v3_ext]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment,digitalSignature
extendedKeyUsage=serverAuth
subjectAltName=@alt_names
[alt_names]
DNS.1 = kubemedic-webhook-service.kubemedic.svc
DNS.2 = kubemedic-webhook-service.kubemedic.svc.cluster.local
EOF

# Generate certificate
openssl req -new -key deploy/webhook-server.key -out deploy/webhook-server.csr -config deploy/csr.conf
openssl x509 -req -in deploy/webhook-server.csr -CA deploy/ca.crt -CAkey deploy/ca.key -CAcreateserial -out deploy/webhook-server.crt -days 365 -extensions v3_ext -extfile deploy/csr.conf

# Create the Secret yaml
echo "apiVersion: v1
kind: Secret
metadata:
  name: kubemedic-webhook-cert
  namespace: kubemedic
type: kubernetes.io/tls
data:
  tls.crt: $(cat deploy/webhook-server.crt | base64 | tr -d '\n')
  tls.key: $(cat deploy/webhook-server.key | base64 | tr -d '\n')
  ca.crt: $(cat deploy/ca.crt | base64 | tr -d '\n')" > deploy/webhook-secret.yaml

# Create the ValidatingWebhookConfiguration yaml with the CA bundle
echo "apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: kubemedic-validating-webhook
webhooks:
- name: validate.remediation.kubemedic.io
  rules:
  - apiGroups: [\"remediation.kubemedic.io\"]
    apiVersions: [\"v1alpha1\"]
    operations: [\"CREATE\", \"UPDATE\"]
    resources: [\"selfremediationpolicies\"]
    scope: \"Namespaced\"
  clientConfig:
    service:
      namespace: kubemedic
      name: kubemedic-webhook-service
      path: \"/validate\"
    caBundle: $(cat deploy/ca.crt | base64 | tr -d '\n')
  admissionReviewVersions: [\"v1\"]
  sideEffects: None
  timeoutSeconds: 10
  failurePolicy: Fail" > deploy/webhook-config.yaml

echo "Generated files in deploy/:"
echo "- ca.key: CA private key"
echo "- ca.crt: CA certificate"
echo "- webhook-server.key: Webhook server private key"
echo "- webhook-server.crt: Webhook server certificate"
echo "- webhook-secret.yaml: Kubernetes Secret with certificates"
echo "- webhook-config.yaml: ValidatingWebhookConfiguration with CA bundle"
echo ""
echo "To use custom certificates:"
echo "1. kubectl apply -f deploy/webhook-secret.yaml"
echo "2. kubectl apply -f deploy/webhook-config.yaml" 