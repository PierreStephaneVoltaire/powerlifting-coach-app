output "cluster_id" {
  description = "Kubernetes cluster ID"
  value       = digitalocean_kubernetes_cluster.k8s.id
}

output "cluster_name" {
  description = "Kubernetes cluster name"
  value       = digitalocean_kubernetes_cluster.k8s.name
}

output "cluster_endpoint" {
  description = "Kubernetes cluster endpoint"
  value       = digitalocean_kubernetes_cluster.k8s.endpoint
}

output "cluster_ca_certificate" {
  description = "Kubernetes cluster CA certificate"
  value       = digitalocean_kubernetes_cluster.k8s.kube_config[0].cluster_ca_certificate
  sensitive   = true
}

output "cluster_token" {
  description = "Kubernetes cluster token"
  value       = digitalocean_kubernetes_cluster.k8s.kube_config[0].token
  sensitive   = true
}

output "vpc_id" {
  description = "VPC ID"
  value       = digitalocean_vpc.this.id
}

output "spaces_bucket_name" {
  description = "Spaces bucket name"
  value       = digitalocean_spaces_bucket.videos.name
}

output "spaces_bucket_endpoint" {
  description = "Spaces bucket endpoint"
  value       = "https://${var.region}.digitaloceanspaces.com"
}

output "spaces_access_key_id" {
  description = "Spaces access key ID"
  value       = digitalocean_spaces_key.default.access_key
  sensitive   = true
}

output "spaces_secret_access_key" {
  description = "Spaces secret access key"
  value       = digitalocean_spaces_key.default.secret_key
  sensitive   = true
}

output "region" {
  description = "Deployment region"
  value       = var.region
}

output "postgres_password" {
  description = "PostgreSQL database password"
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
