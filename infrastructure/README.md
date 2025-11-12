# AWS k3s Cluster with Spot Instances - Infrastructure as Code

This Terraform configuration deploys a cost-optimized, high-availability k3s Kubernetes cluster on AWS using spot instances.

## Architecture Overview

- **Region**: ca-central-1 (Canada Central)
- **Control Plane**: 3 spot instances (t3a.small, t3.small, or t2.small) behind a Network Load Balancer
- **Workers**: Auto-scaling group of spot instances (1-5 nodes, t3a.small preferred)
- **Network**: Public subnets only (no NAT gateway for cost savings)
- **Storage**: EBS volumes with gp3 type, S3 for video storage
- **DNS**: Route53 for automatic DNS management via external-dns
- **TLS**: cert-manager with Let's Encrypt

## Cost Optimization Features

1. **100% Spot Instances**: All EC2 instances use spot pricing
2. **No NAT Gateway**: Public subnets only to avoid NAT costs
3. **Cheap Instance Types**: Prefers t3a (AMD) instances for lower cost
4. **Auto-scaling**: Workers scale based on CPU utilization (30-70%)
5. **gp3 Storage**: Cost-effective EBS volumes
6. **Multiple Instance Types**: ASG configured with multiple instance types for spot availability

## Cluster Add-ons

- **NGINX Ingress Controller**: Service type LoadBalancer (creates NLB)
- **AWS EBS CSI Driver**: For persistent volumes
- **metrics-server**: For HPA and resource metrics
- **cert-manager**: Automatic TLS certificates from Let's Encrypt
- **external-dns**: Automatic Route53 DNS record management

## Prerequisites

1. **AWS Account** with appropriate permissions
2. **AWS CLI** configured with credentials
3. **Terraform** >= 1.5
4. **kubectl** for cluster access
5. **Domain**: nolift.training (registered with Namecheap)

## Deployment Steps

### Step 1: Initialize and Plan

```bash
cd infrastructure
terraform init
terraform plan
```

### Step 2: Create the Cluster (First Apply)

Create the cluster infrastructure without Kubernetes resources:

```bash
terraform apply -var="kubernetes_resources_enabled=false"
```

This will create:
- VPC, subnets, security groups
- IAM roles and policies
- Control plane ASG with 3 spot instances
- Worker ASG with 2 spot instances (default)
- Network Load Balancer for control plane
- S3 buckets for k3s config and videos
- Route53 hosted zone

**Wait 5-10 minutes** for the control plane to initialize and upload the kubeconfig to S3.

### Step 3: Get AWS Route53 Nameservers

After the first apply, get the nameservers:

```bash
terraform output aws_nameservers
```

You'll see output like:
```
[
  "ns-1234.awsdns-12.org",
  "ns-5678.awsdns-56.com",
  "ns-9012.awsdns-90.net",
  "ns-3456.awsdns-34.co.uk"
]
```

### Step 4: Update Domain Nameservers in Namecheap

1. Log in to your **Namecheap account**
2. Go to **Domain List** and click **Manage** next to `nolift.training`
3. Under **Nameservers**, select **Custom DNS**
4. Add all 4 AWS nameservers from the terraform output
5. Click **Save**

**Wait 10-30 minutes** for DNS propagation.

### Step 5: Deploy Kubernetes Add-ons (Second Apply)

Once DNS is propagated and the control plane is ready:

```bash
terraform apply -var="kubernetes_resources_enabled=true"
```

This will:
- Download kubeconfig from S3
- Deploy NGINX Ingress Controller
- Deploy EBS CSI driver
- Deploy metrics-server
- Deploy cert-manager
- Deploy external-dns
- Create Let's Encrypt ClusterIssuer

### Step 6: Access the Cluster

The kubeconfig is automatically downloaded to `infrastructure/kubeconfig.yaml`:

```bash
export KUBECONFIG=$(pwd)/kubeconfig.yaml
kubectl get nodes
kubectl get pods -A
```

You should see 3 control plane nodes and 2+ worker nodes.

## Cost Estimation

**Monthly costs** (approximate, spot pricing varies):

- Control plane (3 x t3a.small spot): ~$10-15/month
- Workers (2 x t3a.small spot): ~$7-10/month
- EBS volumes (5 x 30GB gp3): ~$10/month
- Network Load Balancer: ~$16/month
- Data transfer: Variable
- S3 storage: <$1/month

**Total: ~$45-55/month** (vs $200+/month for managed EKS/AKS)

## Next Steps

After the cluster is running, you need to:

1. **Update Namecheap nameservers** with the AWS Route53 nameservers from terraform output
2. Deploy your applications using Ingress resources with TLS annotations
3. external-dns will automatically create DNS records in Route53
4. cert-manager will automatically request Let's Encrypt certificates

For more details, see the full documentation in this README.
