provider "rancher2" {
  alias     = "bootstrap"
  api_url   = "https://rancher.${var.domain_name}"
  bootstrap = true
  insecure  = true
}

resource "rancher2_bootstrap" "admin" {
  provider = rancher2.bootstrap
  count    = var.rancher_cluster_enabled ? 1 : 0

  initial_password = random_password.rancher_admin.result
  password         = random_password.rancher_admin.result

  depends_on = [
    aws_instance.rancher_server,
    aws_eip_association.rancher_server
  ]
}

provider "rancher2" {
  api_url   = "https://rancher.${var.domain_name}"
  bootstrap = false
  insecure  = true
  token_key = try(rancher2_bootstrap.admin[0].token, "")
}
resource "rancher2_setting" "agenttlsmode" {
    count = var.rancher_cluster_enabled ? 1 : 0
  depends_on = [rancher2_bootstrap.admin]

  name = "agent-tls-mode"
  value = "system-store"
}
resource "rancher2_cloud_credential" "aws" {
  count = var.rancher_cluster_enabled ? 1 : 0

  name = "${local.cluster_name}-aws-creds"

  amazonec2_credential_config {
    access_key = var.rancher_cluster_enabled ? data.aws_caller_identity.current.account_id : ""
    secret_key = ""
  }

  depends_on = [rancher2_bootstrap.admin]
}

resource "aws_ssm_parameter" "password" {
  name = "/rancher_admin"
  type = "String"
  value =   random_password.rancher_admin.result

}

resource "rancher2_machine_config_v2" "control_plane_nodes" {
  count = var.rancher_cluster_enabled ? 1 : 0

  generate_name = "${local.cluster_name}-control-plane"

  amazonec2_config {
    ami                   = data.aws_ami.amazon_linux_2.id
    region                = var.aws_region
    security_group        = [aws_security_group.rancher_node[0].name]
    subnet_id             = aws_subnet.public[0].id
    vpc_id                = aws_vpc.main.id
    zone                  = "a"
    instance_type         = "t4g.medium"
    root_size             = "30"
    iam_instance_profile  = aws_iam_instance_profile.rancher_node[0].name
    ssh_user              = "ec2-user"
    request_spot_instance = false
  }

  depends_on = [rancher2_bootstrap.admin, aws_ssm_parameter.password]
}

resource "rancher2_machine_config_v2" "worker_nodes" {
  count = var.rancher_cluster_enabled ? 1 : 0

  generate_name = "${local.cluster_name}-worker"

data "rancher2_cluster_v2" "local" {
  count = var.rancher_cluster_enabled ? 1 : 0
  name  = "local"

  depends_on = [rancher2_bootstrap.admin, aws_ssm_parameter.password]
}

resource "aws_security_group" "rancher_node" {
  count = var.rancher_cluster_enabled ? 1 : 0

  name        = "${local.cluster_name}-rancher-node-sg"
  description = "Security group for Rancher managed cluster nodes"
  vpc_id      = aws_vpc.main.id

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
    security_groups = [aws_security_group.rancher_server.id]
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
    Name                                          = "${local.cluster_name}-rancher-node-sg"
    Environment                                   = var.environment
    Project                                       = var.project_name
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  }
}

resource "aws_iam_role" "rancher_node" {
  count = var.rancher_cluster_enabled ? 1 : 0

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
  count = var.rancher_cluster_enabled ? 1 : 0

  name = "${local.cluster_name}-rancher-node-policy"
  role = aws_iam_role.rancher_node[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["ec2:*", "elasticloadbalancing:*", "ecr:*", "s3:*", "route53:*", "ebs:*","iam:*"]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "rancher_node_ssm" {
  count = var.rancher_cluster_enabled ? 1 : 0

  role       = aws_iam_role.rancher_node[0].name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "rancher_node" {
  count = var.rancher_cluster_enabled ? 1 : 0

  name = "${local.cluster_name}-rancher-node-profile"
  role = aws_iam_role.rancher_node[0].name
}

resource "rancher2_node_pool" "worker_spot_pool" {
  count = var.rancher_cluster_enabled ? 1 : 0

  name               = local.cluster_name
  kubernetes_version = var.kubernetes_version

  rke_config {
    machine_pools {
      name                         = "control-plane-pool"
      cloud_credential_secret_name = rancher2_cloud_credential.aws[0].id
      control_plane_role           = true
      etcd_role                    = true
      worker_role                  = false
      quantity                     = var.stopped ? 0 : 1
      max_unhealthy                = "100%"

      machine_config {
        kind = rancher2_machine_config_v2.control_plane_nodes[0].kind
        name = rancher2_machine_config_v2.control_plane_nodes[0].name
      }

      rolling_update {
        max_unavailable = "1"
        max_surge       = "1"
      }
    }

    machine_pools {
      name                         = "worker-spot-pool"
      cloud_credential_secret_name = rancher2_cloud_credential.aws[0].id
      control_plane_role           = false
      etcd_role                    = false
      worker_role                  = true
      quantity                     = var.stopped ? 0 : var.worker_desired_capacity
      max_unhealthy                = "100%"

      machine_config {
        kind = rancher2_machine_config_v2.worker_nodes[0].kind
        name = rancher2_machine_config_v2.worker_nodes[0].name
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
    rancher2_machine_config_v2.control_plane_nodes,
    rancher2_machine_config_v2.worker_nodes
  ]
}

resource "rancher2_node_template" "worker" {
  count = var.rancher_cluster_enabled ? 1 : 0

  name        = "${local.cluster_name}-worker-template"
  description = "Spot worker node template"

  amazonec2_config {
    ami                   = data.aws_ami.amazon_linux_2.id
    region                = var.aws_region
    security_group        = [aws_security_group.rancher_node[0].name]
    subnet_id             = aws_subnet.public[0].id
    vpc_id                = aws_vpc.main.id
    zone                  = "a"
    instance_type         = "t4g.medium"
    root_size             = 30
    iam_instance_profile  = aws_iam_instance_profile.rancher_node[0].name
    ssh_user              = "ec2-user"
    request_spot_instance = true
    spot_price            = "0.05"
  }

  depends_on = [rancher2_bootstrap.admin]
}

resource "local_file" "kubeconfig_rancher" {
  count = var.rancher_cluster_enabled ? 1 : 0

  filename        = "${path.module}/kubeconfig.yaml"
  content         = data.rancher2_cluster_v2.local[0].kube_config
  file_permission = "0600"

  depends_on = [data.rancher2_cluster_v2.local]
}

resource "aws_route53_record" "cluster_wildcard" {
  count = var.rancher_cluster_enabled ? 1 : 0

  zone_id = aws_route53_zone.main.zone_id
  name    = "*.${var.domain_name}"
  type    = "CNAME"
  ttl     = 300
  records = [aws_route53_record.rancher_server.fqdn]
}

output "rancher_cluster_id" {
  description = "ID of the Rancher local cluster"
  value       = var.rancher_cluster_enabled ? data.rancher2_cluster_v2.local[0].id : null
}

output "rancher_cluster_name" {
  description = "Name of the Rancher local cluster"
  value       = var.rancher_cluster_enabled ? data.rancher2_cluster_v2.local[0].name : null
}

output "rancher_admin_token" {
  description = "Rancher admin API token"
  value       = var.rancher_cluster_enabled ? rancher2_bootstrap.admin[0].token : null
  sensitive   = true
}

output "cluster_kubeconfig" {
  description = "Kubeconfig for the local cluster (use terraform output -raw cluster_kubeconfig)"
  value       = var.rancher_cluster_enabled ? data.rancher2_cluster_v2.local[0].kube_config : null
  sensitive   = true
}
