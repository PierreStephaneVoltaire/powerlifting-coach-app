output "cluster_id" {
  description = "ID of the Rancher-managed cluster"
  value       = rancher2_cluster_v2.main.id
}

output "cluster_name" {
  description = "Name of the Rancher-managed cluster"
  value       = rancher2_cluster_v2.main.name
}

output "admin_token" {
  description = "Rancher admin API token"
  value       = rancher2_bootstrap.admin.token
  sensitive   = true
}

output "kubeconfig" {
  description = "Kubeconfig for the cluster"
  value       = rancher2_cluster_v2.main.kube_config
  sensitive   = true
}

output "kube_host" {
  description = "Kubernetes API server host"
  value       = yamldecode(rancher2_cluster_v2.main.kube_config).clusters[0].cluster.server
}

output "kube_token" {
  description = "Kubernetes API server token"
  value       = yamldecode(rancher2_cluster_v2.main.kube_config).users[0].user.token
  sensitive   = true
}
