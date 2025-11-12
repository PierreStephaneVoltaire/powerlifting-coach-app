# Route53 DNS Configuration for AWS

# Create Route53 hosted zone
resource "aws_route53_zone" "main" {
  name = var.domain_name

  tags = {
    Name        = "${local.cluster_name}-hosted-zone"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Data source to get the ALB hostname from NGINX Ingress Controller
# This will be populated after the ingress controller creates the ALB
data "kubernetes_service" "nginx_ingress" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  metadata {
    name      = "ingress-nginx-controller"
    namespace = "ingress-nginx"
  }

  depends_on = [
    helm_release.nginx_ingress
  ]
}

locals {
  # Get the ALB hostname from the ingress controller LoadBalancer
  dns_lb_hostname = var.kubernetes_resources_enabled && !var.stopped ? (
    length(data.kubernetes_service.nginx_ingress) > 0 && length(data.kubernetes_service.nginx_ingress[0].status[0].load_balancer[0].ingress) > 0 ?
    data.kubernetes_service.nginx_ingress[0].status[0].load_balancer[0].ingress[0].hostname : null
  ) : null
}

# Note: external-dns will automatically manage these records
# We only create the zone here and let external-dns handle the records
# But we can create a wildcard record manually if needed

# Wildcard A record pointing to ALB (managed by external-dns)
# external-dns will create records automatically based on Ingress resources
# This is just a fallback/example
resource "aws_route53_record" "wildcard" {
  count   = var.kubernetes_resources_enabled && !var.stopped && local.dns_lb_hostname != null ? 1 : 0
  zone_id = aws_route53_zone.main.zone_id
  name    = "*.${var.domain_name}"
  type    = "CNAME"
  ttl     = 300
  records = [local.dns_lb_hostname]
}

# Root domain record
resource "aws_route53_record" "root" {
  count   = var.kubernetes_resources_enabled && !var.stopped && local.dns_lb_hostname != null ? 1 : 0
  zone_id = aws_route53_zone.main.zone_id
  name    = var.domain_name
  type    = "CNAME"
  ttl     = 300
  records = [local.dns_lb_hostname]
}
