data "aws_caller_identity" "current" {}

# EKS Configuration - Commented out in favor of Rancher k3s setup
# Uncomment this section if you want to switch back to EKS

# resource "aws_iam_role" "eks_cluster" {
#   name = "${local.cluster_name}-cluster-role"

#   assume_role_policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [{
#       Action = "sts:AssumeRole"
#       Effect = "Allow"
#       Principal = {
#         Service = "eks.amazonaws.com"
#       }
#     }]
#   })

#   tags = {
#     Name        = "${local.cluster_name}-cluster-role"
#     Environment = var.environment
#     Project     = var.project_name
#   }
# }

# resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
#   policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
#   role       = aws_iam_role.eks_cluster.name
# }

# resource "aws_iam_role_policy_attachment" "eks_vpc_resource_controller" {
#   policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
#   role       = aws_iam_role.eks_cluster.name
# }

# resource "aws_security_group" "eks_cluster" {
#   name        = "${local.cluster_name}-cluster-sg"
#   description = "EKS cluster security group"
#   vpc_id      = aws_vpc.main.id

#   tags = {
#     Name        = "${local.cluster_name}-cluster-sg"
#     Environment = var.environment
#     Project     = var.project_name
#   }
# }

# resource "aws_security_group_rule" "cluster_ingress_https" {
#   description       = "Allow workstation to communicate with the cluster API Server"
#   type              = "ingress"
#   from_port         = 443
#   to_port           = 443
#   protocol          = "tcp"
#   cidr_blocks       = ["0.0.0.0/0"]
#   security_group_id = aws_security_group.eks_cluster.id
# }

# resource "aws_security_group_rule" "cluster_egress_all" {
#   description       = "Allow cluster egress"
#   type              = "egress"
#   from_port         = 0
#   to_port           = 0
#   protocol          = "-1"
#   cidr_blocks       = ["0.0.0.0/0"]
#   security_group_id = aws_security_group.eks_cluster.id
# }

# resource "aws_eks_cluster" "main" {
#   name     = local.cluster_name
#   version  = var.kubernetes_version
#   role_arn = aws_iam_role.eks_cluster.arn

#   vpc_config {
#     subnet_ids              = aws_subnet.public[*].id
#     endpoint_public_access  = true
#     endpoint_private_access = false
#     security_group_ids      = [aws_security_group.eks_cluster.id]
#   }

#   access_config {
#     authentication_mode = "API_AND_CONFIG_MAP"
#   }

#   upgrade_policy {
#     support_type = "STANDARD"
#   }

#   zonal_shift_config {
#     enabled = true
#   }

#   tags = {
#     Name        = local.cluster_name
#     Environment = var.environment
#     Project     = var.project_name
#   }

#   depends_on = [
#     aws_iam_role_policy_attachment.eks_cluster_policy,
#     aws_iam_role_policy_attachment.eks_vpc_resource_controller,
#   ]
# }

# data "tls_certificate" "eks" {
#   url = aws_eks_cluster.main.identity[0].oidc[0].issuer
# }

# resource "aws_iam_openid_connect_provider" "eks" {
#   client_id_list  = ["sts.amazonaws.com"]
#   thumbprint_list = [data.tls_certificate.eks.certificates[0].sha1_fingerprint]
#   url             = aws_eks_cluster.main.identity[0].oidc[0].issuer

#   tags = {
#     Name        = "${local.cluster_name}-oidc"
#     Environment = var.environment
#     Project     = var.project_name
#   }
# }

# resource "aws_eks_addon" "vpc_cni" {
#   cluster_name             = aws_eks_cluster.main.name
#   addon_name               = "vpc-cni"
#   addon_version            = "v1.20.3-eksbuild.1"
#   resolve_conflicts_on_create = "OVERWRITE"
#   resolve_conflicts_on_update = "OVERWRITE"

#   configuration_values = jsonencode({
#     env = {
#       AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG = "true"
#       ENI_CONFIG_LABEL_DEF               = "topology.kubernetes.io/zone"
#       WARM_IP_TARGET                     = "5"
#       MINIMUM_IP_TARGET                  = "2"
#       ENABLE_PREFIX_DELEGATION           = "true"
#       WARM_PREFIX_TARGET                 = "1"
#     }
#   })

#   tags = {
#     Name        = "${local.cluster_name}-vpc-cni"
#     Environment = var.environment
#     Project     = var.project_name
#   }
# }

# resource "aws_eks_addon" "coredns" {
#   cluster_name             = aws_eks_cluster.main.name
#   addon_name               = "coredns"
#   resolve_conflicts_on_create = "OVERWRITE"
#   resolve_conflicts_on_update = "OVERWRITE"

#   tags = {
#     Name        = "${local.cluster_name}-coredns"
#     Environment = var.environment
#     Project     = var.project_name
#   }

#   depends_on = [aws_eks_node_group.main]
# }

# resource "aws_eks_addon" "kube_proxy" {
#   cluster_name             = aws_eks_cluster.main.name
#   addon_name               = "kube-proxy"
#   resolve_conflicts_on_create = "OVERWRITE"
#   resolve_conflicts_on_update = "OVERWRITE"

#   tags = {
#     Name        = "${local.cluster_name}-kube-proxy"
#     Environment = var.environment
#     Project     = var.project_name
#   }
# }

# resource "aws_eks_addon" "eks_node_monitoring_agent" {
#   cluster_name             = aws_eks_cluster.main.name
#   addon_name               = "eks-node-monitoring-agent"
#   resolve_conflicts_on_create = "OVERWRITE"
#   resolve_conflicts_on_update = "OVERWRITE"

#   tags = {
#     Name        = "${local.cluster_name}-monitoring-agent"
#     Environment = var.environment
#     Project     = var.project_name
#   }

#   depends_on = [aws_eks_node_group.main]
# }

# resource "aws_iam_role" "eks_node" {
#   name = "${local.cluster_name}-node-role"

#   assume_role_policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [{
#       Action = "sts:AssumeRole"
#       Effect = "Allow"
#       Principal = {
#         Service = "ec2.amazonaws.com"
#       }
#     }]
#   })

#   tags = {
#     Name        = "${local.cluster_name}-node-role"
#     Environment = var.environment
#     Project     = var.project_name
#   }
# }

# resource "aws_iam_role_policy_attachment" "eks_worker_node_policy" {
#   policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
#   role       = aws_iam_role.eks_node.name
# }

# resource "aws_iam_role_policy_attachment" "eks_cni_policy" {
#   policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
#   role       = aws_iam_role.eks_node.name
# }

# resource "aws_iam_role_policy_attachment" "eks_container_registry" {
#   policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
#   role       = aws_iam_role.eks_node.name
# }

# resource "aws_iam_role_policy_attachment" "eks_ebs_csi_driver" {
#   policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
#   role       = aws_iam_role.eks_node.name
# }

# resource "aws_security_group" "eks_node" {
#   name        = "${local.cluster_name}-node-sg"
#   description = "EKS node security group"
#   vpc_id      = aws_vpc.main.id

#   tags = {
#     Name                                          = "${local.cluster_name}-node-sg"
#     Environment                                   = var.environment
#     Project                                       = var.project_name
#     "kubernetes.io/cluster/${local.cluster_name}" = "owned"
#   }
# }

# resource "aws_security_group_rule" "node_ingress_self" {
#   description       = "Node to node all ports/protocols"
#   type              = "ingress"
#   from_port         = 0
#   to_port           = 0
#   protocol          = "-1"
#   self              = true
#   security_group_id = aws_security_group.eks_node.id
# }

# resource "aws_security_group_rule" "node_ingress_cluster" {
#   description              = "Allow cluster to communicate with nodes"
#   type                     = "ingress"
#   from_port                = 0
#   to_port                  = 65535
#   protocol                 = "tcp"
#   source_security_group_id = aws_security_group.eks_cluster.id
#   security_group_id        = aws_security_group.eks_node.id
# }

# resource "aws_security_group_rule" "node_egress_all" {
#   description       = "Node all egress"
#   type              = "egress"
#   from_port         = 0
#   to_port           = 0
#   protocol          = "-1"
#   cidr_blocks       = ["0.0.0.0/0"]
#   security_group_id = aws_security_group.eks_node.id
# }

# resource "aws_security_group_rule" "cluster_ingress_node" {
#   description              = "Allow nodes to communicate with cluster"
#   type                     = "ingress"
#   from_port                = 443
#   to_port                  = 443
#   protocol                 = "tcp"
#   source_security_group_id = aws_security_group.eks_node.id
#   security_group_id        = aws_security_group.eks_cluster.id
# }

# resource "aws_eks_node_group" "main" {
#   cluster_name    = aws_eks_cluster.main.name
#   node_group_name = "${local.cluster_name}-default"
#   node_role_arn   = aws_iam_role.eks_node.arn
#   subnet_ids      = aws_subnet.public[*].id
#   capacity_type   = "SPOT"
#   instance_types  = ["t3a.medium", "t3.medium", "t2.medium"]

#   scaling_config {
#     min_size     = var.stopped ? 0 : var.worker_min_size
#     max_size     = var.worker_max_size
#     desired_size = var.stopped ? 0 : var.worker_desired_capacity
#   }

#   update_config {
#     max_unavailable = 1
#   }

#   tags = {
#     Name        = "${local.cluster_name}-default"
#     NodeSize    = "medium"
#     Environment = var.environment
#     Project     = var.project_name
#   }

#   depends_on = [
#     aws_iam_role_policy_attachment.eks_worker_node_policy,
#     aws_iam_role_policy_attachment.eks_cni_policy,
#     aws_iam_role_policy_attachment.eks_container_registry,
#     aws_iam_role_policy_attachment.eks_ebs_csi_driver,
#     aws_eks_addon.vpc_cni,
#     aws_eks_addon.kube_proxy,
#   ]
# }

# resource "aws_iam_role_policy" "eks_node_s3" {
#   name_prefix = "${local.cluster_name}-eks-node-s3-"
#   role        = aws_iam_role.eks_node.name

#   policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [
#       {
#         Effect = "Allow"
#         Action = [
#           "s3:PutObject",
#           "s3:GetObject",
#           "s3:DeleteObject",
#           "s3:ListBucket"
#         ]
#         Resource = [
#           aws_s3_bucket.videos.arn,
#           "${aws_s3_bucket.videos.arn}/*"
#         ]
#       }
#     ]
#   })
# }
