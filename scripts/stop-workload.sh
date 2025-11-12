#!/bin/bash
set -e

cd "$(dirname "$0")/../infrastructure"

echo "=== Stopping workloads ==="
echo "This will:"
echo "  - Destroy all helm releases (nginx, ArgoCD, cert-manager, etc.)"
echo "  - Scale worker nodes to 0"
echo "  - Delete LoadBalancer (~$15-20/month savings)"
echo "  - Remove application DNS records"
echo "  - Keep control plane running (3 spot instances ~$10-15/month)"
echo "  - Keep DNS zone and email services"
echo ""

terraform apply -var="stopped=true"

echo ""
echo "✅ Workloads stopped!"
echo "Cost: ~$25-30/month (control plane + NLB + DNS)"
echo ""
echo "To restart: ./start-workload.sh"
