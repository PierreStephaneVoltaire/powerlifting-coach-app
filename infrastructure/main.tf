locals {
  cluster_name = "${var.project_name}-${var.environment}"
}

# VPC
resource "digitalocean_vpc" "this" {
  name   = "${local.cluster_name}-${var.environment}-vpc"
  region = var.region
}

# Spaces key for bucket access
resource "digitalocean_spaces_key" "default" {
  name = "${local.cluster_name}-spaces-key"
  grant {
    bucket     = ""
    permission = "fullaccess"
  }
}

# Spaces bucket for video storage
resource "digitalocean_spaces_bucket" "videos" {
  provider = digitalocean.spaces
  name     = var.spaces_bucket_name
  region   = var.region
  acl      = "public-read"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["*"]
    max_age_seconds = 3600
  }

  lifecycle_rule {
    id      = "expire-after-120-days"
    prefix  = ""
    enabled = true

    expiration {
      days = 120
    }
  }

  force_destroy = false
}

resource "digitalocean_spaces_bucket_policy" "public_get" {
  provider = digitalocean.spaces
  bucket   = digitalocean_spaces_bucket.videos.name
  region   = var.region
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowPublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = ["s3:GetObject"]
        Resource = [
          "arn:aws:s3:::${digitalocean_spaces_bucket.videos.name}/*"
        ]
      }
    ]
  })
}

# Kubernetes cluster
resource "digitalocean_kubernetes_cluster" "k8s" {
  name    = local.cluster_name
  region  = var.region
  version = var.kubernetes_version

  vpc_uuid = digitalocean_vpc.this.id

  node_pool {
    name       = "default-pool"
    size       = var.node_size
    auto_scale = true
    min_nodes  = 0
    max_nodes  = 3
  }
}
