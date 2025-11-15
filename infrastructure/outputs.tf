output "cluster_name" {
  value = local.cluster_name
}

output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "eks_cluster_id" {
  value = module.eks.cluster_id
}

output "eks_cluster_arn" {
  value = module.eks.cluster_arn
}

output "eks_cluster_version" {
  value = module.eks.cluster_version
}

output "eks_node_group_small_id" {
  value = module.eks.eks_managed_node_groups["small"].node_group_id
}

output "vpc_id" {
  value = aws_vpc.main.id
}

output "public_subnet_ids" {
  value = aws_subnet.public[*].id
}

output "s3_videos_bucket" {
  value = aws_s3_bucket.videos.id
}

output "s3_videos_bucket_arn" {
  value = aws_s3_bucket.videos.arn
}

output "s3_videos_endpoint" {
  value = "https://${aws_s3_bucket.videos.bucket_regional_domain_name}"
}

output "region" {
  value = var.aws_region
}

output "postgres_password" {
  value     = var.kubernetes_resources_enabled ? random_password.postgres_password[0].result : null
  sensitive = true
}

output "rabbitmq_password" {
  value     = var.kubernetes_resources_enabled ? random_password.rabbitmq_password[0].result : null
  sensitive = true
}

output "keycloak_client_secret" {
  value     = var.kubernetes_resources_enabled ? random_password.keycloak_client_secret[0].result : null
  sensitive = true
}

output "keycloak_admin_password" {
  value     = var.kubernetes_resources_enabled ? random_password.keycloak_admin_password[0].result : null
  sensitive = true
}

output "argocd_url" {
  value = var.kubernetes_resources_enabled ? "https://argocd.${var.domain_name}" : null
}

output "frontend_url" {
  value = var.kubernetes_resources_enabled ? "https://app.${var.domain_name}" : null
}

output "api_url" {
  value = var.kubernetes_resources_enabled ? "https://api.${var.domain_name}" : null
}

output "auth_url" {
  value = var.kubernetes_resources_enabled ? "https://auth.${var.domain_name}" : null
}

output "grafana_url" {
  value = var.kubernetes_resources_enabled ? "https://grafana.${var.domain_name}" : null
}

output "grafana_admin_password" {
  value     = var.kubernetes_resources_enabled ? random_password.grafana_admin_password[0].result : null
  sensitive = true
}

output "prometheus_url" {
  value = var.kubernetes_resources_enabled ? "https://prometheus.${var.domain_name}" : null
}

output "loki_url" {
  value = var.kubernetes_resources_enabled ? "https://loki.${var.domain_name}" : null
}

output "rabbitmq_management_url" {
  value = var.kubernetes_resources_enabled ? "https://rabbitmq.${var.domain_name}" : null
}

output "rabbitmq_management_username" {
  value = "admin"
}

output "openwebui_url" {
  value = var.kubernetes_resources_enabled ? "https://openwebui.${var.domain_name}" : null
}

output "route53_zone_id" {
  value = aws_route53_zone.main.zone_id
}

output "route53_zone_name" {
  value = aws_route53_zone.main.name
}

output "aws_nameservers" {
  value = aws_route53_zone.main.name_servers
}

output "domain_urls" {
  value = {
    frontend = "https://app.${var.domain_name}"
    api      = "https://api.${var.domain_name}"
    auth     = "https://auth.${var.domain_name}"
    grafana  = "https://grafana.${var.domain_name}"
    argocd   = "https://argocd.${var.domain_name}"
  }
}

output "ses_smtp_endpoint" {
  value = "email-smtp.${var.aws_region}.amazonaws.com"
}

output "ses_smtp_username" {
  value     = var.kubernetes_resources_enabled ? aws_iam_access_key.ses_smtp[0].id : null
  sensitive = true
}

output "ses_smtp_password" {
  value     = var.kubernetes_resources_enabled ? aws_iam_access_key.ses_smtp[0].ses_smtp_password_v4 : null
  sensitive = true
}
