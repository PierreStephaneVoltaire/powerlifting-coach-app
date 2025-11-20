output "argocd_namespace" {
  description = "ArgoCD namespace name"
  value       = kubernetes_namespace.argocd.metadata[0].name
}

output "argocd_release_name" {
  description = "ArgoCD release name"
  value       = var.stopped ? null : helm_release.argocd[0].name
}
