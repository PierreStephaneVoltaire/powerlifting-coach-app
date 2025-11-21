variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ca-central-1"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "nolift"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "domain_name" {
  description = "Domain name"
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version (K3s format)"
  type        = string
  default     = "v1.31.13-k3s1"
}

variable "worker_desired_capacity" {
  description = "Desired number of worker nodes"
  type        = number
  default     = 3
}

variable "stopped" {
  description = "Whether cluster is stopped"
  type        = bool
  default     = false
}

variable "admin_ips" {
  description = "Admin IPs for SSH access"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}
