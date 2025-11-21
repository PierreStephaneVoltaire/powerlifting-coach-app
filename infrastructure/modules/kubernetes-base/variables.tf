variable "environment" {
  description = "Environment name"
  type        = string
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
}

variable "s3_access_key_id" {
  description = "S3 access key ID"
  type        = string
  sensitive   = true
}

variable "s3_secret_access_key" {
  description = "S3 secret access key"
  type        = string
  sensitive   = true
}

variable "s3_bucket_domain" {
  description = "S3 bucket domain name"
  type        = string
}

variable "s3_bucket_id" {
  description = "S3 bucket ID"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "ses_smtp_endpoint" {
  description = "SES SMTP endpoint"
  type        = string
}

variable "ses_smtp_username" {
  description = "SES SMTP username"
  type        = string
  sensitive   = true
}

variable "ses_smtp_password" {
  description = "SES SMTP password"
  type        = string
  sensitive   = true
}

variable "google_oauth_client_id" {
  description = "Google OAuth client ID"
  type        = string
  sensitive   = true
}

variable "google_oauth_client_secret" {
  description = "Google OAuth client secret"
  type        = string
  sensitive   = true
}

variable "stopped" {
  description = "Whether cluster is stopped"
  type        = bool
  default     = false
}

variable "project_root" {
  description = "Project root path for local file resources"
  type        = string
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
