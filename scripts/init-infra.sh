#!/bin/bash
set -e

cd "$(dirname "$0")/../infrastructure"

echo "=== Initializing Terraform ==="
terraform init

echo ""
echo "=== Creating base infrastructure ==="
echo "This will create:"
echo "  - VPC with public subnets (no NAT gateway)"
echo "  - k3s control plane (3 spot instances behind NLB)"
echo "  - k3s workers (2 spot instances, auto-scaling 1-5)"
echo "  - Route53 hosted zone"
echo "  - S3 bucket for videos"
echo "  - SES for email"
echo ""

terraform apply \
  -var="kubernetes_resources_enabled=false" \
  -var="argocd_resources_enabled=false"

echo ""
echo "✅ Base infrastructure created!"
echo ""
echo "Next steps:"
echo "  1. Configure nameservers in your domain registrar (Namecheap):"
terraform output aws_nameservers
echo ""
echo "  2. Wait for DNS propagation (5-30 minutes)"
echo ""
echo "  3. Once DNS propagates, deploy Kubernetes add-ons:"
echo "     terraform apply -var='kubernetes_resources_enabled=true'"
echo ""
echo "  4. Verify SES domain in AWS:"
echo "     - Go to AWS Console -> SES -> Verified identities"
echo "     - Check domain verification status"
echo ""
echo "  5. Run ./enable-argo.sh to enable ArgoCD and continuous deployment"
