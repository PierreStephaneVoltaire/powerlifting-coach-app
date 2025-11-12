#!/bin/bash
# Stop all workloads to save costs
# Keeps 1 default node, destroys LoadBalancer and all applications
# Cost when stopped: ~$10/month

set -e

cd "$(dirname "$0")/../infrastructure"

echo "=== Stopping workloads ==="
echo "This will:"
echo "  - Destroy all helm releases (nginx, ArgoCD, cert-manager, etc.)"
echo "  - Scale spot node pool to 0"
echo "  - Delete LoadBalancer (~$20-30/month savings)"
echo "  - Remove application DNS records"
echo "  - Keep 1 default node running (~$10/month)"
echo "  - Keep DNS zone and email services"
echo ""

terraform apply -var="stopped=true"

echo ""
echo "✅ Workloads stopped!"
echo "Cost: ~$10/month (1 Standard_B2s node)"
echo ""
echo "To restart: ./start-workload.sh"
