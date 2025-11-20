variable "project_name" {
  description = "Project name"
  type        = string
}

variable "argocd_namespace" {
  description = "ArgoCD namespace name"
  type        = string
}

variable "app_namespace" {
  description = "Application namespace name"
  type        = string
}

variable "deploy_frontend" {
  description = "Deploy frontend application"
  type        = bool
  default     = true
}

variable "deploy_datalayer" {
  description = "Deploy datalayer application"
  type        = bool
  default     = true
}

variable "deploy_backend" {
  description = "Deploy backend application"
  type        = bool
  default     = false
}
