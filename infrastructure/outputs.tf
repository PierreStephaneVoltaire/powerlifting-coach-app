output "cluster_name" {
  value = local.cluster_name
}

output "cluster_endpoint" {
  description = "Kubernetes API endpoint (from Rancher kubeconfig)"
  value       = var.rancher_cluster_enabled ? "See cluster_kubeconfig output" : "Rancher Server: https://${aws_eip.rancher.public_ip}"
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

output "s3_videos_access_key" {
  value     = try(aws_iam_access_key.s3_videos[0].id, null)
  sensitive = true
}

output "s3_videos_secret_key" {
  value     = try(aws_iam_access_key.s3_videos[0].secret, null)
  sensitive = true
}

output "region" {
  value = var.aws_region
}

output "postgres_password" {
  value     = var.kubernetes_resources_enabled ? module.kubernetes_base[0].postgres_password : null
  sensitive = true
}

output "rabbitmq_password" {
  value     = var.kubernetes_resources_enabled ? module.kubernetes_base[0].rabbitmq_password : null
  sensitive = true
}

output "keycloak_client_secret" {
  value     = var.kubernetes_resources_enabled ? module.kubernetes_base[0].keycloak_client_secret : null
  sensitive = true
}

output "keycloak_admin_password" {
  value     = var.kubernetes_resources_enabled ? module.kubernetes_base[0].keycloak_admin_password : null
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
  value     = var.kubernetes_resources_enabled ? module.kubernetes_base[0].grafana_admin_password : null
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
  value     = try(aws_iam_access_key.ses_smtp[0].id, null)
  sensitive = true
}

output "ses_smtp_password" {
  value     = try(aws_iam_access_key.ses_smtp[0].ses_smtp_password_v4, null)
  sensitive = true
}

output "rancher_admin" {
  value     = random_password.rancher_admin
  sensitive = true
}

output "rancher_cluster_id" {
  description = "ID of the Rancher-managed cluster"
  value       = var.rancher_cluster_enabled ? module.rancher_cluster[0].cluster_id : null
}

output "rancher_cluster_name" {
  description = "Name of the Rancher-managed cluster"
  value       = var.rancher_cluster_enabled ? module.rancher_cluster[0].cluster_name : null
}

output "rancher_admin_token" {
  description = "Rancher admin API token"
  value       = var.rancher_cluster_enabled ? module.rancher_cluster[0].admin_token : null
  sensitive   = true
}

output "cluster_kubeconfig" {
  description = "Kubeconfig for the cluster (use terraform output -raw cluster_kubeconfig)"
  value       = var.rancher_cluster_enabled ? module.rancher_cluster[0].kubeconfig : null
  sensitive   = true
}