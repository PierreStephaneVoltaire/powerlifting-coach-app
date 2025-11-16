module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.8.0"
name  = local.cluster_name
  kubernetes_version = var.kubernetes_version

  vpc_id     = aws_vpc.main.id
  subnet_ids = aws_subnet.public[*].id

  endpoint_public_access  = true
  endpoint_private_access = false
  authentication_mode   = "API_AND_CONFIG_MAP"
  upgrade_policy = {
    support_type = "STANDARD"
  }

  zonal_shift_config = {
    enabled = true
  }
  dataplane_wait_duration = "240s"
  addons = {
    vpc-cni = {
      addon_version  = "v1.20.3-eksbuild.1"
      most_recent       = false
      before_compute    = true
      resolve_conflicts_on_create = "OVERWRITE"
      resolve_conflicts_on_update = "OVERWRITE"
      configuration_values = jsonencode({env={
        AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG = "true"
        ENI_CONFIG_LABEL_DEF               = "topology.kubernetes.io/zone"
        WARM_IP_TARGET ="5"
        MINIMUM_IP_TARGET="2"
        ENABLE_PREFIX_DELEGATION = "true"
        WARM_PREFIX_TARGET       = "1"
      }}
      )
    }
    coredns = {
      most_recent = true
      before_compute    = false
      resolve_conflicts_on_create = "OVERWRITE"
      resolve_conflicts_on_update = "OVERWRITE"
    }
    eks-node-monitoring-agent = {
      most_recent = true
            resolve_conflicts_on_create = "OVERWRITE"
      resolve_conflicts_on_update = "OVERWRITE"
    }
    kube-proxy = {
      most_recent = true
            resolve_conflicts_on_create = "OVERWRITE"
      resolve_conflicts_on_update = "OVERWRITE"
    }
}

  eks_managed_node_groups = {
    medium = {
      name            = "${local.cluster_name}-default"
      use_name_prefix = false

      instance_types = ["t3a.medium", "t3.medium", "t2.medium"]
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
        NodeSize = "medium"
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

  security_group_additional_rules = {
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
  role        = module.eks.eks_managed_node_groups["medium"].iam_role_name

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
