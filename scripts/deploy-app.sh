#!/bin/bash
# Manually trigger application deployment via ArgoCD
# Normally happens automatically on git push

set -e

cd "$(dirname "$0")/../infrastructure"
export KUBECONFIG=./kubeconfig.yaml

echo "=== Syncing ArgoCD application ==="
argocd app sync nolift --grpc-web --insecure

echo ""
echo "=== Watching deployment ==="
argocd app wait nolift --grpc-web --insecure --timeout 600

echo ""
echo "✅ Application deployed!"
echo ""
echo "Frontend: https://app.nolift.training"
echo "API: https://api.nolift.training"
