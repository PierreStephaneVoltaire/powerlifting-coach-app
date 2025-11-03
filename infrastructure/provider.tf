terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.25"
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

provider "digitalocean" {
  token = var.do_token
}

provider "digitalocean" {
  alias             = "spaces"
  spaces_access_id  = digitalocean_spaces_key.default.access_key
  spaces_secret_key = digitalocean_spaces_key.default.secret_key
}

provider "kubernetes" {
  host                   = digitalocean_kubernetes_cluster.k8s.endpoint
  cluster_ca_certificate = base64decode(digitalocean_kubernetes_cluster.k8s.kube_config[0].cluster_ca_certificate)
  token                  = digitalocean_kubernetes_cluster.k8s.kube_config[0].token
}

provider "helm" {
  kubernetes {
    host                   = digitalocean_kubernetes_cluster.k8s.endpoint
    cluster_ca_certificate = base64decode(digitalocean_kubernetes_cluster.k8s.kube_config[0].cluster_ca_certificate)
    token                  = digitalocean_kubernetes_cluster.k8s.kube_config[0].token
  }
}

# Generate kubeconfig file
resource "local_file" "kubeconfig" {
  content = yamlencode({
    apiVersion = "v1"
    kind       = "Config"
    clusters = [{
      name = digitalocean_kubernetes_cluster.k8s.name
      cluster = {
        certificate-authority-data = digitalocean_kubernetes_cluster.k8s.kube_config[0].cluster_ca_certificate
        server                     = digitalocean_kubernetes_cluster.k8s.endpoint
      }
    }]
    users = [{
      name = digitalocean_kubernetes_cluster.k8s.name
      user = {
        token = digitalocean_kubernetes_cluster.k8s.kube_config[0].token
      }
    }]
    contexts = [{
      name = digitalocean_kubernetes_cluster.k8s.name
      context = {
        cluster = digitalocean_kubernetes_cluster.k8s.name
        user    = digitalocean_kubernetes_cluster.k8s.name
      }
    }]
    current-context = digitalocean_kubernetes_cluster.k8s.name
  })
  filename = "${path.module}/kubeconfig.yaml"
}

output "kubeconfig_path" {
  value = local_file.kubeconfig.filename
}