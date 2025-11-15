module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.8.0"

  cluster_name    = local.cluster_name
  cluster_version = var.kubernetes_version

  vpc_id     = aws_vpc.main.id
  subnet_ids = aws_subnet.public[*].id

  cluster_endpoint_public_access  = true
  cluster_endpoint_private_access = false

  enable_irsa = true

  eks_managed_node_groups = {
    small = {
      name            = "${local.cluster_name}-spot-small"
      use_name_prefix = false

      instance_types = ["t3a.small", "t3.small", "t2.small"]
      capacity_type  = "SPOT"

      min_size     = var.stopped ? 0 : var.worker_min_size
      max_size     = var.worker_max_size
      desired_size = var.stopped ? 0 : var.worker_desired_capacity

      update_config = {
        max_unavailable = 1
      }

      iam_role_additional_policies = {
        AmazonEBSCSIDriverPolicy = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
      }

      tags = {
        NodeSize = "small"
      }
    }
  }

  node_security_group_additional_rules = {
    ingress_self_all = {
      description = "Node to node all ports/protocols"
      protocol    = "-1"
      from_port   = 0
      to_port     = 0
      type        = "ingress"
      self        = true
    }
    egress_all = {
      description      = "Node all egress"
      protocol         = "-1"
      from_port        = 0
      to_port          = 0
      type             = "egress"
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = ["::/0"]
    }
  }

  cluster_security_group_additional_rules = {
    ingress_workstation_https = {
      description = "Allow workstation to communicate with the cluster API Server"
      protocol    = "tcp"
      from_port   = 443
      to_port     = 443
      type        = "ingress"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  tags = {
    Name        = local.cluster_name
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_role_policy" "eks_node_s3" {
  name_prefix = "${local.cluster_name}-eks-node-s3-"
  role        = module.eks.eks_managed_node_groups["small"].iam_role_name

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
