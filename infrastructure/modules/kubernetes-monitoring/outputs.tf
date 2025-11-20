output "prometheus_release_name" {
  description = "Prometheus release name"
  value       = var.stopped ? null : helm_release.kube_prometheus_stack[0].name
}

output "loki_release_name" {
  description = "Loki release name"
  value       = var.stopped ? null : helm_release.loki[0].name
}
