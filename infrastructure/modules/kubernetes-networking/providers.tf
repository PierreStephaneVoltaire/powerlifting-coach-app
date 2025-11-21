terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.7"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "~> 1.14"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
  }
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
