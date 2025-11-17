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

resource "local_file" "rancher_private_key" {
  filename        = "${path.module}/rancher-key.pem"
  content         = tls_private_key.rancher.private_key_pem
  file_permission = "0400"
}

resource "aws_eip" "rancher" {
  domain = "vpc"

  tags = {
    Name        = "${local.cluster_name}-rancher-eip"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_security_group" "rancher_server" {
  name        = "${local.cluster_name}-rancher-server-sg"
  description = "Security group for Rancher Server"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${local.cluster_name}-rancher-server-sg"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_role" "rancher_server" {
  name = "${local.cluster_name}-rancher-server-role"

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
    Name        = "${local.cluster_name}-rancher-server-role"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_role_policy" "rancher_server_permissive" {
  name = "${local.cluster_name}-rancher-server-policy"
  role = aws_iam_role.rancher_server.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["ec2:*", "elasticloadbalancing:*", "ecr:*", "s3:*", "route53:*","iam:*"]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_instance_profile" "rancher_server" {
  name = "${local.cluster_name}-rancher-server-profile"
  role = aws_iam_role.rancher_server.name
}

resource "random_password" "rancher_admin" {
  length  = 16
  special = false
}

resource "aws_instance" "rancher_server" {
  ami                    = data.aws_ami.amazon_linux_2.id
  instance_type          = "t3a.medium"
  key_name               = aws_key_pair.rancher.key_name
  vpc_security_group_ids = [aws_security_group.rancher_server.id]
  subnet_id              = aws_subnet.public[0].id
  iam_instance_profile   = aws_iam_instance_profile.rancher_server.name

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 30
    encrypted             = true
    delete_on_termination = true
  }

  user_data = <<-EOF
#!/bin/bash
set -e
exec > >(tee /var/log/user-data.log) 2>&1

yum update -y
amazon-linux-extras install docker -y
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user

docker run -d --restart=unless-stopped \
  -p 80:80 -p 443:443 \
  --privileged \
  -e CATTLE_BOOTSTRAP_PASSWORD="${random_password.rancher_admin.result}" \
  rancher/rancher:latest

echo "Rancher Server started. Bootstrap password: ${random_password.rancher_admin.result}"
EOF

  tags = {
    Name        = "${local.cluster_name}-rancher-server"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    ignore_changes = [user_data]
  }
}

resource "aws_eip_association" "rancher_server" {
  instance_id   = aws_instance.rancher_server.id
  allocation_id = aws_eip.rancher.id
}

resource "aws_route53_record" "rancher_server" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "rancher.${var.domain_name}"
  type    = "A"
  ttl     = 300
  records = [aws_eip.rancher.public_ip]
}

output "rancher_server_url" {
  description = "URL of the Rancher Server"
  value       = "https://rancher.${var.domain_name}"
}

output "rancher_server_ip" {
  description = "Public IP of Rancher Server"
  value       = aws_eip.rancher.public_ip
}

output "rancher_admin_password" {
  description = "Initial admin password for Rancher"
  value       = random_password.rancher_admin.result
  sensitive   = true
}

output "rancher_ssh_command" {
  description = "SSH command to connect to Rancher Server"
  value       = "ssh -i ${path.module}/rancher-key.pem ec2-user@${aws_eip.rancher.public_ip}"
}

output "rancher_private_key_path" {
  description = "Path to SSH private key"
  value       = local_file.rancher_private_key.filename
}
