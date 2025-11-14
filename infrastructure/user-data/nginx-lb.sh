#!/bin/bash
set -e

CLUSTER_NAME="${cluster_name}"
S3_BUCKET="${s3_bucket}"
REGION="${region}"

# Wait for cloud-init to complete
cloud-init status --wait

# Update and install required packages
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get upgrade -y
apt-get install -y python3-pip python3-venv awscli jq curl

# Install ansible
pip3 install --upgrade pip
pip3 install ansible

# Download nginx playbook from S3
aws s3 cp s3://$S3_BUCKET/nginx-lb.yml /tmp/nginx-lb.yml --region $REGION

# Set environment variables for ansible
export CLUSTER_NAME="$CLUSTER_NAME"
export AWS_REGION="$REGION"

# Run ansible playbook
ansible-playbook /tmp/nginx-lb.yml

# Clean up
rm -f /tmp/nginx-lb.yml

echo "Nginx load balancer setup complete"
