variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "powerlifting-coach"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "region" {
  description = "DigitalOcean region"
  type        = string
  default     = "tor1"
}

variable "node_size" {
  description = "Size of Kubernetes nodes"
  type        = string
  default     = "s-2vcpu-2gb"
}

variable "node_count" {
  description = "Number of Kubernetes nodes"
  type        = number
  default     = 1
}

variable "kubernetes_version" {
  description = "Kubernetes cluster version"
  type        = string
  default     = "1.33.1-do.5"
}

variable "spaces_bucket_name" {
  description = "DigitalOcean Spaces bucket name"
  type        = string
  default     = "powerlifting-coach-videos"
}
