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
