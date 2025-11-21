output "frontend_app_name" {
  description = "Frontend ArgoCD application name"
  value       = var.deploy_frontend ? kubernetes_manifest.app_frontend[0].manifest.metadata.name : null
}

output "datalayer_app_name" {
  description = "Datalayer ArgoCD application name"
  value       = var.deploy_datalayer ? kubernetes_manifest.app_datalayer[0].manifest.metadata.name : null
}

output "backend_app_name" {
  description = "Backend ArgoCD application name"
  value       = var.deploy_backend ? kubernetes_manifest.app_backend[0].manifest.metadata.name : null
}
