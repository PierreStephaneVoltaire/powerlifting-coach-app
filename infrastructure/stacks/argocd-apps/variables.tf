variable "project_name" {
  description = "Project name"
  type        = string
  default     = "nolift"
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
