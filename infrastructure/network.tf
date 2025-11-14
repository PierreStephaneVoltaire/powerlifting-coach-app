# VPC Configuration - Public subnets only (no NAT gateway for cost savings)

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name              = "${local.cluster_name}-vpc"
    Environment       = var.environment
    Project           = var.project_name
    KubernetesCluster = local.cluster_name
  }
}

# Internet Gateway for public subnet access
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name        = "${local.cluster_name}-igw"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Get availability zones in the region
data "aws_availability_zones" "available" {
  state = "available"
}

# Public subnets across 3 AZs for HA
resource "aws_subnet" "public" {
  count                   = 3
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.${count.index}.0/24"
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name              = "${local.cluster_name}-public-${count.index + 1}"
    Environment       = var.environment
    Project           = var.project_name
    Type              = "public"
    KubernetesCluster = local.cluster_name
    KubernetesRole    = "elb"
  }
}

# Route table for public subnets
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name        = "${local.cluster_name}-public-rt"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Associate route table with public subnets
resource "aws_route_table_association" "public" {
  count          = 3
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Security group for control plane nodes
resource "aws_security_group" "control_plane" {
  name_prefix = "${local.cluster_name}-control-plane-"
  description = "Security group for k3s control plane nodes"
  vpc_id      = aws_vpc.main.id

  # Allow k3s API server from anywhere (public access)
  ingress {
    description = "Kubernetes API server"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow etcd communication between control plane nodes
  ingress {
    description = "etcd client"
    from_port   = 2379
    to_port     = 2380
    protocol    = "tcp"
    self        = true
  }

  # Allow kubelet API
  ingress {
    description = "Kubelet API"
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
    self        = true
  }

  # Allow SSH for management
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound traffic
  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${local.cluster_name}-control-plane-sg"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Security group for worker nodes
resource "aws_security_group" "worker" {
  name_prefix = "${local.cluster_name}-worker-"
  description = "Security group for k3s worker nodes"
  vpc_id      = aws_vpc.main.id

  # Allow kubelet API
  ingress {
    description = "Kubelet API"
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
    self        = true
  }

  # Allow NodePort services
  ingress {
    description = "NodePort services"
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all traffic between workers
  ingress {
    description = "Worker to worker traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = true
  }

  # Allow SSH for management
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound traffic
  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${local.cluster_name}-worker-sg"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Separate security group rules to avoid circular dependencies
resource "aws_security_group_rule" "control_plane_from_worker" {
  type                     = "ingress"
  description              = "Allow all traffic from worker nodes"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.control_plane.id
  source_security_group_id = aws_security_group.worker.id
}

resource "aws_security_group_rule" "worker_from_control_plane" {
  type                     = "ingress"
  description              = "Allow all traffic from control plane"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.worker.id
  source_security_group_id = aws_security_group.control_plane.id
}

resource "aws_ssm_parameter" "control_plane_ips" {
  name  = "/${local.cluster_name}/k3s/control-plane-ips"
  type  = "String"
  value = "initializing"

  tags = {
    Name        = "${local.cluster_name}-control-plane-ips"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    ignore_changes = [value]
  }
}

resource "aws_security_group" "nginx_lb" {
  name_prefix = "${local.cluster_name}-nginx-lb-"
  description = "Security group for nginx load balancer"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "Kubernetes API server"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.nginx_lb_ssh_allowed_cidr]
  }

  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${local.cluster_name}-nginx-lb-sg"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group_rule" "control_plane_from_nginx" {
  type                     = "ingress"
  description              = "Allow traffic from nginx load balancer"
  from_port                = 6443
  to_port                  = 6443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.control_plane.id
  source_security_group_id = aws_security_group.nginx_lb.id
}

resource "aws_iam_role" "nginx_lb" {
  name_prefix = "${local.cluster_name}-nginx-lb-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Name        = "${local.cluster_name}-nginx-lb-role"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_role_policy" "nginx_lb" {
  name_prefix = "${local.cluster_name}-nginx-lb-"
  role        = aws_iam_role.nginx_lb.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameter",
          "ssm:GetParameters"
        ]
        Resource = [
          aws_ssm_parameter.control_plane_ips.arn,
          "arn:${data.aws_partition.current.partition}:ssm:${var.aws_region}:*:parameter/${local.cluster_name}/k3s/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.ansible_playbooks.arn,
          "${aws_s3_bucket.ansible_playbooks.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "ec2:DescribeInstances",
          "ec2:DescribeTags"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "nginx_lb_ssm" {
  role       = aws_iam_role.nginx_lb.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "nginx_lb" {
  name_prefix = "${local.cluster_name}-nginx-lb-"
  role        = aws_iam_role.nginx_lb.name

  tags = {
    Name        = "${local.cluster_name}-nginx-lb-profile"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3_object" "nginx_lb_playbook" {
  bucket = aws_s3_bucket.ansible_playbooks.id
  key    = "nginx-lb.yml"
  source = "${path.module}/ansible/nginx-lb.yml"
  etag   = filemd5("${path.module}/ansible/nginx-lb.yml")
}

resource "aws_instance" "nginx_lb" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t4g.micro"
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.nginx_lb.id]
  iam_instance_profile   = aws_iam_instance_profile.nginx_lb.name

  root_block_device {
    volume_size           = 8
    volume_type           = "gp3"
    delete_on_termination = true
    encrypted             = true
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
    instance_metadata_tags      = "enabled"
  }

  monitoring = true

  user_data = base64encode(templatefile("${path.module}/user-data/nginx-lb.sh", {
    cluster_name = local.cluster_name
    s3_bucket    = aws_s3_bucket.ansible_playbooks.id
    region       = var.aws_region
  }))

  tags = {
    Name        = "${local.cluster_name}-nginx-lb"
    Environment = var.environment
    Project     = var.project_name
    Role        = "nginx-lb"
  }

  lifecycle {
    ignore_changes = [ami]
  }

  depends_on = [
    aws_ssm_parameter.control_plane_ips,
    aws_s3_object.nginx_lb_playbook
  ]
}

resource "aws_eip" "nginx_lb" {
  domain   = "vpc"
  instance = aws_instance.nginx_lb.id

  tags = {
    Name        = "${local.cluster_name}-nginx-lb-eip"
    Environment = var.environment
    Project     = var.project_name
  }

  depends_on = [aws_instance.nginx_lb]
}

resource "aws_iam_role" "nginx_restart_lambda" {
  name_prefix = "${local.cluster_name}-nginx-restart-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Name        = "${local.cluster_name}-nginx-restart-lambda-role"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_iam_role_policy" "nginx_restart_lambda" {
  name_prefix = "${local.cluster_name}-nginx-restart-"
  role        = aws_iam_role.nginx_restart_lambda.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ec2:RebootInstances",
          "ec2:DescribeInstances",
          "ec2:DescribeTags"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:${data.aws_partition.current.partition}:logs:${var.aws_region}:*:log-group:/aws/lambda/*"
      }
    ]
  })
}

resource "aws_lambda_function" "nginx_restart" {
  filename      = "${path.module}/lambda/nginx-restart.zip"
  function_name = "${local.cluster_name}-nginx-restart"
  role          = aws_iam_role.nginx_restart_lambda.arn
  handler       = "index.handler"
  runtime       = "python3.12"
  timeout       = 60

  environment {
    variables = {
      NGINX_INSTANCE_ID = aws_instance.nginx_lb.id
      CLUSTER_NAME      = local.cluster_name
    }
  }

  tags = {
    Name        = "${local.cluster_name}-nginx-restart"
    Environment = var.environment
    Project     = var.project_name
  }

  depends_on = [
    aws_iam_role_policy.nginx_restart_lambda
  ]
}

data "archive_file" "nginx_restart_lambda" {
  type        = "zip"
  output_path = "${path.module}/lambda/nginx-restart.zip"

  source {
    content  = <<-EOF
import boto3
import os
import json

ec2 = boto3.client('ec2')

def handler(event, context):
    print(f"Received event: {json.dumps(event)}")

    instance_id = os.environ['NGINX_INSTANCE_ID']
    cluster_name = os.environ['CLUSTER_NAME']

    if 'detail' in event and 'name' in event['detail']:
        param_name = event['detail']['name']
        expected_param = f"/{cluster_name}/k3s/control-plane-ips"

        if param_name != expected_param:
            print(f"Ignoring parameter change for {param_name}")
            return {
                'statusCode': 200,
                'body': json.dumps('Not the control-plane-ips parameter, ignoring')
            }

    print(f"Rebooting nginx instance {instance_id}")

    try:
        response = ec2.reboot_instances(InstanceIds=[instance_id])
        print(f"Reboot initiated: {response}")

        return {
            'statusCode': 200,
            'body': json.dumps(f'Successfully initiated reboot of {instance_id}')
        }
    except Exception as e:
        print(f"Error rebooting instance: {str(e)}")
        raise
EOF
    filename = "index.py"
  }
}

resource "aws_cloudwatch_event_rule" "parameter_change" {
  name_prefix = "${local.cluster_name}-param-change-"
  description = "Trigger nginx restart when control plane IPs parameter changes"

  event_pattern = jsonencode({
    source      = ["aws.ssm"]
    detail-type = ["Parameter Store Change"]
    detail = {
      name      = [aws_ssm_parameter.control_plane_ips.name]
      operation = ["Update"]
    }
  })

  tags = {
    Name        = "${local.cluster_name}-param-change-rule"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.nginx_restart.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.parameter_change.arn
}

resource "aws_cloudwatch_event_target" "lambda" {
  rule      = aws_cloudwatch_event_rule.parameter_change.name
  target_id = "NginxRestartLambda"
  arn       = aws_lambda_function.nginx_restart.arn
}
