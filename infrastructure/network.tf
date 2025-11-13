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

# Security group for NLB
resource "aws_security_group" "nlb" {
  name_prefix = "${local.cluster_name}-nlb-"
  description = "Security group for Network Load Balancer"
  vpc_id      = aws_vpc.main.id

  # Allow k3s API server from anywhere
  ingress {
    description = "Kubernetes API server"
    from_port   = 6443
    to_port     = 6443
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
    Name        = "${local.cluster_name}-nlb-sg"
    Environment = var.environment
    Project     = var.project_name
  }

  lifecycle {
    create_before_destroy = true
  }
}
