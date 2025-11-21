output "app_namespace" {
  description = "Application namespace name"
  value       = kubernetes_namespace.app.metadata[0].name
}

output "grafana_admin_password" {
  description = "Grafana admin password"
  value       = random_password.grafana_admin_password.result
  sensitive   = true
}

output "postgres_password" {
  description = "PostgreSQL password"
  value       = random_password.postgres_password.result
  sensitive   = true
}

output "rabbitmq_password" {
  description = "RabbitMQ password"
  value       = random_password.rabbitmq_password.result
  sensitive   = true
}

output "keycloak_client_secret" {
  description = "Keycloak client secret"
  value       = random_password.keycloak_client_secret.result
  sensitive   = true
}

output "keycloak_admin_password" {
  description = "Keycloak admin password"
  value       = random_password.keycloak_admin_password.result
  sensitive   = true
}
