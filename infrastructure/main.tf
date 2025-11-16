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
