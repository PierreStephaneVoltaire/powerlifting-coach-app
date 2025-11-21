output "monitoring_namespace" {
  description = "Monitoring namespace"
  value       = var.stopped ? null : helm_release.kube_prometheus_stack[0].namespace
}
