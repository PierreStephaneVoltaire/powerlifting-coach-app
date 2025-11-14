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

# Control Plane Configuration
variable "control_plane_instance_type" {
  description = "EC2 instance type for control plane nodes (will use spot instances)"
  type        = string
  default     = "t3.small"
}

variable "control_plane_volume_size" {
  description = "EBS volume size in GB for control plane nodes"
  type        = number
  default     = 30
}

# Worker Configuration
variable "worker_instance_type" {
  description = "EC2 instance type for worker nodes (will use spot instances)"
  type        = string
  default     = "t3.medium"
}

variable "worker_volume_size" {
  description = "EBS volume size in GB for worker nodes"
  type        = number
  default     = 30
}

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

variable "max_pods_per_node" {
  description = "Maximum number of pods per node"
  type        = number
  default     = 110
}

# Network Configuration
variable "pod_network_cidr" {
  description = "CIDR block for pod network"
  type        = string
  default     = "10.42.0.0/16"
}

variable "nginx_lb_ssh_allowed_cidr" {
  description = "CIDR block allowed to SSH into nginx load balancer (e.g., your IP: 1.2.3.4/32)"
  type        = string
  default     = "0.0.0.0/0"
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

