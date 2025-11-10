variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
  sensitive   = true
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "coachpotato"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "region" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "node_size" {
  description = "Azure VM size for Kubernetes nodes"
  type        = string
  default     = "Standard_B2s"
}

variable "node_count" {
  description = "Number of Kubernetes nodes"
  type        = number
  default     = 1
}

variable "kubernetes_version" {
  description = "Kubernetes cluster version"
  type        = string
  default     = "1.28"
}

variable "storage_container_name" {
  description = "Azure Storage container name for videos"
  type        = string
  default     = "coachpotato-videos"
}

variable "kubernetes_resources_enabled" {
  description = "Enable Kubernetes resources (namespaces, secrets, ArgoCD). Set to false for initial cluster creation, then true for second apply"
  type        = bool
  default     = false
}

variable "argocd_resources_enabled" {
  description = "Enable ArgoCD Application and RBAC resources. Requires ArgoCD CRDs to be installed. Set to false initially, then true after ArgoCD is fully deployed"
  type        = bool
  default     = false
}

variable "openai_api_key" {
  description = "OpenAI API key for LLM access in OpenWebUI"
  type        = string
  sensitive   = true
  default     = ""
}

variable "litellm_endpoint" {
  description = "LiteLLM endpoint URL for OpenWebUI to connect to various LLM providers"
  type        = string
  default     = "https://api.openai.com/v1"
}

variable "domain_name" {
  description = "Domain name for the application (e.g., coachpotato.app)"
  type        = string
  default     = "localhost"
}

variable "ses_smtp_host" {
  description = "AWS SES SMTP host (e.g., email-smtp.us-east-1.amazonaws.com)"
  type        = string
  default     = "email-smtp.us-east-1.amazonaws.com"
}

variable "ses_smtp_username" {
  description = "AWS SES SMTP username (IAM SMTP credentials access key)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "ses_smtp_password" {
  description = "AWS SES SMTP password (IAM SMTP credentials secret key)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "ses_from_email" {
  description = "Verified email address in AWS SES to send from (e.g., noreply@powerliftingcoach.app)"
  type        = string
  default     = "noreply@powerliftingcoach.app"
}

variable "google_oauth_client_id" {
  description = "Google OAuth 2.0 Client ID for social login"
  type        = string
  sensitive   = true
  default     = ""
}

variable "google_oauth_client_secret" {
  description = "Google OAuth 2.0 Client Secret for social login"
  type        = string
  sensitive   = true
  default     = ""
}
