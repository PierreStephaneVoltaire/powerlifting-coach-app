#!/bin/bash
set -e

CLUSTER_NAME="${cluster_name}"
S3_BUCKET="${s3_bucket}"
NGINX_LB_IP="${nginx_lb_ip}"
REGION="${region}"
POD_NETWORK_CIDR="${pod_network_cidr}"

cloud-init status --wait

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get upgrade -y
apt-get install -y python3-pip python3-venv awscli jq curl git

pip3 install --upgrade pip
pip3 install ansible

PRIVATE_IP=$(curl -s http://169.254.169.254/latest/meta-data/local-ipv4)
K3S_TOKEN=$(aws ssm get-parameter --name "/$CLUSTER_NAME/k3s/token" --with-decryption --query 'Parameter.Value' --output text --region $REGION)

aws s3 cp s3://$S3_BUCKET/control-plane.yml /tmp/control-plane.yml --region $REGION

export K3S_TOKEN="$K3S_TOKEN"
export NGINX_LB_IP="$NGINX_LB_IP"
export PRIVATE_IP="$PRIVATE_IP"
export POD_NETWORK_CIDR="$POD_NETWORK_CIDR"
export CLUSTER_NAME="$CLUSTER_NAME"
export AWS_REGION="$REGION"

ansible-playbook /tmp/control-plane.yml

rm -f /tmp/control-plane.yml
