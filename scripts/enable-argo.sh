#!/bin/bash
set -e

cd "$(dirname "$0")/../infrastructure"

echo "=== Enabling ArgoCD Applications (requires ArgoCD helm chart installed) ==="
terraform apply \
  -var="kubernetes_resources_enabled=true" \
  -var="argocd_resources_enabled=true"

echo ""
echo "=== Getting ArgoCD admin password ==="
export KUBECONFIG=./kubeconfig.yaml
ARGOCD_PASSWORD=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)

echo ""
echo "✅ ArgoCD enabled!"
echo ""
echo "ArgoCD URL: https://argocd.nolift.training"
echo "Username: admin"
echo "Password: $ARGOCD_PASSWORD"
echo ""
echo "Your application will now auto-deploy from GitHub on every push."
