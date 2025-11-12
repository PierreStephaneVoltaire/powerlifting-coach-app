#!/bin/bash
set -e

CLUSTER_NAME="${cluster_name}"
S3_BUCKET="${s3_bucket}"
NLB_DNS_NAME="${nlb_dns_name}"
REGION="${region}"
MAX_PODS="${max_pods}"

cloud-init status --wait

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get upgrade -y
apt-get install -y python3-pip python3-venv awscli jq curl git

pip3 install --upgrade pip
pip3 install ansible

PRIVATE_IP=$(curl -s http://169.254.169.254/latest/meta-data/local-ipv4)

for i in {1..30}; do
    K3S_TOKEN=$(aws ssm get-parameter --name "/$CLUSTER_NAME/k3s/token" --with-decryption --query 'Parameter.Value' --output text --region $REGION 2>/dev/null) && break || sleep 10
done

aws s3 cp s3://$S3_BUCKET/worker.yml /tmp/worker.yml --region $REGION

export K3S_TOKEN="$K3S_TOKEN"
export NLB_DNS_NAME="$NLB_DNS_NAME"
export PRIVATE_IP="$PRIVATE_IP"
export MAX_PODS="$MAX_PODS"

ansible-playbook /tmp/worker.yml

rm -f /tmp/worker.yml
