variable "cluster_name" {
  description = "Name of the cluster"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "project_name" {
  description = "Project name"
  type        = string
}

variable "domain_name" {
  description = "Domain name"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "aws_account_id" {
  description = "AWS account ID"
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version (K3s format)"
  type        = string
}

variable "worker_desired_capacity" {
  description = "Desired number of worker nodes"
  type        = number
}

variable "stopped" {
  description = "Whether cluster is stopped"
  type        = bool
  default     = false
}

variable "ami_id" {
  description = "AMI ID for cluster nodes"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID"
  type        = string
}

variable "rancher_server_sg_id" {
  description = "Rancher server security group ID"
  type        = string
}

variable "rancher_server_fqdn" {
  description = "Rancher server FQDN"
  type        = string
}

variable "route53_zone_id" {
  description = "Route53 zone ID"
  type        = string
}

variable "admin_ips" {
  description = "Admin IPs for SSH access"
  type        = list(string)
}

variable "rancher_admin_password" {
  description = "Rancher admin password"
  type        = string
  sensitive   = true
}

variable "project_root" {
  description = "Project root path"
  type        = string
}
