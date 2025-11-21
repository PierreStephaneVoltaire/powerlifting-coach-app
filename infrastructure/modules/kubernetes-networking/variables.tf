variable "domain_name" {
  description = "Domain name for the application"
  type        = string
}

variable "cluster_name" {
  description = "Cluster name"
  type        = string
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
