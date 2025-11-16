# AWS Configuration
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ca-central-1"
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "nolift"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

# EKS Configuration
variable "kubernetes_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.34"
}

# EKS Node Configuration (all spot instances, smallest possible)

variable "worker_desired_capacity" {
  description = "Desired number of worker nodes"
  type        = number
  default     = 2
}

variable "worker_min_size" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 1
}

variable "worker_max_size" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 5
}


# Kubernetes Configuration
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

variable "stopped" {
  description = "When true, scales worker nodes to 0 to save costs. Control plane remains running."
  type        = bool
  default     = false
}

variable "ai_features_enabled" {
  description = "Enable AI features including LiteLLM deployment, chat interface, and AI coaching. Set to false to reduce costs."
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
  description = "Domain name for the application (e.g., nolift.training)"
  type        = string
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

variable "email_domain_verified" {
  description = "Whether the email domain has been verified in AWS SES. Set to false initially, then true after domain verification completes."
  type        = bool
  default     = false
}

variable "monthly_budget_limit" {
  description = "Monthly budget limit in USD for AWS cost monitoring"
  type        = number
  default     = 150
}

variable "budget_notification_email" {
  description = "Email address to receive budget notifications when spend exceeds thresholds"
  type        = string
  default = "psvoltaire96@gmai.com"
}

# ArgoCD Deployment Toggles
variable "deploy_frontend" {
  description = "Deploy frontend application via ArgoCD"
  type        = bool
  default     = true
}

variable "deploy_datalayer" {
  description = "Deploy datalayer (postgres, valkey, rabbitmq, keycloak) via ArgoCD"
  type        = bool
  default     = true
}

variable "deploy_backend" {
  description = "Deploy backend services via ArgoCD"
  type        = bool
  default     = true
}

variable "rancher_cluster_enabled" {
  description = "Enable Rancher cluster creation. Set to false for initial Rancher Server deployment, then true after Rancher Server is running."
  type        = bool
  default     = false
}

