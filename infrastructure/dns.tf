# Route53 DNS Configuration for AWS

# Create Route53 hosted zone for your Namecheap domain
# After creating this, you'll need to update your Namecheap nameservers
# to point to the AWS Route53 nameservers shown in terraform output
resource "aws_route53_zone" "main" {
  name = var.domain_name

  tags = {
    Name        = "${local.cluster_name}-hosted-zone"
    Environment = var.environment
    Project     = var.project_name
  }
}

# DNS records are automatically managed by external-dns (see kubernetes-addons.tf)
# external-dns watches for Kubernetes Services and Ingresses with the annotation:
#   external-dns.alpha.kubernetes.io/hostname
#
# The nginx ingress controller LoadBalancer service has this annotation configured
# to automatically create DNS records for *.${var.domain_name}
#
# This approach is better than manual DNS records because:
# 1. No circular dependencies - external-dns creates records after resources exist
# 2. Automatic updates - DNS records update when services change
# 3. Multiple records - Can handle multiple services/ingresses automatically
