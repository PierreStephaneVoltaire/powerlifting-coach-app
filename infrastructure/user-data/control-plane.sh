#!/bin/bash
set -e

# k3s Control Plane Node Bootstrap Script
# This script installs Ansible and runs the control plane playbook

# Variables from Terraform
CLUSTER_NAME="${cluster_name}"
S3_BUCKET="${s3_bucket}"
NLB_DNS_NAME="${nlb_dns_name}"
REGION="${region}"
POD_NETWORK_CIDR="${pod_network_cidr}"

# Wait for cloud-init to complete
cloud-init status --wait

# Update system
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get upgrade -y

# Install dependencies
apt-get install -y \
    software-properties-common \
    python3-pip \
    python3-venv \
    awscli \
    jq \
    curl \
    git

# Install Ansible
pip3 install --upgrade pip
pip3 install ansible

# Get k3s token from S3
aws s3 cp s3://$${S3_BUCKET}/k3s-token /tmp/k3s-token --region $${REGION}
K3S_TOKEN=$(cat /tmp/k3s-token)

# Get instance private IP
PRIVATE_IP=$(curl -s http://169.254.169.254/latest/meta-data/local-ipv4)

# Create Ansible playbook for control plane setup
cat > /tmp/control-plane-playbook.yml <<'ANSIBLE_EOF'
---
- name: Setup k3s control plane node
  hosts: localhost
  become: yes
  vars:
    k3s_version: "v1.28.5+k3s1"
    k3s_token: "{{ lookup('env', 'K3S_TOKEN') }}"
    nlb_dns_name: "{{ lookup('env', 'NLB_DNS_NAME') }}"
    private_ip: "{{ lookup('env', 'PRIVATE_IP') }}"
    pod_cidr: "{{ lookup('env', 'POD_NETWORK_CIDR') }}"
    cluster_name: "{{ lookup('env', 'CLUSTER_NAME') }}"

  tasks:
    - name: Check if k3s is already installed
      stat:
        path: /usr/local/bin/k3s
      register: k3s_binary

    - name: Check if this node is already initialized
      stat:
        path: /var/lib/rancher/k3s/server/node-token
      register: k3s_initialized
      when: k3s_binary.stat.exists

    - name: Install k3s server
      shell: |
        curl -sfL https://get.k3s.io | sh -s - server \
          --token="{{ k3s_token }}" \
          --tls-san="{{ nlb_dns_name }}" \
          --tls-san="{{ private_ip }}" \
          --node-ip="{{ private_ip }}" \
          --cluster-init \
          --disable=traefik \
          --disable=servicelb \
          --write-kubeconfig-mode=644 \
          --kubelet-arg="max-pods=110" \
          --cluster-cidr="{{ pod_cidr }}" \
          --node-label="node-role.kubernetes.io/control-plane=true" \
          --node-taint="node-role.kubernetes.io/control-plane:NoSchedule"
      environment:
        INSTALL_K3S_VERSION: "{{ k3s_version }}"
      when: not k3s_binary.stat.exists or not k3s_initialized.stat.exists

    - name: Wait for k3s to be ready
      wait_for:
        path: /etc/rancher/k3s/k3s.yaml
        state: present
        timeout: 300

    - name: Check if node is ready
      shell: k3s kubectl get nodes | grep -i ready
      register: node_ready
      retries: 30
      delay: 10
      until: node_ready.rc == 0
      changed_when: false

    - name: Create kubeconfig for uploading to S3
      shell: |
        sed "s/127.0.0.1/{{ nlb_dns_name }}/g" /etc/rancher/k3s/k3s.yaml > /tmp/kubeconfig.yaml
        chmod 600 /tmp/kubeconfig.yaml

    - name: Upload kubeconfig to S3 (first control plane only)
      shell: |
        # Only upload if this is the first control plane node
        if ! aws s3 ls s3://{{ lookup('env', 'S3_BUCKET') }}/kubeconfig.yaml --region {{ lookup('env', 'REGION') }} 2>/dev/null; then
          aws s3 cp /tmp/kubeconfig.yaml s3://{{ lookup('env', 'S3_BUCKET') }}/kubeconfig.yaml --region {{ lookup('env', 'REGION') }}
        fi
      environment:
        AWS_DEFAULT_REGION: "{{ lookup('env', 'REGION') }}"
      ignore_errors: yes

    - name: Enable and start k3s
      systemd:
        name: k3s
        enabled: yes
        state: started

    - name: Create cluster-info ConfigMap for workers
      shell: |
        k3s kubectl create configmap cluster-info \
          --from-literal=nlb-dns={{ nlb_dns_name }} \
          --from-literal=token={{ k3s_token }} \
          --namespace=kube-system \
          --dry-run=client -o yaml | k3s kubectl apply -f -
      ignore_errors: yes
ANSIBLE_EOF

# Run Ansible playbook
cd /tmp
export K3S_TOKEN="$K3S_TOKEN"
export NLB_DNS_NAME="$NLB_DNS_NAME"
export PRIVATE_IP="$PRIVATE_IP"
export POD_NETWORK_CIDR="$POD_NETWORK_CIDR"
export CLUSTER_NAME="$CLUSTER_NAME"
export S3_BUCKET="$S3_BUCKET"
export REGION="$REGION"

ansible-playbook control-plane-playbook.yml -vv

# Cleanup sensitive data
rm -f /tmp/k3s-token
rm -f /tmp/control-plane-playbook.yml
rm -f /tmp/kubeconfig.yaml

echo "Control plane node setup complete!"
