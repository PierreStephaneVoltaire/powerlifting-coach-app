output "cluster_name" {
  value = local.cluster_name
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

output "s3_videos_bucket_domain" {
  value = aws_s3_bucket.videos.bucket_regional_domain_name
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

output "ami_id" {
  description = "AMI ID for cluster nodes"
  value       = data.aws_ami.amazon_linux_2.id
}

output "rancher_server_sg_id" {
  description = "Rancher server security group ID"
  value       = aws_security_group.rancher_server.id
}

output "rancher_server_fqdn" {
  description = "Rancher server FQDN"
  value       = aws_route53_record.rancher_server.fqdn
}

output "rancher_admin_password" {
  description = "Rancher admin password"
  value       = random_password.rancher_admin.result
  sensitive   = true
}
