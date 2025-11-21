locals {
  cluster_name = "${var.project_name}-${var.environment}"
}

resource "rancher2_bootstrap" "admin" {
  initial_password = data.terraform_remote_state.base.outputs.rancher_admin_password
  password         = data.terraform_remote_state.base.outputs.rancher_admin_password
}

resource "rancher2_setting" "agenttlsmode" {
  depends_on = [rancher2_bootstrap.admin]

  name  = "agent-tls-mode"
  value = "system-store"
}

resource "rancher2_cloud_credential" "aws" {
  name = "${local.cluster_name}-aws-creds"

  amazonec2_credential_config {
    access_key = data.aws_caller_identity.current.account_id
    secret_key = ""
  }

  depends_on = [rancher2_bootstrap.admin]
}

resource "aws_ssm_parameter" "password" {
  name  = "/rancher_admin"
  type  = "String"
  value = data.terraform_remote_state.base.outputs.rancher_admin_password
}

resource "rancher2_machine_config_v2" "cluster_nodes" {
  generate_name = "${local.cluster_name}-node"

  amazonec2_config {
    ami                   = data.terraform_remote_state.base.outputs.ami_id
    region                = var.aws_region
    security_group        = [aws_security_group.rancher_node.name]
    subnet_id             = data.terraform_remote_state.base.outputs.public_subnet_ids[0]
    vpc_id                = data.terraform_remote_state.base.outputs.vpc_id
    zone                  = "a"
    instance_type         = "t4g.large"
    root_size             = "30"
    iam_instance_profile  = aws_iam_instance_profile.rancher_node.name
    ssh_user              = "ec2-user"
    request_spot_instance = false
  }

  depends_on = [rancher2_bootstrap.admin, aws_ssm_parameter.password]
}

resource "aws_security_group" "rancher_node" {
  name        = "${local.cluster_name}-rancher-node-sg"
  description = "Security group for Rancher managed cluster nodes"
  vpc_id      = data.terraform_remote_state.base.outputs.vpc_id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [data.terraform_remote_state.base.outputs.rancher_server_sg_id]
  }

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.admin_ips
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name                                        = "${local.cluster_name}-rancher-node-sg"
    Environment                                 = var.environment
    Project                                     = var.project_name
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  }
}

resource "aws_iam_role" "rancher_node" {
  name = "${local.cluster_name}-rancher-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })

  tags = {
    Name        = "${local.cluster_name}-rancher-node-role"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_role_policy" "rancher_node_permissive" {
  name = "${local.cluster_name}-rancher-node-policy"
  role = aws_iam_role.rancher_node.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["ec2:*", "elasticloadbalancing:*", "ecr:*", "s3:*", "route53:*", "ebs:*", "iam:*"]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "rancher_node_ssm" {
  role       = aws_iam_role.rancher_node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "rancher_node" {
  name = "${local.cluster_name}-rancher-node-profile"
  role = aws_iam_role.rancher_node.name
}

resource "rancher2_cluster_v2" "main" {
  name               = local.cluster_name
  kubernetes_version = var.kubernetes_version

  rke_config {
    machine_pools {
      name                         = "all-roles-pool"
      cloud_credential_secret_name = rancher2_cloud_credential.aws.id
      control_plane_role           = true
      etcd_role                    = true
      worker_role                  = true
      quantity                     = var.stopped ? 0 : var.worker_desired_capacity
      max_unhealthy                = "100%"

      machine_config {
        kind = rancher2_machine_config_v2.cluster_nodes.kind
        name = rancher2_machine_config_v2.cluster_nodes.name
      }

      rolling_update {
        max_unavailable = "1"
        max_surge       = "1"
      }
    }

    machine_global_config = <<-EOF
      cni: "canal"
      disable:
        - traefik
      tls-san:
        - "${local.cluster_name}.${var.domain_name}"
    EOF
  }

  depends_on = [
    rancher2_bootstrap.admin,
    rancher2_machine_config_v2.cluster_nodes
  ]
}

resource "local_file" "kubeconfig_rancher" {
  filename        = "${path.module}/../../kubeconfig.yaml"
  content         = rancher2_cluster_v2.main.kube_config
  file_permission = "0600"

  depends_on = [rancher2_cluster_v2.main]
}

resource "aws_route53_record" "cluster_wildcard" {
  zone_id = data.terraform_remote_state.base.outputs.route53_zone_id
  name    = "*.${var.domain_name}"
  type    = "CNAME"
  ttl     = 300
  records = [data.terraform_remote_state.base.outputs.rancher_server_fqdn]
}
