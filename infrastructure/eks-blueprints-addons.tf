module "eks_blueprints_addons" {
  source  = "aws-ia/eks-blueprints-addons/aws"
  version = "~> 1.0"

  cluster_name      = aws_eks_cluster.main.name
  cluster_endpoint  = aws_eks_cluster.main.endpoint
  cluster_version   = aws_eks_cluster.main.version
  oidc_provider_arn = aws_iam_openid_connect_provider.eks.arn

  create_delay_dependencies = [aws_eks_node_group.main.arn]

  eks_addons = {
    aws-ebs-csi-driver = {
      most_recent              = true
      service_account_role_arn = module.ebs_csi_driver_irsa.iam_role_arn
    }
  }

  enable_karpenter = var.kubernetes_resources_enabled
  karpenter = {
    repository_username = data.aws_ecrpublic_authorization_token.token.user_name
    repository_password = data.aws_ecrpublic_authorization_token.token.password
    chart_version       = "1.1.1"
    namespace           = "karpenter"
  }
  karpenter_enable_spot_termination          = true
  karpenter_enable_instance_profile_creation = true
  karpenter_node = {
    iam_role_additional_policies = {
      AmazonSSMManagedInstanceCore = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
    }
  }

  enable_metrics_server = var.kubernetes_resources_enabled
  metrics_server = {
    values = [
      yamlencode({
        args = [
          "--kubelet-insecure-tls",
          "--kubelet-preferred-address-types=InternalIP"
        ]
        resources = {
          requests = {
            cpu    = "50m"
            memory = "128Mi"
          }
          limits = {
            cpu    = "100m"
            memory = "256Mi"
          }
        }
      })
    ]
  }

  enable_cert_manager = var.kubernetes_resources_enabled
  cert_manager = {
    chart_version = "v1.14.2"
    values = [
      yamlencode({
        installCRDs = true
        resources = {
          requests = {
            cpu    = "50m"
            memory = "128Mi"
          }
          limits = {
            cpu    = "100m"
            memory = "256Mi"
          }
        }
      })
    ]
  }
  cert_manager_route53_hosted_zone_arns = [aws_route53_zone.main.arn]

  enable_external_dns = var.kubernetes_resources_enabled
  external_dns = {
    values = [
      yamlencode({
        provider      = "aws"
        domainFilters = [var.domain_name]
        policy        = "sync"
        registry      = "txt"
        txtOwnerId    = local.cluster_name
        interval      = "1m"
        resources = {
          requests = {
            cpu    = "50m"
            memory = "64Mi"
          }
          limits = {
            cpu    = "100m"
            memory = "128Mi"
          }
        }
      })
    ]
  }
  external_dns_route53_zone_arns = [aws_route53_zone.main.arn]

  enable_aws_load_balancer_controller = var.kubernetes_resources_enabled

  enable_kube_prometheus_stack = var.kubernetes_resources_enabled && !var.stopped
  kube_prometheus_stack = {
    values = [
      yamlencode({
        prometheus = {
          prometheusSpec = {
            retention = "15d"
            resources = {
              requests = {
                cpu    = "200m"
                memory = "512Mi"
              }
              limits = {
                cpu    = "500m"
                memory = "2Gi"
              }
            }
            storageSpec = {
              volumeClaimTemplate = {
                spec = {
                  storageClassName = "gp3"
                  accessModes      = ["ReadWriteOnce"]
                  resources = {
                    requests = {
                      storage = "50Gi"
                    }
                  }
                }
              }
            }
          }
        }
        grafana = {
          enabled       = true
          adminPassword = var.kubernetes_resources_enabled ? random_password.grafana_admin_password[0].result : ""
          persistence = {
            enabled          = true
            storageClassName = "gp3"
            size             = "10Gi"
          }
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
        alertmanager = {
          enabled = false
        }
      })
    ]
  }

  tags = {
    Environment = var.environment
    Project     = var.project_name
  }

  depends_on = [aws_eks_cluster.main]
}

module "ebs_csi_driver_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name_prefix = "${local.cluster_name}-ebs-csi-"

  attach_ebs_csi_policy = true

  oidc_providers = {
    main = {
      provider_arn               = aws_iam_openid_connect_provider.eks.arn
      namespace_service_accounts = ["kube-system:ebs-csi-controller-sa"]
    }
  }

  tags = {
    Name        = "${local.cluster_name}-ebs-csi-driver-role"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_role_policy" "karpenter_node_s3" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  name_prefix = "${local.cluster_name}-karpenter-node-s3-"
  role        = module.eks_blueprints_addons.karpenter.node_iam_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "s3:*"
        Resource = "*"
      }
    ]
  })
}

data "aws_ecrpublic_authorization_token" "token" {
  provider = aws.virginia
}

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
      role      = module.eks_blueprints_addons.karpenter.node_iam_role_name
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
        /etc/eks/bootstrap.sh ${aws_eks_cluster.main.name}
      EOT
    }
  })

  depends_on = [module.eks_blueprints_addons]
}

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

  depends_on = [kubectl_manifest.karpenter_node_class]
}

resource "kubectl_manifest" "letsencrypt_prod" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = "letsencrypt-prod"
    }
    spec = {
      acme = {
        server = "https://acme-v02.api.letsencrypt.org/directory"
        email  = "admin@${var.domain_name}"
        privateKeySecretRef = {
          name = "letsencrypt-prod"
        }
        solvers = [
          {
            http01 = {
              ingress = {
                class = "nginx"
              }
            }
          }
        ]
      }
    }
  })

  depends_on = [module.eks_blueprints_addons]
}

resource "helm_release" "nginx_ingress" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  name             = "ingress-nginx"
  repository       = "https://kubernetes.github.io/ingress-nginx"
  chart            = "ingress-nginx"
  namespace        = "ingress-nginx"
  create_namespace = true
  version          = "4.10.0"

  values = [
    yamlencode({
      controller = {
        service = {
          type = "LoadBalancer"
          annotations = {
            "service.beta.kubernetes.io/aws-load-balancer-type"                              = "nlb"
            "service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled" = "true"
            "service.beta.kubernetes.io/aws-load-balancer-backend-protocol"                  = "tcp"
            "external-dns.alpha.kubernetes.io/hostname"                                      = "*.${var.domain_name}"
          }
        }
        ingressClassResource = {
          default = true
        }
        metrics = {
          enabled = true
        }
        resources = {
          requests = {
            cpu    = "100m"
            memory = "128Mi"
          }
          limits = {
            cpu    = "200m"
            memory = "256Mi"
          }
        }
      }
    })
  ]

  depends_on = [module.eks_blueprints_addons]
}
