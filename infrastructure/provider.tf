terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.52"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.7"
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

provider "azurerm" {
  features {}
  subscription_id = var.azure_subscription_id
}

provider "kubernetes" {
  host                   = azurerm_kubernetes_cluster.k8s.kube_config[0].host
  cluster_ca_certificate = base64decode(azurerm_kubernetes_cluster.k8s.kube_config[0].cluster_ca_certificate)
  client_certificate     = base64decode(azurerm_kubernetes_cluster.k8s.kube_config[0].client_certificate)
  client_key             = base64decode(azurerm_kubernetes_cluster.k8s.kube_config[0].client_key)
}

provider "helm" {
  kubernetes {
    host                   = azurerm_kubernetes_cluster.k8s.kube_config[0].host
    cluster_ca_certificate = base64decode(azurerm_kubernetes_cluster.k8s.kube_config[0].cluster_ca_certificate)
    client_certificate     = base64decode(azurerm_kubernetes_cluster.k8s.kube_config[0].client_certificate)
    client_key             = base64decode(azurerm_kubernetes_cluster.k8s.kube_config[0].client_key)
  }
}

# Generate kubeconfig file
resource "local_file" "kubeconfig" {
  content  = azurerm_kubernetes_cluster.k8s.kube_config_raw
  filename = "${path.module}/kubeconfig.yaml"
}

output "kubeconfig_path" {
  value = local_file.kubeconfig.filename
}