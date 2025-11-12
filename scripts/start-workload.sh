#!/bin/bash
set -e

cd "$(dirname "$0")/../infrastructure"

echo "=== Starting workloads ==="
echo "This will:"
echo "  - Reinstall all helm releases"
echo "  - Scale worker nodes back to desired capacity"
echo "  - Create LoadBalancer"
echo "  - Restore application DNS records"
echo ""

terraform apply -var="stopped=false"

echo ""
echo "✅ Workloads started!"
echo ""
echo "Your applications should be accessible within 5-10 minutes:"
echo "  - Frontend: https://app.nolift.training"
echo "  - ArgoCD: https://argocd.nolift.training"
echo "  - Grafana: https://grafana.nolift.training"
