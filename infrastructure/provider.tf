terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
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
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
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

# Kubernetes provider configuration
# This requires the kubeconfig to be available from S3
# We'll use config_path to read from the local kubeconfig file
provider "kubernetes" {
  config_path = var.kubernetes_resources_enabled ? local_file.kubeconfig[0].filename : null
}

provider "helm" {
  kubernetes {
    config_path = var.kubernetes_resources_enabled ? local_file.kubeconfig[0].filename : null
  }
}

provider "kubectl" {
  config_path = var.kubernetes_resources_enabled ? local_file.kubeconfig[0].filename : null
}

# Download and save kubeconfig from S3
resource "local_file" "kubeconfig" {
  count    = var.kubernetes_resources_enabled ? 1 : 0
  content  = data.aws_s3_object.kubeconfig[0].body
  filename = "${path.module}/kubeconfig.yaml"

  depends_on = [
    data.aws_s3_object.kubeconfig
  ]
}

output "kubeconfig_path" {
  value       = var.kubernetes_resources_enabled ? local_file.kubeconfig[0].filename : "Run terraform apply with kubernetes_resources_enabled=true to generate kubeconfig"
  description = "Path to the kubeconfig file for accessing the k3s cluster"
}