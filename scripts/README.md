# Infrastructure Scripts

Management scripts for the Nolift powerlifting app infrastructure.

## Prerequisites

1. **Azure CLI**: `az login`
2. **Terraform**: v1.0+
3. **kubectl**: For Kubernetes management
4. **argocd CLI**: For deploy-app.sh (optional)

## Setup

1. Configure `infrastructure/terraform.tfvars` with your values:
   ```hcl
   azure_subscription_id = "your-subscription-id"
   domain_name = "nolift.training"
   openai_api_key = "your-key"
   google_oauth_client_id = "your-client-id"
   google_oauth_client_secret = "your-secret"
   ```

2. Login to Azure:
   ```bash
   az login
   ```

## Scripts

### `./init-infra.sh`
**Stages 1 & 2: Initialize base infrastructure and Kubernetes components**

Creates:
- **Stage 1**: AKS cluster, DNS zone, storage, email services, DNS records
- **Stage 2**: nginx-ingress, cert-manager, ArgoCD helm chart, namespaces, secrets, monitoring

Does NOT enable ArgoCD Applications yet (need ArgoCD to be fully running first).

Cost: ~$50-80/month

### `./enable-argo.sh`
**Stage 3: Enable ArgoCD Applications for continuous deployment**

Creates ArgoCD Application CRDs that define what to deploy.

**Requirements**:
- Must run `init-infra.sh` first
- ArgoCD helm chart must be fully deployed (usually takes 2-3 minutes)
- Sets up auto-deployment from GitHub

### `./stop-workload.sh`
**Stop workloads to save costs**

- Destroys all helm releases
- Scales spot nodes to 0
- Deletes LoadBalancer
- Keeps 1 node + infrastructure

Cost when stopped: ~$10/month

### `./start-workload.sh`
**Restart workloads after stopping**

Restores everything back to running state.

### `./deploy-app.sh`
**Manually trigger deployment**

Forces ArgoCD to sync and deploy latest code.
Normally happens automatically on git push.

### `./destroy-infra.sh`
**Destroy ALL infrastructure**

⚠️ WARNING: Deletes everything including data!
Requires typing 'destroy' to confirm.

## Typical Workflows

### First time setup:
```bash
# Stages 1 & 2: Create infrastructure + K8s components
./init-infra.sh

# Stage 3: Enable ArgoCD Applications (wait 2-3 min for ArgoCD to be ready)
./enable-argo.sh

# Configure nameservers in domain registrar
terraform -chdir=../infrastructure output azure_nameservers

# Wait for DNS propagation (5-60 minutes)

# After domain is verified in Azure Portal, link it:
terraform -chdir=../infrastructure apply \
  -var='kubernetes_resources_enabled=true' \
  -var='argocd_resources_enabled=true' \
  -var='email_domain_verified=true'
```

### Daily development:
- Push to GitHub → auto-deploys via ArgoCD
- Monitor: https://argocd.nolift.training

### Save costs overnight/weekend:
```bash
./stop-workload.sh   # Saves ~$40-70/month
# Later...
./start-workload.sh  # Restore everything
```

### Complete teardown:
```bash
./destroy-infra.sh
```

## Cost Breakdown

**Running (all services active):**
- Default node pool: ~$10/month (Standard_B2s)
- Spot node pool: ~$2-5/month (Standard_B2ms at 70-90% discount)
- LoadBalancer: ~$20-30/month
- Storage: ~$5-10/month
- Communication Services: Free tier
- **Total: ~$50-80/month**

**Stopped:**
- Default node pool: ~$10/month
- Everything else: $0
- **Total: ~$10/month**

## Troubleshooting

**DNS not resolving:**
- Check nameservers configured in domain registrar
- Wait up to 48 hours for propagation
- Verify: `dig nolift.training NS`

**SSL certificates failing:**
- Ensure DNS is fully propagated first
- Check cert-manager logs: `kubectl --kubeconfig=infrastructure/kubeconfig.yaml logs -n cert-manager -l app=cert-manager`

**Email domain verification failing:**
- Ensure all DNS records are created (TXT @, DKIM CNAMEs, DMARC)
- Wait for DNS propagation
- Check Azure Portal: Communication Services → Email Services
- Only set `email_domain_verified=true` after Azure shows verified

**ArgoCD not syncing:**
- Check ArgoCD logs in the UI
- Verify GitHub repo is accessible
- Try manual sync: `./deploy-app.sh`
