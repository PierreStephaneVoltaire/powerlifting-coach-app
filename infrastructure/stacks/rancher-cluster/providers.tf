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

provider "rancher2" {
  api_url   = "https://rancher.${var.domain_name}"
  bootstrap = true
  insecure  = true
}
