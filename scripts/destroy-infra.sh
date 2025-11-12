#!/bin/bash
# Destroy ALL infrastructure
# WARNING: This deletes everything including data!

set -e

cd "$(dirname "$0")/../infrastructure"

echo "⚠️  WARNING: This will destroy ALL infrastructure ⚠️"
echo ""
echo "This includes:"
echo "  - AKS cluster and all applications"
echo "  - DNS zone and records"
echo "  - Storage account (all videos will be deleted!)"
echo "  - Communication services"
echo "  - Load balancers"
echo ""
echo "This action cannot be undone!"
echo ""
read -p "Type 'destroy' to confirm: " confirm

if [ "$confirm" != "destroy" ]; then
    echo "Cancelled."
    exit 1
fi

echo ""
echo "=== Destroying infrastructure ==="

# Destroy in stages to avoid dependency issues
echo "Stage 1: Destroying Kubernetes resources..."
terraform destroy \
  -target=helm_release.nginx_ingress \
  -target=helm_release.argocd \
  -target=helm_release.argocd_image_updater \
  -target=helm_release.cert_manager \
  -target=helm_release.openwebui \
  -auto-approve || true

echo ""
echo "Stage 2: Destroying everything else..."
terraform destroy

echo ""
echo "✅ Infrastructure destroyed"
