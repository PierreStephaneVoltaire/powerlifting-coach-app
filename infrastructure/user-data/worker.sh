#!/bin/bash
set -e

# k3s Worker Node Bootstrap Script
# This script installs Ansible and runs the worker playbook

# Variables from Terraform
CLUSTER_NAME="${cluster_name}"
S3_BUCKET="${s3_bucket}"
NLB_DNS_NAME="${nlb_dns_name}"
REGION="${region}"
MAX_PODS="${max_pods}"

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

# Get k3s token from S3 (with retries for control plane initialization)
echo "Waiting for k3s token..."
for i in {1..30}; do
    if aws s3 cp s3://$${S3_BUCKET}/k3s-token /tmp/k3s-token --region $${REGION}; then
        break
    fi
    echo "Waiting for control plane to initialize... ($i/30)"
    sleep 10
done

K3S_TOKEN=$(cat /tmp/k3s-token)

# Get instance private IP
PRIVATE_IP=$(curl -s http://169.254.169.254/latest/meta-data/local-ipv4)

# Create Ansible playbook for worker setup
cat > /tmp/worker-playbook.yml <<'ANSIBLE_EOF'
---
- name: Setup k3s worker node
  hosts: localhost
  become: yes
  vars:
    k3s_version: "v1.28.5+k3s1"
    k3s_token: "{{ lookup('env', 'K3S_TOKEN') }}"
    nlb_dns_name: "{{ lookup('env', 'NLB_DNS_NAME') }}"
    private_ip: "{{ lookup('env', 'PRIVATE_IP') }}"
    max_pods: "{{ lookup('env', 'MAX_PODS') }}"

  tasks:
    - name: Check if k3s is already installed
      stat:
        path: /usr/local/bin/k3s-agent
      register: k3s_agent_binary

    - name: Wait for control plane to be ready
      shell: |
        curl -k https://{{ nlb_dns_name }}:6443/ping
      register: control_plane_ready
      retries: 60
      delay: 10
      until: control_plane_ready.rc == 0
      changed_when: false
      when: not k3s_agent_binary.stat.exists

    - name: Install k3s agent
      shell: |
        curl -sfL https://get.k3s.io | sh -s - agent \
          --token="{{ k3s_token }}" \
          --server="https://{{ nlb_dns_name }}:6443" \
          --node-ip="{{ private_ip }}" \
          --kubelet-arg="max-pods={{ max_pods }}" \
          --node-label="node-role.kubernetes.io/worker=true"
      environment:
        INSTALL_K3S_VERSION: "{{ k3s_version }}"
      when: not k3s_agent_binary.stat.exists

    - name: Wait for k3s-agent to be ready
      wait_for:
        path: /var/lib/rancher/k3s/agent/kubelet.kubeconfig
        state: present
        timeout: 300

    - name: Enable and start k3s-agent
      systemd:
        name: k3s-agent
        enabled: yes
        state: started

    - name: Verify node registration (may take time)
      shell: |
        # Give some time for node to register
        sleep 30
      changed_when: false
ANSIBLE_EOF

# Run Ansible playbook
cd /tmp
export K3S_TOKEN="$K3S_TOKEN"
export NLB_DNS_NAME="$NLB_DNS_NAME"
export PRIVATE_IP="$PRIVATE_IP"
export MAX_PODS="$MAX_PODS"

ansible-playbook worker-playbook.yml -vv

# Cleanup sensitive data
rm -f /tmp/k3s-token
rm -f /tmp/worker-playbook.yml

echo "Worker node setup complete!"
