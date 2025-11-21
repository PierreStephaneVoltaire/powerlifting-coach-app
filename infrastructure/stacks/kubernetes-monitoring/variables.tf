variable "domain_name" {
  description = "Domain name for the application"
  type        = string
}

variable "stopped" {
  description = "Whether cluster is stopped"
  type        = bool
  default     = false
}
