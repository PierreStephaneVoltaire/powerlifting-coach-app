# Rancher k3s Single Node Setup
# Cost-effective alternative to EKS with single on-demand instance

# Get latest Amazon Linux 2 AMI
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

# SSH Key Pair for EC2 access
resource "tls_private_key" "rancher" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "rancher" {
  key_name   = "${local.cluster_name}-rancher-key"
  public_key = tls_private_key.rancher.public_key_openssh

  tags = {
    Name        = "${local.cluster_name}-rancher-key"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Store private key locally for SSH access
resource "local_file" "rancher_private_key" {
  filename        = "${path.module}/rancher-key.pem"
  content         = tls_private_key.rancher.private_key_pem
  file_permission = "0400"
}

# IAM Role for Rancher EC2 Instance (permissive as requested)
resource "aws_iam_role" "rancher" {
  name = "${local.cluster_name}-rancher-role"

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
    Name        = "${local.cluster_name}-rancher-role"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Permissive IAM Policy for Rancher node
resource "aws_iam_role_policy" "rancher_permissive" {
  name = "${local.cluster_name}-rancher-permissive"
  role = aws_iam_role.rancher.id

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
        Resource = [
          aws_s3_bucket.videos.arn,
          "${aws_s3_bucket.videos.arn}/*",
          "arn:aws:s3:::*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "ebs:*"
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
          "acm:*"
        ]
        Resource = "*"
      }
    ]
  })
}

# Instance Profile for EC2
resource "aws_iam_instance_profile" "rancher" {
  name = "${local.cluster_name}-rancher-profile"
  role = aws_iam_role.rancher.name

  tags = {
    Name        = "${local.cluster_name}-rancher-profile"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Security Group for Rancher/k3s
resource "aws_security_group" "rancher" {
  name        = "${local.cluster_name}-rancher-sg"
  description = "Security group for Rancher k3s node"
  vpc_id      = aws_vpc.main.id

  # SSH access
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Kubernetes API Server
  ingress {
    description = "Kubernetes API"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTP for ingress
  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS for ingress and Rancher UI
  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Rancher Webhook
  ingress {
    description = "Rancher Webhook"
    from_port   = 8443
    to_port     = 8443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # NodePort range
  ingress {
    description = "NodePort Services"
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name                                          = "${local.cluster_name}-rancher-sg"
    Environment                                   = var.environment
    Project                                       = var.project_name
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  }
}

# Elastic IP for stable access
resource "aws_eip" "rancher" {
  domain = "vpc"

  tags = {
    Name        = "${local.cluster_name}-rancher-eip"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Associate EIP with instance after creation
resource "aws_eip_association" "rancher" {
  instance_id   = aws_instance.rancher.id
  allocation_id = aws_eip.rancher.id
}

# User data script for k3s installation only
# Monitoring and other addons are installed via Terraform Helm releases
locals {
  rancher_user_data = <<-EOF
#!/bin/bash
set -e

# Log output to file for debugging
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1

echo "Starting k3s installation..."

# Update system
yum update -y
yum install -y jq curl wget

# Install Docker (for pulling images)
amazon-linux-extras install docker -y
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user

# Set hostname
hostnamectl set-hostname ${local.cluster_name}-rancher

# Install k3s with traefik and servicelb disabled (using NGINX ingress instead)
curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server \
  --disable traefik \
  --disable servicelb \
  --write-kubeconfig-mode 644 \
  --tls-san ${aws_eip.rancher.public_ip} \
  --node-name ${local.cluster_name}-rancher \
  --kube-apiserver-arg default-not-ready-toleration-seconds=30 \
  --kube-apiserver-arg default-unreachable-toleration-seconds=30" sh -

# Wait for k3s to be ready
echo "Waiting for k3s to be ready..."
until kubectl get nodes 2>/dev/null | grep -q "Ready"; do
  echo "Waiting for k3s to be ready..."
  sleep 10
done

echo "k3s is ready!"

# Make local-path storage class the default
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

# Copy kubeconfig to ec2-user home
mkdir -p /home/ec2-user/.kube
cp /etc/rancher/k3s/k3s.yaml /home/ec2-user/.kube/config
chown -R ec2-user:ec2-user /home/ec2-user/.kube
chmod 600 /home/ec2-user/.kube/config

# Update kubeconfig with external IP
sed -i "s/127.0.0.1/${aws_eip.rancher.public_ip}/g" /home/ec2-user/.kube/config

# Save kubeconfig to S3 for Terraform to use
aws s3 cp /home/ec2-user/.kube/config s3://${aws_s3_bucket.videos.id}/kubeconfig.yaml

echo "============================================"
echo "k3s installation complete!"
echo "============================================"
echo "Kubernetes API: https://${aws_eip.rancher.public_ip}:6443"
echo "Kubeconfig uploaded to S3"
echo "============================================"
echo "Next steps:"
echo "1. Download kubeconfig: aws s3 cp s3://${aws_s3_bucket.videos.id}/kubeconfig.yaml infrastructure/kubeconfig.yaml"
echo "2. Set kubernetes_resources_enabled = true"
echo "3. Run terraform apply to install monitoring, ArgoCD, etc."
echo "============================================"
EOF
}

# EC2 Instance for Rancher/k3s
resource "aws_instance" "rancher" {
  ami                    = data.aws_ami.amazon_linux_2.id
  instance_type          = "t3a.medium"  # On-demand, cost effective
  key_name               = aws_key_pair.rancher.key_name
  vpc_security_group_ids = [aws_security_group.rancher.id]
  subnet_id              = aws_subnet.public[0].id
  iam_instance_profile   = aws_iam_instance_profile.rancher.name

  # Root volume - 30GB gp3
  root_block_device {
    volume_type           = "gp3"
    volume_size           = 30
    encrypted             = true
    delete_on_termination = true
  }

  user_data = local.rancher_user_data

  tags = {
    Name                                          = "${local.cluster_name}-rancher"
    Environment                                   = var.environment
    Project                                       = var.project_name
    "kubernetes.io/cluster/${local.cluster_name}" = "owned"
  }

  lifecycle {
    ignore_changes = [user_data]
  }
}

# Route53 DNS records for Rancher
resource "aws_route53_record" "rancher" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "rancher.${var.domain_name}"
  type    = "A"
  ttl     = 300
  records = [aws_eip.rancher.public_ip]
}

# Wildcard DNS for all services pointing to EIP
resource "aws_route53_record" "rancher_wildcard" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "*.${var.domain_name}"
  type    = "A"
  ttl     = 300
  records = [aws_eip.rancher.public_ip]
}

# Outputs for Rancher
output "rancher_public_ip" {
  description = "Public IP of the Rancher k3s node"
  value       = aws_eip.rancher.public_ip
}

output "rancher_instance_id" {
  description = "Instance ID of the Rancher k3s node"
  value       = aws_instance.rancher.id
}

output "rancher_ssh_command" {
  description = "SSH command to connect to Rancher node"
  value       = "ssh -i ${path.module}/rancher-key.pem ec2-user@${aws_eip.rancher.public_ip}"
}

output "rancher_ui_url" {
  description = "Rancher UI URL"
  value       = "https://rancher.${var.domain_name}"
}

output "rancher_bootstrap_password" {
  description = "Initial bootstrap password for Rancher (change after first login)"
  value       = "admin"
}

output "k3s_kubeconfig_s3_path" {
  description = "S3 path to download kubeconfig"
  value       = "s3://${aws_s3_bucket.videos.id}/kubeconfig.yaml"
}

output "rancher_private_key_path" {
  description = "Path to private key for SSH access"
  value       = local_file.rancher_private_key.filename
}
