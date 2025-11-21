terraform {
  required_providers {
    rancher2 = {
      source  = "rancher/rancher2"
      version = "~> 8.3.1"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }
  }
}
