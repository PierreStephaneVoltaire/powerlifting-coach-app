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
  value     = try(aws_iam_access_key.s3_videos.id, null)
  sensitive = true
}

output "s3_videos_secret_key" {
  value     = try(aws_iam_access_key.s3_videos.secret, null)
  sensitive = true
}

output "region" {
  value = var.aws_region
}

output "postgres_password" {
  value     = module.kubernetes_base.postgres_password
  sensitive = true
}

output "rabbitmq_password" {
  value     = module.kubernetes_base.rabbitmq_password
  sensitive = true
}

output "keycloak_client_secret" {
  value     = module.kubernetes_base.keycloak_client_secret
  sensitive = true
}

output "keycloak_admin_password" {
  value     = module.kubernetes_base.keycloak_admin_password
  sensitive = true
}

output "argocd_url" {
  value = "https://argocd.${var.domain_name}"
}

output "frontend_url" {
  value = "https://app.${var.domain_name}"
}

output "api_url" {
  value = "https://api.${var.domain_name}"
}

output "auth_url" {
  value = "https://auth.${var.domain_name}"
}

output "grafana_url" {
  value = "https://grafana.${var.domain_name}"
}

output "grafana_admin_password" {
  value     = module.kubernetes_base.grafana_admin_password
  sensitive = true
}

output "prometheus_url" {
  value = "https://prometheus.${var.domain_name}"
}

output "loki_url" {
  value = "https://loki.${var.domain_name}"
}

output "rabbitmq_management_url" {
  value = "https://rabbitmq.${var.domain_name}"
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
  value     = try(aws_iam_access_key.ses_smtp.id, null)
  sensitive = true
}

output "ses_smtp_password" {
  value     = try(aws_iam_access_key.ses_smtp.ses_smtp_password_v4, null)
  sensitive = true
}

output "rancher_admin" {
  value     = random_password.rancher_admin
  sensitive = true
}

output "rancher_cluster_id" {
  description = "ID of the Rancher-managed cluster"
  value       = module.rancher_cluster.cluster_id
}

output "rancher_cluster_name" {
  description = "Name of the Rancher-managed cluster"
  value       = module.rancher_cluster.cluster_name
}

output "rancher_admin_token" {
  description = "Rancher admin API token"
  value       = module.rancher_cluster.admin_token
  sensitive   = true
}

output "cluster_kubeconfig" {
  description = "Kubeconfig for the cluster (use terraform output -raw cluster_kubeconfig)"
  value       = module.rancher_cluster.kubeconfig
  sensitive   = true
}