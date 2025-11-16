# RKE2 Cluster on AWS using official Rancher Federal module
# Replaces EKS with RKE2 for cost optimization

# Get latest RHEL 8 AMI (recommended for RKE2)
data "aws_ami" "rhel8" {
  most_recent = true
  owners      = ["309956199498"] # Red Hat

  filter {
    name   = "name"
    values = ["RHEL-8*_HVM-*-x86_64-*-Hourly2-GP3"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Or use Amazon Linux 2 (cheaper)
data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# SSH Key Pair for EC2 access (for debugging)
resource "tls_private_key" "rke2" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "rke2" {
  key_name   = "${local.cluster_name}-rke2-key"
  public_key = tls_private_key.rke2.public_key_openssh

  tags = {
    Name        = "${local.cluster_name}-rke2-key"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "local_file" "rke2_private_key" {
  filename        = "${path.module}/rke2-key.pem"
  content         = tls_private_key.rke2.private_key_pem
  file_permission = "0400"
}

# RKE2 Cluster using official Rancher Federal module
module "rke2" {
  source = "git::https://github.com/rancherfederal/rke2-aws-tf.git?ref=v2.7.0"

  cluster_name = local.cluster_name
  vpc_id       = aws_vpc.main.id
  subnets      = aws_subnet.public[*].id
  ami          = data.aws_ami.amazon_linux_2.id

  # Single server node (cost optimization)
  servers = 1

  # Instance configuration
  instance_type = "t3a.medium"
  block_device_mappings = {
    size      = 30
    encrypted = true
    type      = "gp3"
  }

  # SSH key for debugging
  ssh_authorized_keys = [tls_private_key.rke2.public_key_openssh]

  # RKE2 configuration
  rke2_version = "v1.28.4+rke2r1"

  # Disable built-in ingress controller (we'll use NGINX)
  rke2_config = <<-EOT
    disable:
      - rke2-ingress-nginx
    tls-san:
      - ${local.cluster_name}.${var.domain_name}
  EOT

  # Enable public access
  controlplane_internal = false

  # IAM configuration - permissive as requested
  iam_permissions_boundary = null
  iam_instance_profile     = aws_iam_instance_profile.rke2.name

  # Tagging
  tags = {
    Environment                                   = var.environment
    Project                                       = var.project_name
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  }
}

# Permissive IAM Role for RKE2 nodes
resource "aws_iam_role" "rke2" {
  name = "${local.cluster_name}-rke2-role"

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
    Name        = "${local.cluster_name}-rke2-role"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_role_policy" "rke2_permissive" {
  name = "${local.cluster_name}-rke2-permissive"
  role = aws_iam_role.rke2.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ec2:*",
          "elasticloadbalancing:*",
          "ecr:*",
          "iam:CreateServiceLinkedRole",
          "iam:PassRole",
          "kms:DescribeKey",
          "kms:CreateGrant"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:*"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "route53:*"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameter",
          "ssm:GetParameters",
          "ssm:GetParametersByPath"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_instance_profile" "rke2" {
  name = "${local.cluster_name}-rke2-profile"
  role = aws_iam_role.rke2.name

  tags = {
    Name        = "${local.cluster_name}-rke2-profile"
    Environment = var.environment
    Project     = var.project_name
  }
}

# DNS record for the control plane
resource "aws_route53_record" "rke2_api" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api.${var.domain_name}"
  type    = "CNAME"
  ttl     = 300
  records = [module.rke2.server_url]
}

# Wildcard DNS for services (pointing to NLB)
resource "aws_route53_record" "rke2_wildcard" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "*.${var.domain_name}"
  type    = "CNAME"
  ttl     = 300
  records = [module.rke2.server_url]
}

# Save kubeconfig to local file
resource "local_file" "kubeconfig" {
  count           = 1
  filename        = "${path.module}/kubeconfig.yaml"
  content         = module.rke2.kubeconfig
  file_permission = "0600"
}

# Outputs
output "rke2_cluster_name" {
  description = "Name of the RKE2 cluster"
  value       = module.rke2.cluster_name
}

output "rke2_server_url" {
  description = "RKE2 server URL (NLB endpoint)"
  value       = module.rke2.server_url
}

output "rke2_cluster_data" {
  description = "Cluster data for adding agent nodes"
  value       = module.rke2.cluster_data
  sensitive   = true
}

output "rke2_kubeconfig_path" {
  description = "Path to the kubeconfig file"
  value       = local_file.kubeconfig[0].filename
}

output "rke2_ssh_command" {
  description = "SSH command to connect to RKE2 server"
  value       = "ssh -i ${path.module}/rke2-key.pem ec2-user@<instance-ip>"
}

output "rke2_private_key_path" {
  description = "Path to private key for SSH access"
  value       = local_file.rke2_private_key.filename
}
