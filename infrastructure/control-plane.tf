data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_s3_bucket" "ansible_playbooks" {
  bucket_prefix = "${local.cluster_name}-ansible-"

  tags = {
    Name        = "${local.cluster_name}-ansible"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_s3_bucket_public_access_block" "ansible_playbooks" {
  bucket = aws_s3_bucket.ansible_playbooks.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_object" "control_plane_playbook" {
  bucket = aws_s3_bucket.ansible_playbooks.id
  key    = "control-plane.yml"
  source = "${path.module}/ansible/control-plane.yml"
  etag   = filemd5("${path.module}/ansible/control-plane.yml")
}

resource "aws_s3_object" "worker_playbook" {
  bucket = aws_s3_bucket.ansible_playbooks.id
  key    = "worker.yml"
  source = "${path.module}/ansible/worker.yml"
  etag   = filemd5("${path.module}/ansible/worker.yml")
}

resource "random_password" "k3s_token" {
  length  = 32
  special = false
}

resource "aws_ssm_parameter" "k3s_token" {
  name  = "/${local.cluster_name}/k3s/token"
  type  = "SecureString"
  value = random_password.k3s_token.result

  tags = {
    Name        = "${local.cluster_name}-k3s-token"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_lb" "control_plane" {
  name_prefix                      = "cp-"
  internal                         = false
  load_balancer_type               = "network"
  subnets                          = aws_subnet.public[*].id
  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = true

  tags = {
    Name        = "${local.cluster_name}-control-plane-nlb"
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "aws_lb_target_group" "control_plane_api" {
  name_prefix          = "api-"
  port                 = 6443
  protocol             = "TCP"
  vpc_id               = aws_vpc.main.id
  deregistration_delay = 30

  health_check {
    enabled             = true
    protocol            = "TCP"
    port                = "6443"
    interval            = 10
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }

  tags = {
    Name        = "${local.cluster_name}-control-plane-api-tg"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lb_listener" "control_plane_api" {
  load_balancer_arn = aws_lb.control_plane.arn
  port              = "6443"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.control_plane_api.arn
  }
}

resource "aws_launch_template" "control_plane" {
  name_prefix   = "${local.cluster_name}-control-plane-"
  image_id      = data.aws_ami.ubuntu.id
  instance_type = var.control_plane_instance_type

  iam_instance_profile {
    arn = aws_iam_instance_profile.control_plane.arn
  }

  vpc_security_group_ids = [aws_security_group.control_plane.id]

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size           = var.control_plane_volume_size
      volume_type           = "gp3"
      iops                  = 3000
      throughput            = 125
      delete_on_termination = true
      encrypted             = true
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
    instance_metadata_tags      = "enabled"
  }

  monitoring {
    enabled = true
  }

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name                                          = "${local.cluster_name}-control-plane"
      Environment                                   = var.environment
      Project                                       = var.project_name
      Role                                          = "control-plane"
      "kubernetes.io/cluster/${local.cluster_name}" = "owned"
    }
  }

  tag_specifications {
    resource_type = "volume"

    tags = {
      Name        = "${local.cluster_name}-control-plane-volume"
      Environment = var.environment
      Project     = var.project_name
    }
  }

  user_data = base64encode(templatefile("${path.module}/user-data/control-plane.sh", {
    cluster_name     = local.cluster_name
    s3_bucket        = aws_s3_bucket.ansible_playbooks.id
    nlb_dns_name     = aws_lb.control_plane.dns_name
    region           = var.aws_region
    pod_network_cidr = var.pod_network_cidr
  }))

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    aws_ssm_parameter.k3s_token,
    aws_s3_object.control_plane_playbook
  ]
}

resource "aws_autoscaling_group" "control_plane" {
  name_prefix               = "${local.cluster_name}-control-plane-"
  vpc_zone_identifier       = aws_subnet.public[*].id
  target_group_arns         = [aws_lb_target_group.control_plane_api.arn]
  desired_capacity          = 3
  min_size                  = 3
  max_size                  = 3
  health_check_type         = "ELB"
  health_check_grace_period = 300
  default_cooldown          = 300

  enabled_metrics = [
    "GroupDesiredCapacity",
    "GroupInServiceInstances",
    "GroupMaxSize",
    "GroupMinSize",
    "GroupPendingInstances",
    "GroupStandbyInstances",
    "GroupTerminatingInstances",
    "GroupTotalInstances"
  ]

  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity                  = 0
      on_demand_percentage_above_base_capacity = 0
      spot_allocation_strategy                 = "price-capacity-optimized"
      spot_instance_pools                      = 0
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.control_plane.id
        version            = "$Latest"
      }

      override {
        instance_type     = "t3a.small"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "1"
      }
    }
  }

  tag {
    key                 = "Name"
    value               = "${local.cluster_name}-control-plane"
    propagate_at_launch = true
  }

  tag {
    key                 = "Environment"
    value               = var.environment
    propagate_at_launch = true
  }

  tag {
    key                 = "Project"
    value               = var.project_name
    propagate_at_launch = true
  }

  tag {
    key                 = "Role"
    value               = "control-plane"
    propagate_at_launch = true
  }

  tag {
    key                 = "kubernetes.io/cluster/${local.cluster_name}"
    value               = "owned"
    propagate_at_launch = true
  }

  lifecycle {
    create_before_destroy = true
    ignore_changes        = [desired_capacity]
  }

  depends_on = [
    aws_lb_listener.control_plane_api
  ]
}
