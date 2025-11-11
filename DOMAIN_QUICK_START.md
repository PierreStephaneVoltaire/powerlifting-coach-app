# Domain Setup Quick Start

## TL;DR

Domain registered on **Namecheap** (nolift.training), DNS managed by **Azure DNS**.

**Setup in 5 Steps:**

1. ✅ **Domain purchased**: `nolift.training` (Namecheap)
2. **Update terraform.tfvars**: `domain_name = "nolift.training"` ✅ DONE
3. **Run terraform apply** → Get Azure nameservers
4. **Update Namecheap nameservers** → Point to Azure DNS
5. **Configure email domain** in Azure Communication Services

**Total Cost**: ~$36/year ($30 domain + $6 Azure DNS)

---

## 1. ✅ Domain Purchased

Domain **nolift.training** purchased from Namecheap.

## 2. ✅ Terraform Configuration Updated

File **`infrastructure/terraform.tfvars`** has been created with:

```hcl
# Domain Configuration
domain_name = "nolift.training"  ✅

# Project Configuration
project_name = "coachpotato"
environment  = "dev"

# Kubernetes Configuration
kubernetes_resources_enabled = true
argocd_resources_enabled     = true
```

**⚠️ ACTION REQUIRED**: You need to add your Azure subscription ID:
```hcl
azure_subscription_id = "your-subscription-id-here"
```

The rest (email, OAuth) can be filled in later during Step 5.

## 3. Deploy Infrastructure

```bash
cd infrastructure
terraform apply

# IMPORTANT: Copy the Azure nameservers from output!
terraform output azure_nameservers
```

**Output will look like:**
```
azure_nameservers = [
  "ns1-01.azure-dns.com.",
  "ns2-01.azure-dns.net.",
  "ns3-01.azure-dns.org.",
  "ns4-01.azure-dns.info."
]
```

**SAVE THESE NAMESERVERS!** You need them for the next step.

## 4. Point Namecheap to Azure DNS

1. **Log into Namecheap**: https://www.namecheap.com
2. Go to **Domain List** → Click **"Manage"** next to `nolift.training`
3. Scroll to **"Nameservers"** section
4. Select **"Custom DNS"** from the dropdown
5. **Enter the 4 Azure nameservers** from Step 3 (⚠️ remove trailing dots!)
   ```
   ns1-01.azure-dns.com
   ns2-01.azure-dns.net
   ns3-01.azure-dns.org
   ns4-01.azure-dns.info
   ```
6. Click the **checkmark** to save
7. Wait 30 minutes to 2 hours for propagation (Namecheap is usually fast!)

**Verify DNS propagation:**
```bash
dig NS nolift.training
# Should show Azure nameservers

dig app.nolift.training
# Should show your LoadBalancer IP
```

## 5. Configure Email Domain (Azure Communication Services)

### 5a. Create/Configure Azure Communication Services

```bash
# In Azure Portal:
# 1. Search "Communication Services"
# 2. Create resource (if not exists)
# 3. Select same resource group as AKS cluster
# 4. Go to Email → Domains → Add Domain → Custom Domain
# 5. Enter: nolift.training
```

### 5b. Get Verification Records

Azure will show you:
- Domain verification code (TXT record)
- DKIM selector 1 (TXT record)
- DKIM selector 2 (TXT record)

### 5c. Update Terraform with Email Records

**Add to `infrastructure/terraform.tfvars`:**

```hcl
# Copy these values from Azure portal
azure_email_domain_verification_code = "MS=ms12345678"

azure_email_dkim_selector1 = "selector1._domainkey"
azure_email_dkim_value1    = "v=DKIM1; k=rsa; p=MIGfMA0GCS..."

azure_email_dkim_selector2 = "selector2._domainkey"
azure_email_dkim_value2    = "v=DKIM1; k=rsa; p=MIIBIjANB..."
```

### 5d. Apply Terraform

```bash
terraform apply
```

This creates:
- TXT record for domain verification
- 2 DKIM TXT records
- SPF TXT record
- DMARC TXT record

### 5e. Verify in Azure

1. Go back to Azure Communication Services
2. Click **"Verify"** on your custom domain
3. Wait a few minutes
4. Should show status: **"Verified"**

## 6. Update Google OAuth (if using)

```bash
# Go to: https://console.cloud.google.com/
# Navigate to: APIs & Services → Credentials
# Select your OAuth 2.0 Client ID
# Add redirect URI:
#   https://auth.nolift.training/realms/powerlifting-coach/broker/google/endpoint
# Save
```

## 7. Enable HTTPS (Recommended)

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create Let's Encrypt issuer (update email!)
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com  # CHANGE THIS!
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

Then update your ingress resources to use TLS (see full docs/DOMAIN_SETUP.md for details).

---

## Troubleshooting

### DNS not resolving?

```bash
# Check nameservers
dig NS nolift.training

# Force DNS cache flush
sudo systemd-resolve --flush-caches  # Linux
```

### Email not sending?

```bash
# Verify TXT records exist
dig TXT nolift.training
dig TXT selector1._domainkey.nolift.training

# Check Keycloak logs
kubectl logs -n app deployment/keycloak | grep -i email
```

### Need to roll back to nip.io?

```bash
# Update terraform.tfvars
domain_name = "localhost"

# Apply
terraform apply
```

---

## Full Documentation

For detailed instructions, troubleshooting, and architecture details, see:
**[docs/DOMAIN_SETUP.md](docs/DOMAIN_SETUP.md)**

## Cost Summary

| Item | Cost |
|------|------|
| Namecheap .training domain | ~$30/year |
| Azure DNS zone | ~$0.50/month ($6/year) |
| Azure Communication Services | Free (100 emails/day) |
| cert-manager (HTTPS) | Free |
| **Total** | **~$36/year** |

Note: `.training` TLD is pricier than `.com`, but perfect for fitness/coaching apps!

## Your URLs After Setup

With domain `nolift.training`:

- **Frontend**: https://app.nolift.training
- **API**: https://api.nolift.training
- **Auth/Keycloak**: https://auth.nolift.training
- **Grafana**: https://grafana.nolift.training
- **ArgoCD**: https://argocd.nolift.training
- **Prometheus**: https://prometheus.nolift.training
- **RabbitMQ**: https://rabbitmq.nolift.training
- **OpenWebUI**: https://openwebui.nolift.training

## Next Steps

1. Set up monitoring for DNS/SSL
2. Configure email reputation (DMARC reporting)
3. Production hardening (DDoS protection, WAF)

---

**Questions?** See the full guide: [docs/DOMAIN_SETUP.md](docs/DOMAIN_SETUP.md)
