terraform {
  required_providers {
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
  }
}

provider "kubernetes" {
  host  = var.kube_host
  token = var.kube_token
}

provider "helm" {
  kubernetes {
    host  = var.kube_host
    token = var.kube_token
  }
}

provider "kubectl" {
  host             = var.kube_host
  token            = var.kube_token
  load_config_file = false
}
