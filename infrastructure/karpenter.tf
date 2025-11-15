# Karpenter Configuration using official terraform-aws-modules v21.8.0

# Karpenter Module - handles IAM roles, SQS, EventBridge, etc.
module "karpenter" {
  source  = "terraform-aws-modules/eks/aws//modules/karpenter"
  version = "~> 21.8.0"

  cluster_name = module.eks.cluster_name

  enable_irsa            = true
  irsa_oidc_provider_arn = module.eks.oidc_provider_arn

  # Additional policies for Karpenter nodes
  node_iam_role_additional_policies = {
    AmazonSSMManagedInstanceCore = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  }

  tags = {
    Name        = "${local.cluster_name}-karpenter"
    Environment = var.environment
    Project     = var.project_name
  }
}

# S3 access policy for Karpenter nodes (attached separately to avoid circular dependency)
resource "aws_iam_role_policy" "karpenter_node_s3" {
  name_prefix = "${local.cluster_name}-karpenter-node-s3-"
  role        = module.karpenter.node_iam_role_name

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

# Karpenter Helm Release
resource "helm_release" "karpenter" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  name             = "karpenter"
  repository       = "oci://public.ecr.aws/karpenter"
  chart            = "karpenter"
  namespace        = "karpenter"
  create_namespace = true
  version          = "1.1.1"

  values = [
    yamlencode({
      settings = {
        clusterName       = module.eks.cluster_name
        clusterEndpoint   = module.eks.cluster_endpoint
        interruptionQueue = module.karpenter.queue_name
      }
      serviceAccount = {
        annotations = {
          "eks.amazonaws.com/role-arn" = module.karpenter.iam_role_arn
        }
      }
      controller = {
        resources = {
          requests = {
            cpu    = "100m"
            memory = "256Mi"
          }
          limits = {
            cpu    = "200m"
            memory = "512Mi"
          }
        }
      }
    })
  ]

  depends_on = [
    module.eks,
    module.karpenter
  ]
}

# Karpenter EC2NodeClass - Spot instances only
resource "kubectl_manifest" "karpenter_node_class" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "karpenter.k8s.aws/v1"
    kind       = "EC2NodeClass"
    metadata = {
      name = "default"
    }
    spec = {
      amiFamily = "AL2"
      role      = module.karpenter.node_iam_role_name
      subnetSelectorTerms = [
        {
          tags = {
            "kubernetes.io/cluster/${local.cluster_name}" = "owned"
          }
        }
      ]
      securityGroupSelectorTerms = [
        {
          tags = {
            "kubernetes.io/cluster/${local.cluster_name}" = "owned"
          }
        }
      ]
      tags = {
        Name                                           = "${local.cluster_name}-karpenter-node"
        Environment                                    = var.environment
        Project                                        = var.project_name
        "kubernetes.io/cluster/${local.cluster_name}" = "owned"
        "karpenter.sh/discovery"                       = local.cluster_name
      }
      blockDeviceMappings = [
        {
          deviceName = "/dev/xvda"
          ebs = {
            volumeSize          = "20Gi"
            volumeType          = "gp3"
            encrypted           = true
            deleteOnTermination = true
          }
        }
      ]
      userData = <<-EOT
        #!/bin/bash
        /etc/eks/bootstrap.sh ${module.eks.cluster_name}
      EOT
    }
  })

  depends_on = [helm_release.karpenter]
}

# Karpenter NodePool - Medium spot instances only (no on-demand fallback)
resource "kubectl_manifest" "karpenter_node_pool" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "karpenter.sh/v1"
    kind       = "NodePool"
    metadata = {
      name = "medium-spot"
    }
    spec = {
      template = {
        spec = {
          nodeClassRef = {
            group = "karpenter.k8s.aws"
            kind  = "EC2NodeClass"
            name  = "default"
          }
          requirements = [
            {
              key      = "kubernetes.io/arch"
              operator = "In"
              values   = ["amd64"]
            },
            {
              key      = "kubernetes.io/os"
              operator = "In"
              values   = ["linux"]
            },
            {
              key      = "karpenter.sh/capacity-type"
              operator = "In"
              values   = ["spot"]
            },
            {
              key      = "karpenter.k8s.aws/instance-category"
              operator = "In"
              values   = ["t"]
            },
            {
              key      = "karpenter.k8s.aws/instance-generation"
              operator = "Gt"
              values   = ["2"]
            },
            {
              key      = "node.kubernetes.io/instance-type"
              operator = "In"
              values = [
                "t3.medium",
                "t3a.medium",
                "t2.medium"
              ]
            }
          ]
          taints = []
        }
      }
      limits = {
        cpu    = "100"
        memory = "100Gi"
      }
      disruption = {
        consolidationPolicy = "WhenEmptyOrUnderutilized"
        consolidateAfter    = "1m"
      }
    }
  })

  depends_on = [
    kubectl_manifest.karpenter_node_class
  ]
}
