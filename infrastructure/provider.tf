terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.7"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "~> 1.14"
    }
    rancher2 = {
      source  = "rancher/rancher2"
      version = "~> 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = var.project_name
      ManagedBy   = "Terraform"
    }
  }
}

provider "aws" {
  alias  = "virginia"
  region = "us-east-1"

  default_tags {
    tags = {
      Environment = var.environment
      Project     = var.project_name
      ManagedBy   = "Terraform"
    }
  }
}

data "aws_partition" "current" {}

# Kubernetes providers configured to use k3s kubeconfig
# After first apply, download kubeconfig from S3:
# aws s3 cp s3://${bucket}/kubeconfig.yaml ${path.module}/kubeconfig.yaml
# Then set kubernetes_resources_enabled = true and apply again

provider "kubernetes" {
  config_path = var.kubernetes_resources_enabled ? "${path.module}/kubeconfig.yaml" : null
}

provider "helm" {
  kubernetes {
    config_path = var.kubernetes_resources_enabled ? "${path.module}/kubeconfig.yaml" : null
  }
}

provider "kubectl" {
  config_path    = var.kubernetes_resources_enabled ? "${path.module}/kubeconfig.yaml" : null
  load_config_file = var.kubernetes_resources_enabled
}

output "kubeconfig_path" {
  value = "${path.module}/kubeconfig.yaml"
}

output "kubeconfig_download_command" {
  description = "Command to download kubeconfig from S3 after instance is ready"
  value       = "aws s3 cp s3://${aws_s3_bucket.videos.id}/kubeconfig.yaml ${path.module}/kubeconfig.yaml"
}
