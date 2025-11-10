# Azure Migration Guide

This document describes the migration from Digital Ocean to Azure infrastructure.

## Overview

The infrastructure has been migrated from Digital Ocean to Microsoft Azure with the following changes:

### Infrastructure Changes

1. **Kubernetes Cluster**:
   - **Before**: Digital Ocean Kubernetes (DOKS)
   - **After**: Azure Kubernetes Service (AKS) with **Spot Instances** for cost savings
   - Spot instances use `priority = "Spot"` with automatic eviction handling

2. **Storage**:
   - **Before**: Digital Ocean Spaces (S3-compatible object storage)
   - **After**: Azure Blob Storage with lifecycle management
   - Storage account name format: `powerliftingcoachdevvideos` (no hyphens allowed)
   - Container name: `powerlifting-coach-videos`
   - Lifecycle policy: Automatic deletion after 120 days

3. **Networking**:
   - **Before**: Digital Ocean VPC
   - **After**: Azure Virtual Network (automatically managed by AKS with Azure CNI)

## Prerequisites

Before deploying to Azure, ensure you have:

1. **Azure CLI** installed and configured
2. **Azure subscription** with sufficient permissions
3. **Terraform** v1.0+ installed
4. **kubectl** installed

## Configuration Changes

### Terraform Variables

Update `infrastructure/variables.tf` or create a `terraform.tfvars` file:

```hcl
azure_subscription_id = "your-azure-subscription-id"
project_name          = "powerlifting-coach"
environment           = "dev"
region                = "eastus"  # or your preferred Azure region
node_size             = "Standard_B2s"  # Spot instance VM size
kubernetes_version    = "1.28"
storage_container_name = "powerlifting-coach-videos"
```

### Storage Configuration

Azure Blob Storage endpoints differ from Digital Ocean Spaces:

- **Endpoint format**: `https://<storage-account-name>.blob.core.windows.net`
- **Authentication**: Uses storage account name as access key and storage account key as secret
- **Container access**: Set to `blob` (public read access for blobs)

### Application Configuration

The application services use AWS SDK with S3-compatible API. While Azure Blob Storage doesn't natively support S3 API, the configuration has been updated:

**Environment variables** (set in Kubernetes manifests):
- `SPACES_ENDPOINT`: `https://powerliftingcoachdevvideos.blob.core.windows.net`
- `SPACES_BUCKET`: `powerlifting-coach-videos`
- `SPACES_REGION`: `eastus`
- `SPACES_KEY`: Storage account name (injected via Kubernetes secret)
- `SPACES_SECRET`: Storage account key (injected via Kubernetes secret)

## Deployment Steps

### 1. Initialize Terraform

```bash
cd infrastructure
terraform init
```

### 2. Plan the Deployment

```bash
terraform plan -var="azure_subscription_id=YOUR_SUBSCRIPTION_ID"
```

### 3. Apply Infrastructure

```bash
# First apply: Creates AKS cluster and storage
terraform apply -var="azure_subscription_id=YOUR_SUBSCRIPTION_ID"

# Second apply: Creates Kubernetes resources
terraform apply -var="azure_subscription_id=YOUR_SUBSCRIPTION_ID" \
                -var="kubernetes_resources_enabled=true"

# Third apply: Creates ArgoCD resources
terraform apply -var="azure_subscription_id=YOUR_SUBSCRIPTION_ID" \
                -var="kubernetes_resources_enabled=true" \
                -var="argocd_resources_enabled=true"
```

### 4. Configure kubectl

```bash
# Get AKS credentials
az aks get-credentials \
  --resource-group powerlifting-coach-dev-rg \
  --name powerlifting-coach-dev

# Or use the generated kubeconfig
export KUBECONFIG=$(terraform output -raw kubeconfig_path)
```

### 5. Verify Deployment

```bash
# Check cluster nodes (should show spot instances)
kubectl get nodes

# Check storage configuration
az storage account show \
  --name powerliftingcoachdevvideos \
  --resource-group powerlifting-coach-dev-rg

# List storage containers
az storage container list \
  --account-name powerliftingcoachdevvideos
```

## Cost Optimization with Spot Instances

The AKS cluster uses **Spot Instances** for significant cost savings:

- **Priority**: `Spot`
- **Eviction Policy**: `Delete`
- **Max Price**: `-1` (use current Azure spot price)
- **Auto-scaling**: 0-3 nodes

### Spot Instance Considerations

1. **Eviction Handling**: Spot VMs can be evicted with 30 seconds notice
2. **Workload Suitability**: Best for stateless, fault-tolerant workloads
3. **Auto-scaling**: Cluster can scale to 0 nodes when idle
4. **Cost Savings**: Up to 90% cheaper than regular VMs

## Important Notes

### S3 Compatibility

**Note**: Azure Blob Storage does not natively support the S3 API. The application code currently uses AWS SDK with S3-compatible endpoints. You may need to:

1. **Option 1**: Modify the application to use Azure Blob Storage SDK
2. **Option 2**: Deploy MinIO as an S3-compatible gateway to Azure Blob Storage
3. **Option 3**: Use Azure's experimental S3 compatibility features (when available)

For now, the configuration assumes Option 1 or testing is needed to verify compatibility.

### Storage Account Naming

Azure storage account names must:
- Be 3-24 characters long
- Contain only lowercase letters and numbers
- Be globally unique

The Terraform configuration automatically removes hyphens from the project name.

### Region Considerations

Choose an Azure region based on:
- **Proximity to users**: Lower latency
- **Spot instance availability**: Not all VM sizes available in all regions
- **Cost**: Prices vary by region

Popular regions:
- `eastus`: US East Coast
- `westus2`: US West Coast
- `centralus`: US Central
- `westeurope`: Europe

## Troubleshooting

### Spot Instance Evictions

If pods are frequently evicted:

```bash
# Check node events
kubectl get events --all-namespaces | grep -i evict

# Consider using regular instances for critical workloads
# Update main.tf: priority = "Regular"
```

### Storage Access Issues

```bash
# Verify storage account key
az storage account keys list \
  --account-name powerliftingcoachdevvideos \
  --resource-group powerlifting-coach-dev-rg

# Check CORS configuration
az storage cors list \
  --account-name powerliftingcoachdevvideos \
  --services b
```

### Network Connectivity

```bash
# Check AKS network profile
az aks show \
  --resource-group powerlifting-coach-dev-rg \
  --name powerlifting-coach-dev \
  --query networkProfile
```

## Migration Checklist

- [ ] Azure subscription configured
- [ ] Terraform variables updated
- [ ] Infrastructure deployed (`terraform apply`)
- [ ] kubectl configured
- [ ] Storage account verified
- [ ] Application pods running
- [ ] Storage upload/download tested
- [ ] DNS/Ingress configured
- [ ] Monitoring set up
- [ ] Cost alerts configured

## Rollback

To rollback to Digital Ocean:

1. Revert all changes in this migration
2. Update Terraform to use `digitalocean` provider
3. Restore Digital Ocean Spaces configuration
4. Re-deploy applications

## Cost Monitoring

Monitor your Azure costs:

```bash
# View cost analysis
az consumption usage list

# Set up budget alerts (recommended)
az consumption budget create \
  --budget-name powerlifting-coach-budget \
  --amount 100 \
  --time-grain Monthly \
  --start-date 2025-11-01 \
  --end-date 2026-11-01
```

## Support

For issues with:
- **Terraform**: Check the [Azure Provider docs](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
- **AKS**: See [Azure AKS documentation](https://docs.microsoft.com/azure/aks/)
- **Spot Instances**: Review [Azure Spot VMs guide](https://docs.microsoft.com/azure/virtual-machines/spot-vms)
- **Blob Storage**: Check [Azure Blob Storage docs](https://docs.microsoft.com/azure/storage/blobs/)
