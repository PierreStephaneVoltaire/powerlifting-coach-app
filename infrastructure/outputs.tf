output "cluster_id" {
  description = "Kubernetes cluster ID"
  value       = azurerm_kubernetes_cluster.k8s.id
}

output "cluster_name" {
  description = "Kubernetes cluster name"
  value       = azurerm_kubernetes_cluster.k8s.name
}

output "cluster_endpoint" {
  description = "Kubernetes cluster endpoint"
  value       = azurerm_kubernetes_cluster.k8s.kube_config[0].host
  sensitive = true
}

output "cluster_ca_certificate" {
  description = "Kubernetes cluster CA certificate"
  value       = azurerm_kubernetes_cluster.k8s.kube_config[0].cluster_ca_certificate
  sensitive   = true
}

output "resource_group_name" {
  description = "Resource group name"
  value       = azurerm_resource_group.this.name
}

output "storage_account_name" {
  description = "Storage account name"
  value       = azurerm_storage_account.videos.name
}

output "storage_container_name" {
  description = "Storage container name"
  value       = azurerm_storage_container.videos.name
}

output "storage_account_endpoint" {
  description = "Storage account blob endpoint"
  value       = azurerm_storage_account.videos.primary_blob_endpoint
}

output "storage_access_key" {
  description = "Storage account access key"
  value       = azurerm_storage_account.videos.primary_access_key
  sensitive   = true
}

output "storage_connection_string" {
  description = "Storage account connection string"
  value       = azurerm_storage_account.videos.primary_connection_string
  sensitive   = true
}

output "region" {
  description = "Deployment region"
  value       = var.region
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
  description = "ArgoCD admin password (get with: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)"
  value       = "Run: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
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

output "dns_zone_id" {
  description = "Azure DNS zone ID"
  value       =  azurerm_dns_zone.main[0].id 
}

output "dns_zone_name" {
  description = "Azure DNS zone name"
  value       =  azurerm_dns_zone.main[0].name 
}

output "azure_nameservers" {
  description = "Azure DNS nameservers - Configure these in Route 53 as NS records"
  value       =  azurerm_dns_zone.main[0].name_servers 
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


