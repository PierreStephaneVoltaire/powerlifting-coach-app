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
  description = "Domain name for the application (e.g., powerliftingcoach.app)"
  type        = string
  default     = "localhost"
}
