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
  description = "Domain name for the application"
  type        = string
}

variable "stopped" {
  description = "Whether cluster is stopped"
  type        = bool
  default     = false
}

