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

data "aws_ssm_parameter" "k3s_ca_cert" {
  count      = var.kubernetes_resources_enabled ? 1 : 0
  name       = "/${local.cluster_name}/k3s/ca-cert"
  depends_on = [aws_autoscaling_group.control_plane]
}

data "aws_ssm_parameter" "k3s_client_cert" {
  count      = var.kubernetes_resources_enabled ? 1 : 0
  name       = "/${local.cluster_name}/k3s/client-cert"
  depends_on = [aws_autoscaling_group.control_plane]
}

data "aws_ssm_parameter" "k3s_client_key" {
  count      = var.kubernetes_resources_enabled ? 1 : 0
  name       = "/${local.cluster_name}/k3s/client-key"
  depends_on = [aws_autoscaling_group.control_plane]
}

provider "kubernetes" {
  host                   = var.kubernetes_resources_enabled ? "https://${aws_lb.control_plane.dns_name}:6443" : null
  cluster_ca_certificate = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_ca_cert[0].value) : null
  client_certificate     = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_client_cert[0].value) : null
  client_key             = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_client_key[0].value) : null
}

provider "helm" {
  kubernetes {
    host                   = var.kubernetes_resources_enabled ? "https://${aws_lb.control_plane.dns_name}:6443" : null
    cluster_ca_certificate = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_ca_cert[0].value) : null
    client_certificate     = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_client_cert[0].value) : null
    client_key             = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_client_key[0].value) : null
  }
}

provider "kubectl" {
  host                   = var.kubernetes_resources_enabled ? "https://${aws_lb.control_plane.dns_name}:6443" : null
  cluster_ca_certificate = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_ca_cert[0].value) : null
  client_certificate     = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_client_cert[0].value) : null
  client_key             = var.kubernetes_resources_enabled ? base64decode(data.aws_ssm_parameter.k3s_client_key[0].value) : null
  load_config_file       = false
}

resource "local_file" "kubeconfig" {
  count    = var.kubernetes_resources_enabled ? 1 : 0
  filename = "${path.module}/kubeconfig.yaml"
  content = yamlencode({
    apiVersion = "v1"
    kind       = "Config"
    clusters = [{
      name = local.cluster_name
      cluster = {
        certificate-authority-data = data.aws_ssm_parameter.k3s_ca_cert[0].value
        server                     = "https://${aws_lb.control_plane.dns_name}:6443"
      }
    }]
    users = [{
      name = "admin"
      user = {
        client-certificate-data = data.aws_ssm_parameter.k3s_client_cert[0].value
        client-key-data         = data.aws_ssm_parameter.k3s_client_key[0].value
      }
    }]
    contexts = [{
      name = local.cluster_name
      context = {
        cluster = local.cluster_name
        user    = "admin"
      }
    }]
    current-context = local.cluster_name
  })
}

output "kubeconfig_path" {
  value = var.kubernetes_resources_enabled ? local_file.kubeconfig[0].filename : null
}
