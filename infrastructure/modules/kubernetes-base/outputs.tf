output "app_namespace" {
  description = "Application namespace name"
  value       = kubernetes_namespace.app.metadata[0].name
}

output "grafana_admin_password" {
  description = "Grafana admin password"
  value       = random_password.grafana_admin_password.result
  sensitive   = true
}
