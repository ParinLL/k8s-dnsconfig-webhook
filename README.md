# k8s-dnsconfig-webhook
Auto modify Pod dnsConfig with MutatingWebhookConfiguration


docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t dokfish/k8s-dnsconfig-webhook:0.1 \
  --push .
