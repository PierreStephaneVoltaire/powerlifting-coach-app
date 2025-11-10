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
   - Storage account name format: `lazyliftsdevvideos` (no hyphens allowed)
   - Container name: `lazylifts-videos`
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
project_name          = "lazylifts"
environment           = "dev"
region                = "eastus"  # or your preferred Azure region
node_size             = "Standard_B2s"  # Spot instance VM size
kubernetes_version    = "1.28"
storage_container_name = "lazylifts-videos"
```

### Storage Configuration

Azure Blob Storage endpoints differ from Digital Ocean Spaces:

- **Endpoint format**: `https://<storage-account-name>.blob.core.windows.net`
- **Authentication**: Uses storage account name as access key and storage account key as secret
- **Container access**: Set to `blob` (public read access for blobs)

### Application Configuration

The application services have been **fully migrated** to use the Azure Blob Storage SDK natively instead of the AWS S3 SDK.

**Code changes**:
- Replaced AWS SDK (`github.com/aws/aws-sdk-go`) with Azure SDK (`github.com/Azure/azure-sdk-for-go/sdk/storage/azblob`)
- Updated `services/video-service/internal/storage/spaces.go` to use Azure Blob API
- Updated `services/media-processor-service/internal/storage/spaces.go` to use Azure Blob API
- Updated both services' `go.mod` files with Azure SDK dependencies
- All storage operations now use native Azure APIs (upload, download, SAS URLs, etc.)

**Environment variables** (set in Kubernetes manifests):
- `SPACES_ENDPOINT`: `https://lazyliftsdevvideos.blob.core.windows.net`
- `SPACES_BUCKET`: `lazylifts-videos` (container name)
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
  --resource-group lazylifts-dev-rg \
  --name lazylifts-dev

# Or use the generated kubeconfig
export KUBECONFIG=$(terraform output -raw kubeconfig_path)
```

### 5. Verify Deployment

```bash
# Check cluster nodes (should show spot instances)
kubectl get nodes

# Check storage configuration
az storage account show \
  --name lazyliftsdevvideos \
  --resource-group lazylifts-dev-rg

# List storage containers
az storage container list \
  --account-name lazyliftsdevvideos
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

### Azure SDK Integration

✅ **Fully Migrated**: The application code has been completely updated to use the Azure Blob Storage SDK natively. No S3 compatibility layer is needed.

**Key features implemented**:
- Native Azure Blob uploads and downloads
- SAS (Shared Access Signature) URL generation for secure file access
- Public and private blob support
- Azure Hot tier storage for frequently accessed content
- Full compatibility with Azure Blob Storage lifecycle policies

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
  --account-name lazyliftsdevvideos \
  --resource-group lazylifts-dev-rg

# Check CORS configuration
az storage cors list \
  --account-name lazyliftsdevvideos \
  --services b
```

### Network Connectivity

```bash
# Check AKS network profile
az aks show \
  --resource-group lazylifts-dev-rg \
  --name lazylifts-dev \
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
  --budget-name lazylifts-budget \
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
