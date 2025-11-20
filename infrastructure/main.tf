locals {
  cluster_name      = "${var.project_name}-${var.environment}"
  ses_smtp_endpoint = "email-smtp.${var.aws_region}.amazonaws.com"
}

resource "aws_s3_bucket" "videos" {
  bucket_prefix = "${local.cluster_name}-videos-"

  tags = {
    Name        = "${local.cluster_name}-videos"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_s3_bucket_versioning" "videos" {
  bucket = aws_s3_bucket.videos.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_cors_configuration" "videos" {
  bucket = aws_s3_bucket.videos.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["*"]
    expose_headers  = ["ETag", "Content-Length", "Content-Type"]
    max_age_seconds = 3600
  }
}

resource "aws_s3_bucket_public_access_block" "videos" {
  bucket = aws_s3_bucket.videos.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "videos" {
  bucket = aws_s3_bucket.videos.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.videos.arn}/*"
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.videos]
}

resource "aws_s3_bucket_lifecycle_configuration" "videos" {
  bucket = aws_s3_bucket.videos.id

  rule {
    id     = "expire-after-120-days"
    status = "Enabled"

    expiration {
      days = 120
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

resource "aws_iam_user" "s3_videos" {
  count = var.kubernetes_resources_enabled ? 1 : 0
  name  = "${local.cluster_name}-s3-videos"

  tags = {
    Name        = "${local.cluster_name}-s3-videos"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_access_key" "s3_videos" {
  count = var.kubernetes_resources_enabled ? 1 : 0
  user  = aws_iam_user.s3_videos[0].name
}

resource "aws_iam_user_policy" "s3_videos" {
  count = var.kubernetes_resources_enabled ? 1 : 0
  name  = "${local.cluster_name}-s3-videos-policy"
  user  = aws_iam_user.s3_videos[0].name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.videos.arn,
          "${aws_s3_bucket.videos.arn}/*"
        ]
      }
    ]
  })
}

module "kubernetes_base" {
  count  = var.kubernetes_resources_enabled ? 1 : 0
  source = "./modules/kubernetes-base"

  environment                = var.environment
  domain_name                = var.domain_name
  s3_access_key_id           = aws_iam_access_key.s3_videos[0].id
  s3_secret_access_key       = aws_iam_access_key.s3_videos[0].secret
  s3_bucket_domain           = aws_s3_bucket.videos.bucket_regional_domain_name
  s3_bucket_id               = aws_s3_bucket.videos.id
  aws_region                 = var.aws_region
  ses_smtp_endpoint          = local.ses_smtp_endpoint
  ses_smtp_username          = aws_iam_access_key.ses_smtp[0].id
  ses_smtp_password          = aws_iam_access_key.ses_smtp[0].ses_smtp_password_v4
  google_oauth_client_id     = var.google_oauth_client_id
  google_oauth_client_secret = var.google_oauth_client_secret
  stopped                    = var.stopped
  project_root               = path.module
}

module "kubernetes_networking" {
  count  = var.kubernetes_resources_enabled ? 1 : 0
  source = "./modules/kubernetes-networking"

  domain_name  = var.domain_name
  cluster_name = local.cluster_name
  stopped      = var.stopped
}

module "kubernetes_monitoring" {
  count  = var.kubernetes_resources_enabled ? 1 : 0
  source = "./modules/kubernetes-monitoring"

  domain_name            = var.domain_name
  app_namespace          = module.kubernetes_base[0].app_namespace
  grafana_admin_password = module.kubernetes_base[0].grafana_admin_password
  stopped                = var.stopped

  depends_on = [module.kubernetes_networking]
}

module "argocd" {
  count  = var.kubernetes_resources_enabled ? 1 : 0
  source = "./modules/argocd"

  domain_name = var.domain_name
  stopped     = var.stopped

  depends_on = [module.kubernetes_networking]
}

module "argocd_apps" {
  count  = var.kubernetes_resources_enabled && var.argocd_resources_enabled ? 1 : 0
  source = "./modules/argocd-apps"

  project_name     = var.project_name
  argocd_namespace = module.argocd[0].argocd_namespace
  app_namespace    = module.kubernetes_base[0].app_namespace
  deploy_frontend  = var.deploy_frontend
  deploy_datalayer = var.deploy_datalayer
  deploy_backend   = var.deploy_backend

  depends_on = [module.argocd]
}
