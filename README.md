# Kubernetes DNS Config Webhook

A Kubernetes mutating admission webhook that automatically configures DNS settings for Pods in your cluster. This webhook allows you to enforce consistent DNS configurations across all Pods by modifying their DNS settings during the admission process.

## Features

- Automatically injects or updates DNS configuration for all Pods
- Configurable through Kubernetes ConfigMap
- Supports all Kubernetes DNS configuration options (nameservers, searches, options)
- High availability with multiple replicas
- Health checks and monitoring
- TLS encryption for webhook communication

## Prerequisites

- Kubernetes cluster (1.16+)
- `kubectl` configured to communicate with your cluster
- Cluster admin privileges to create MutatingWebhookConfiguration

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ParinLL/k8s-dnsconfig-webhook.git
cd k8s-dnsconfig-webhook
```

2. Create the necessary Kubernetes resources:

```bash
# Create RBAC resources
kubectl apply -f deploy/rbac.yaml

# Generate TLS certificates
kubectl apply -f deploy/certificate.yml

# Create ConfigMap with DNS settings
kubectl apply -f deploy/configmap.yaml

# Deploy the webhook
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/service.yaml

# Configure the webhook
kubectl apply -f deploy/MutatingWebhookConfiguration.yaml
```

## Configuration

The DNS configuration is managed through a ConfigMap. Here's an example configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dns-config-settings
data:
  config.yaml: |
    dnsConfig:
      nameservers:
        - 8.8.8.8
        - 8.8.4.4
      searches:
        - ns1.svc.cluster.local
        - svc.cluster.local
        - cluster.local
      options:
        - name: ndots
          value: "5"
        - name: timeout
          value: "3"
        - name: attempts
          value: "2"
```

### Configuration Options

- `nameservers`: List of DNS server IP addresses
- `searches`: List of DNS search domains
- `options`: List of DNS resolver options
  - `ndots`: Threshold for number of dots in name resolution
  - `timeout`: DNS resolution timeout in seconds
  - `attempts`: Number of attempts before giving up
  - Other valid resolv.conf options

## Monitoring

The webhook exposes the following endpoints for monitoring:

- `/health`: Health check endpoint (HTTPS)
  - Used by Kubernetes liveness and readiness probes
  - Returns 200 OK when the webhook is healthy

## Resource Requirements

Minimal resource requirements:
- Memory: 64Mi (request), 128Mi (limit)
- CPU: 100m (request), 200m (limit)

## Troubleshooting

1. Check webhook logs:
```bash
kubectl logs -l app=dns-config-webhook
```

2. Verify webhook configuration:
```bash
kubectl get mutatingwebhookconfigurations
```

3. Check if the webhook service is running:
```bash
kubectl get pods -l app=dns-config-webhook
```

## Building from Source

1. Requirements:
   - Go 1.23.6 or higher
   - Docker (for building container images)

2. Build the binary:
```bash
go build -o webhook ./cmd/webhook
```

3. Build Docker image:
```bash
docker buildx build  --debug \
  --platform linux/amd64,linux/arm64 \
  -t webhook-dns-config:0.1 \
  --push .
```

## License

This project is licensed under the terms of the LICENSE file included in the repository.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
