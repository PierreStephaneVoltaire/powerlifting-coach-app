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

variable "spot_node_size" {
  description = "Azure VM size for spot instance nodes"
  type        = string
  default     = "Standard_B2s"
}

variable "spot_node_min_count" {
  description = "Minimum number of spot instance nodes"
  type        = number
  default     = 0
}

variable "spot_node_max_count" {
  description = "Maximum number of spot instance nodes"
  type        = number
  default     = 5
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

variable "azure_email_smtp_host" {
  description = "Azure Communication Services Email SMTP host (smtp.azurecomm.net)"
  type        = string
  default     = "smtp.azurecomm.net"
}

variable "azure_email_smtp_username" {
  description = "Azure Communication Services Email SMTP username (your verified email address)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "azure_email_smtp_password" {
  description = "Azure Communication Services Email SMTP password (connection string or access key)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "azure_email_from_email" {
  description = "Verified email address in Azure Communication Services to send from (e.g., noreply@coachpotato.app)"
  type        = string
  default     = "noreply@coachpotato.app"
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

# DNS Configuration Variables

variable "azure_email_domain_verification_code" {
  description = "Azure Communication Services domain verification TXT record value"
  type        = string
  sensitive   = true
  default     = ""
}

variable "enable_mx_records" {
  description = "Enable MX records for receiving email (only needed if you want to receive emails)"
  type        = bool
  default     = false
}

variable "azure_email_mx_endpoint" {
  description = "Azure Communication Services MX endpoint (only needed if enable_mx_records is true)"
  type        = string
  default     = ""
}

variable "azure_email_dkim_selector1" {
  description = "DKIM selector 1 from Azure Communication Services (e.g., selector1._domainkey)"
  type        = string
  default     = ""
}

variable "azure_email_dkim_value1" {
  description = "DKIM value 1 from Azure Communication Services"
  type        = string
  default     = ""
}

variable "azure_email_dkim_selector2" {
  description = "DKIM selector 2 from Azure Communication Services (e.g., selector2._domainkey)"
  type        = string
  default     = ""
}

variable "azure_email_dkim_value2" {
  description = "DKIM value 2 from Azure Communication Services"
  type        = string
  default     = ""
}
