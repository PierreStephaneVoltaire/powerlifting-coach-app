variable "domain_name" {
  description = "Domain name for the application"
  type        = string
}

variable "app_namespace" {
  description = "Application namespace name"
  type        = string
}

variable "grafana_admin_password" {
  description = "Grafana admin password"
  type        = string
  sensitive   = true
}

variable "stopped" {
  description = "Whether cluster is stopped"
  type        = bool
  default     = false
}

variable "kube_host" {
  description = "Kubernetes API server host"
  type        = string
}

variable "kube_token" {
  description = "Kubernetes API server token"
  type        = string
  sensitive   = true
}
