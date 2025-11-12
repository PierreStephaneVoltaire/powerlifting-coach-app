# k3s Cluster Outputs

output "cluster_name" {
  description = "k3s cluster name"
  value       = local.cluster_name
}

output "cluster_endpoint" {
  description = "k3s API server endpoint (via NLB)"
  value       = "https://${aws_lb.control_plane.dns_name}:6443"
}

output "nlb_dns_name" {
  description = "Network Load Balancer DNS name for control plane"
  value       = aws_lb.control_plane.dns_name
}

output "control_plane_asg_name" {
  description = "Control plane Auto Scaling Group name"
  value       = aws_autoscaling_group.control_plane.name
}

output "worker_asg_name" {
  description = "Worker Auto Scaling Group name"
  value       = aws_autoscaling_group.worker.name
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = aws_subnet.public[*].id
}

# S3 Storage Outputs
output "s3_videos_bucket" {
  description = "S3 bucket name for videos"
  value       = aws_s3_bucket.videos.id
}

output "s3_videos_bucket_arn" {
  description = "S3 bucket ARN for videos"
  value       = aws_s3_bucket.videos.arn
}

output "s3_videos_endpoint" {
  description = "S3 bucket endpoint URL"
  value       = "https://${aws_s3_bucket.videos.bucket_regional_domain_name}"
}

output "s3_config_bucket" {
  description = "S3 bucket for k3s configuration"
  value       = aws_s3_bucket.k3s_config.id
}

output "region" {
  description = "AWS deployment region"
  value       = var.aws_region
}

output "postgres_password" {
  description = "PostgreSQL database password"
  value       = var.kubernetes_resources_enabled ? random_password.postgres_password[0].result : "not-yet-generated"
  sensitive   = true
}

output "rabbitmq_password" {
  description = "RabbitMQ password"
  value       = var.kubernetes_resources_enabled ? random_password.rabbitmq_password[0].result : "not-yet-generated"
  sensitive   = true
}

output "keycloak_client_secret" {
  description = "Keycloak client secret"
  value       = var.kubernetes_resources_enabled ? random_password.keycloak_client_secret[0].result : "not-yet-generated"
  sensitive   = true
}

output "keycloak_admin_password" {
  description = "Keycloak admin password"
  value       = var.kubernetes_resources_enabled ? random_password.keycloak_admin_password[0].result : "not-yet-generated"
  sensitive   = true
}

output "argocd_url" {
  description = "ArgoCD UI URL"
  value       = var.kubernetes_resources_enabled ? "https://argocd.nolift.training" : "not-yet-available"
}

output "argocd_admin_password" {
  description = "ArgoCD admin password"
  value       = var.kubernetes_resources_enabled ? "stored-in-k8s-secret" : "not-yet-available"
}

output "frontend_url" {
  description = "Frontend application URL"
  value       = var.kubernetes_resources_enabled ? "https://app.nolift.training" : "not-yet-available"
}

output "api_url" {
  description = "API base URL"
  value       = var.kubernetes_resources_enabled ? "https://api.nolift.training" : "not-yet-available"
}

output "auth_url" {
  description = "Keycloak authentication URL"
  value       = var.kubernetes_resources_enabled ? "https://auth.nolift.training" : "not-yet-available"
}

output "keycloak_url" {
  description = "Keycloak admin console URL"
  value       = var.kubernetes_resources_enabled ? "https://auth.nolift.training" : "not-yet-available"
}

output "grafana_url" {
  description = "Grafana dashboard URL"
  value       = var.kubernetes_resources_enabled ? "https://grafana.nolift.training" : "not-yet-available"
}

output "grafana_admin_password" {
  description = "Grafana admin password"
  value       = var.kubernetes_resources_enabled ? random_password.grafana_admin_password[0].result : "not-yet-generated"
  sensitive   = true
}

output "prometheus_url" {
  description = "Prometheus metrics URL"
  value       = var.kubernetes_resources_enabled ? "https://prometheus.nolift.training" : "not-yet-available"
}

output "loki_url" {
  description = "Loki logs URL"
  value       = var.kubernetes_resources_enabled ? "https://loki.nolift.training" : "not-yet-available"
}

output "rabbitmq_management_url" {
  description = "RabbitMQ management console URL"
  value       = var.kubernetes_resources_enabled ? "https://rabbitmq.nolift.training" : "not-yet-available"
}

output "rabbitmq_management_username" {
  description = "RabbitMQ management console username"
  value       = "admin"
}

output "openwebui_url" {
  description = "OpenWebUI chat interface URL"
  value       = var.kubernetes_resources_enabled ? "https://openwebui.nolift.training" : "not-yet-available"
}

# DNS Outputs

output "route53_zone_id" {
  description = "Route53 hosted zone ID"
  value       = aws_route53_zone.main.zone_id
}

output "route53_zone_name" {
  description = "Route53 hosted zone name"
  value       = aws_route53_zone.main.name
}

output "aws_nameservers" {
  description = "AWS Route53 nameservers - Update these in your domain registrar (Namecheap)"
  value       = aws_route53_zone.main.name_servers
}

output "domain_urls" {
  description = "Application URLs with custom domain (if configured)"
  value = {
    frontend = "https://app.${var.domain_name}"
    api      = "https://api.${var.domain_name}"
    auth     = "https://auth.${var.domain_name}"
    grafana  = "https://grafana.${var.domain_name}"
    argocd   = "https://argocd.${var.domain_name}"
  } 
}


