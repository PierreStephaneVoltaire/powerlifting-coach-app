#!/bin/bash
# Start workloads after stopping
# Restores LoadBalancer, helm releases, and scales nodes back up

set -e

cd "$(dirname "$0")/../infrastructure"

echo "=== Starting workloads ==="
echo "This will:"
echo "  - Reinstall all helm releases"
echo "  - Scale spot node pool to 1"
echo "  - Create LoadBalancer"
echo "  - Restore application DNS records"
echo ""

# Note: This uses existing terraform.tfvars for kubernetes_resources_enabled, argocd_resources_enabled, etc.
terraform apply -var="stopped=false"

echo ""
echo "✅ Workloads started!"
echo ""
echo "Your applications should be accessible within 5-10 minutes:"
echo "  - Frontend: https://app.nolift.training"
echo "  - ArgoCD: https://argocd.nolift.training"
echo "  - Grafana: https://grafana.nolift.training"
