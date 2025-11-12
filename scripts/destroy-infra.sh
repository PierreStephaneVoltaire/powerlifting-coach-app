#!/bin/bash
set -e

cd "$(dirname "$0")/../infrastructure"

echo "⚠️  WARNING: This will destroy ALL infrastructure ⚠️"
echo ""
echo "This includes:"
echo "  - k3s cluster (control plane + workers)"
echo "  - Network Load Balancer"
echo "  - VPC and subnets"
echo "  - Route53 hosted zone and DNS records"
echo "  - S3 buckets (all videos will be deleted!)"
echo "  - SES email configuration"
echo "  - All EC2 instances"
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

echo "Stage 1: Destroying Kubernetes resources..."
terraform destroy \
  -target=helm_release.nginx_ingress \
  -target=helm_release.cert_manager \
  -target=helm_release.ebs_csi_driver \
  -target=helm_release.metrics_server \
  -target=helm_release.external_dns \
  -target=helm_release.openwebui \
  -auto-approve || true

echo ""
echo "Stage 2: Destroying everything else..."
terraform destroy

echo ""
echo "✅ Infrastructure destroyed"
