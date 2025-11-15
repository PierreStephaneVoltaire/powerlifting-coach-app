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

data "aws_eks_cluster_auth" "main" {
  count = var.kubernetes_resources_enabled ? 1 : 0
  name  = module.eks.cluster_name
}

provider "kubernetes" {
  host                   = var.kubernetes_resources_enabled ? module.eks.cluster_endpoint : null
  cluster_ca_certificate = var.kubernetes_resources_enabled ? base64decode(module.eks.cluster_certificate_authority_data) : null
  token                  = var.kubernetes_resources_enabled ? data.aws_eks_cluster_auth.main[0].token : null
}

provider "helm" {
  kubernetes {
    host                   = var.kubernetes_resources_enabled ? module.eks.cluster_endpoint : null
    cluster_ca_certificate = var.kubernetes_resources_enabled ? base64decode(module.eks.cluster_certificate_authority_data) : null
    token                  = var.kubernetes_resources_enabled ? data.aws_eks_cluster_auth.main[0].token : null
  }
}

provider "kubectl" {
  host                   = var.kubernetes_resources_enabled ? module.eks.cluster_endpoint : null
  cluster_ca_certificate = var.kubernetes_resources_enabled ? base64decode(module.eks.cluster_certificate_authority_data) : null
  token                  = var.kubernetes_resources_enabled ? data.aws_eks_cluster_auth.main[0].token : null
  load_config_file       = false
}

resource "local_file" "kubeconfig" {
  count    = var.kubernetes_resources_enabled ? 1 : 0
  filename = "${path.module}/kubeconfig.yaml"
  content = templatefile("${path.module}/kubeconfig-template.yaml", {
    cluster_name = module.eks.cluster_name
    endpoint     = module.eks.cluster_endpoint
    ca_data      = module.eks.cluster_certificate_authority_data
    region       = var.aws_region
  })
}

output "kubeconfig_path" {
  value = var.kubernetes_resources_enabled ? local_file.kubeconfig[0].filename : null
}
