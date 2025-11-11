# Domain Setup Quick Start

## TL;DR

You want to use AWS Route 53 for domain registration but Azure DNS for management.

**Setup in 5 Steps:**

1. **Buy domain on Route 53** (~$12/year)
2. **Update terraform.tfvars**: `domain_name = "yourdomain.com"`
3. **Run terraform apply** → Get Azure nameservers
4. **Update Route 53 nameservers** → Point to Azure DNS
5. **Configure email domain** in Azure Communication Services

**Total Cost**: ~$18-21/year ($12 domain + $6 Azure DNS)

---

## 1. Purchase Domain on Route 53

```bash
# Go to AWS Console → Route 53 → Register Domain
# Search and buy your domain (e.g., coachpotato.app)
# Wait 10-15 minutes for registration
```

## 2. Update Terraform Configuration

**Create or update `infrastructure/terraform.tfvars`:**

```hcl
# Required: Your domain name
domain_name = "yourdomain.com"  # Change this!

# Existing variables (already set)
azure_subscription_id = "your-subscription-id"
project_name         = "coachpotato"
environment          = "dev"

# These will be filled in later (Step 5)
azure_email_domain_verification_code = ""  # From Azure portal
azure_email_dkim_selector1          = ""  # From Azure portal
azure_email_dkim_value1             = ""  # From Azure portal
azure_email_dkim_selector2          = ""  # From Azure portal
azure_email_dkim_value2             = ""  # From Azure portal
```

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

## 4. Point Route 53 to Azure DNS

1. Go to **AWS Console → Route 53 → Registered Domains**
2. Click your domain
3. Click **"Add or edit name servers"**
4. **Replace** the existing nameservers with the Azure nameservers from Step 3
5. **Save**
6. Wait 1-2 hours for propagation (can take up to 48 hours)

**Verify DNS propagation:**
```bash
dig NS yourdomain.com
# Should show Azure nameservers

dig app.yourdomain.com
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
# 5. Enter your domain name
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
#   https://auth.yourdomain.com/realms/powerlifting-coach/broker/google/endpoint
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
dig NS yourdomain.com

# Force DNS cache flush
sudo systemd-resolve --flush-caches  # Linux
```

### Email not sending?

```bash
# Verify TXT records exist
dig TXT yourdomain.com
dig TXT selector1._domainkey.yourdomain.com

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
| Route 53 domain | ~$12-15/year |
| Azure DNS zone | ~$0.50/month ($6/year) |
| Azure Communication Services | Free (100 emails/day) |
| cert-manager (HTTPS) | Free |
| **Total** | **~$18-21/year** |

## Your URLs After Setup

With domain `yourdomain.com`:

- **Frontend**: https://app.yourdomain.com
- **API**: https://api.yourdomain.com
- **Auth/Keycloak**: https://auth.yourdomain.com
- **Grafana**: https://grafana.yourdomain.com
- **ArgoCD**: https://argocd.yourdomain.com
- **Prometheus**: https://prometheus.yourdomain.com
- **RabbitMQ**: https://rabbitmq.yourdomain.com

## Next Steps

1. Set up monitoring for DNS/SSL
2. Configure email reputation (DMARC reporting)
3. Production hardening (DDoS protection, WAF)

---

**Questions?** See the full guide: [docs/DOMAIN_SETUP.md](docs/DOMAIN_SETUP.md)
