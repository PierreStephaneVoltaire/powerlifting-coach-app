# k3s Worker Nodes Configuration
# Auto-scaling spot instances with max-pods=110

# Launch template for worker nodes
resource "aws_launch_template" "worker" {
  name_prefix   = "${local.cluster_name}-worker-"
  image_id      = data.aws_ami.ubuntu.id
  instance_type = var.worker_instance_type

  iam_instance_profile {
    arn = aws_iam_instance_profile.worker.arn
  }

  vpc_security_group_ids = [aws_security_group.worker.id]

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size           = var.worker_volume_size
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
      Name        = "${local.cluster_name}-worker"
      Environment = var.environment
      Project     = var.project_name
      Role        = "worker"
      "kubernetes.io/cluster/${local.cluster_name}" = "owned"
    }
  }

  tag_specifications {
    resource_type = "volume"

    tags = {
      Name        = "${local.cluster_name}-worker-volume"
      Environment = var.environment
      Project     = var.project_name
    }
  }

  user_data = base64encode(templatefile("${path.module}/user-data/worker.sh", {
    cluster_name = local.cluster_name
    s3_bucket    = aws_s3_bucket.k3s_config.id
    nlb_dns_name = aws_lb.control_plane.dns_name
    region       = var.aws_region
    max_pods     = var.max_pods_per_node
  }))

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    aws_autoscaling_group.control_plane
  ]
}

# Auto Scaling Group for worker nodes
resource "aws_autoscaling_group" "worker" {
  name_prefix         = "${local.cluster_name}-worker-"
  vpc_zone_identifier = aws_subnet.public[*].id

  desired_capacity = var.worker_desired_capacity
  min_size         = var.worker_min_size
  max_size         = var.worker_max_size

  health_check_type         = "EC2"
  health_check_grace_period = 300
  default_cooldown          = 60

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
        launch_template_id = aws_launch_template.worker.id
        version            = "$Latest"
      }

      # Multiple cheap spot instance types for flexibility
      # Prefer t3a (AMD) and t4g (ARM) for cost savings
      override {
        instance_type     = "t3a.small"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t3a.medium"
        weighted_capacity = "4"
      }

      override {
        instance_type     = "t3.medium"
        weighted_capacity = "4"
      }

      override {
        instance_type     = "t2.small"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t2.medium"
        weighted_capacity = "4"
      }
    }
  }

  tag {
    key                 = "Name"
    value               = "${local.cluster_name}-worker"
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
    value               = "worker"
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
    aws_autoscaling_group.control_plane
  ]
}

# Auto scaling policies for worker nodes based on CPU utilization
resource "aws_autoscaling_policy" "worker_scale_up" {
  name                   = "${local.cluster_name}-worker-scale-up"
  scaling_adjustment     = 1
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 60
  autoscaling_group_name = aws_autoscaling_group.worker.name
}

resource "aws_autoscaling_policy" "worker_scale_down" {
  name                   = "${local.cluster_name}-worker-scale-down"
  scaling_adjustment     = -1
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  autoscaling_group_name = aws_autoscaling_group.worker.name
}

# CloudWatch alarms for auto scaling
resource "aws_cloudwatch_metric_alarm" "worker_cpu_high" {
  alarm_name          = "${local.cluster_name}-worker-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Average"
  threshold           = "70"
  alarm_description   = "Scale up if CPU exceeds 70%"
  alarm_actions       = [aws_autoscaling_policy.worker_scale_up.arn]

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.worker.name
  }
}

resource "aws_cloudwatch_metric_alarm" "worker_cpu_low" {
  alarm_name          = "${local.cluster_name}-worker-cpu-low"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "30"
  alarm_description   = "Scale down if CPU below 30%"
  alarm_actions       = [aws_autoscaling_policy.worker_scale_down.arn]

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.worker.name
  }
}
