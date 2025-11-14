# Nginx Load Balancer for Control Plane (replaces expensive NLB)
# This creates a cost-effective t4g.micro instance running nginx to load balance
# control plane nodes. Control plane instances update their IPs in Parameter Store,
# triggering a Lambda to restart nginx which pulls updated IPs on startup.

# SSM Parameter to store control plane node IPs (comma-separated)
resource "aws_ssm_parameter" "control_plane_ips" {
  name  = "/${local.cluster_name}/k3s/control-plane-ips"
  type  = "String"
  value = "initializing" # Will be updated by control plane instances

  tags = {
    Name        = "${local.cluster_name}-control-plane-ips"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    ignore_changes = [value] # Control plane instances will update this
  }
}

# Security group for nginx load balancer
resource "aws_security_group" "nginx_lb" {
  name_prefix = "${local.cluster_name}-nginx-lb-"
  description = "Security group for nginx load balancer"
  vpc_id      = aws_vpc.main.id

  # Allow Kubernetes API server from anywhere
  ingress {
    description = "Kubernetes API server"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
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
    Name        = "${local.cluster_name}-nginx-lb-sg"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Allow nginx to connect to control plane API
resource "aws_security_group_rule" "control_plane_from_nginx" {
  type                     = "ingress"
  description              = "Allow traffic from nginx load balancer"
  from_port                = 6443
  to_port                  = 6443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.control_plane.id
  source_security_group_id = aws_security_group.nginx_lb.id
}

# IAM role for nginx load balancer
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

# IAM policy for nginx load balancer
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

# Attach SSM managed instance core policy for systems manager access
resource "aws_iam_role_policy_attachment" "nginx_lb_ssm" {
  role       = aws_iam_role.nginx_lb.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# Instance profile for nginx load balancer
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

# Upload nginx ansible playbook to S3
resource "aws_s3_object" "nginx_lb_playbook" {
  bucket = aws_s3_bucket.ansible_playbooks.id
  key    = "nginx-lb.yml"
  source = "${path.module}/ansible/nginx-lb.yml"
  etag   = filemd5("${path.module}/ansible/nginx-lb.yml")
}

# Nginx load balancer instance (on-demand for reliability)
resource "aws_instance" "nginx_lb" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t4g.micro" # ARM-based, free tier eligible
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.nginx_lb.id]
  iam_instance_profile   = aws_iam_instance_profile.nginx_lb.name

  root_block_device {
    volume_size           = 8 # Minimum size for free tier
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
    ignore_changes = [ami] # Don't replace instance on AMI updates
  }

  depends_on = [
    aws_ssm_parameter.control_plane_ips,
    aws_s3_object.nginx_lb_playbook
  ]
}

# Elastic IP for nginx load balancer (static IP for kubeconfig)
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

# Lambda execution role
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

# Lambda policy for restarting nginx instance
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

# Lambda function to restart nginx instance when Parameter Store changes
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

# Create the Lambda deployment package
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

    # Check if the parameter that changed is the control-plane-ips
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

# CloudWatch Events rule to trigger Lambda on Parameter Store changes
resource "aws_cloudwatch_event_rule" "parameter_change" {
  name_prefix = "${local.cluster_name}-param-change-"
  description = "Trigger nginx restart when control plane IPs parameter changes"

  event_pattern = jsonencode({
    source      = ["aws.ssm"]
    detail-type = ["Parameter Store Change"]
    detail = {
      name = [aws_ssm_parameter.control_plane_ips.name]
      operation = ["Update"]
    }
  })

  tags = {
    Name        = "${local.cluster_name}-param-change-rule"
    Environment = var.environment
    Project     = var.project_name
  }
}

# Allow EventBridge to invoke Lambda
resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.nginx_restart.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.parameter_change.arn
}

# EventBridge target to invoke Lambda
resource "aws_cloudwatch_event_target" "lambda" {
  rule      = aws_cloudwatch_event_rule.parameter_change.name
  target_id = "NginxRestartLambda"
  arn       = aws_lambda_function.nginx_restart.arn
}
