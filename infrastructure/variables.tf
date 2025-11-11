variable "azure_subscription_id" {
  description = "Azure subscription ID"
  type        = string
  sensitive   = true
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
  description = "Azure VM size for spot instance nodes. Using B-series for cost efficiency."
  type        = string
  default     = "Standard_B2ms"
}

variable "spot_node_min_count" {
  description = "Minimum number of spot instance nodes"
  type        = number
  default     = 1
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
  default     = "nolift-videos"
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

variable "stopped" {
  description = "When true, scales spot node pool to 0 and deletes LoadBalancer to save costs. Default node pool (nginx-ingress only) remains at 1 node."
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
  default     = "localhost"
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

