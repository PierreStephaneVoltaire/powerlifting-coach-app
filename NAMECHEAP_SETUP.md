# Namecheap + Azure DNS Setup for nolift.training

Quick guide for setting up your Namecheap domain (nolift.training) with Azure DNS.

## What You Have

- **Domain**: `nolift.training` (registered on Namecheap)
- **DNS Management**: Azure DNS (keeping DNS close to your infrastructure)
- **Infrastructure**: Azure AKS

## Setup Steps

### Step 1: Apply Terraform to Create Azure DNS Zone

```bash
cd infrastructure

# Make sure you have your Azure subscription ID
# Update terraform.tfvars if needed:
# azure_subscription_id = "your-subscription-id"

# Initialize and apply
terraform init
terraform apply
```

**What this creates:**
- Azure DNS zone for `nolift.training`
- A records for: `app`, `api`, `auth` subdomains
- Email DNS records (TXT, SPF, DKIM, DMARC) when configured

### Step 2: Get Azure Nameservers

After terraform apply completes:

```bash
terraform output azure_nameservers
```

You'll see 4 nameservers like:
```
[
  "ns1-01.azure-dns.com.",
  "ns2-01.azure-dns.net.",
  "ns3-01.azure-dns.org.",
  "ns4-01.azure-dns.info."
]
```

**Copy these nameservers** - you'll need them in the next step!

### Step 3: Update Nameservers in Namecheap

1. **Log into Namecheap**: https://www.namecheap.com/myaccount/login/
2. **Go to Domain List**: Dashboard → Domain List
3. **Select your domain**: Click "Manage" next to `nolift.training`
4. **Scroll to Nameservers section**
5. **Select "Custom DNS"** from the dropdown
6. **Enter the Azure nameservers** (remove the trailing dots):
   ```
   ns1-01.azure-dns.com
   ns2-01.azure-dns.net
   ns3-01.azure-dns.org
   ns4-01.azure-dns.info
   ```
7. **Click the checkmark** to save

**Important**: Remove the trailing dots when entering in Namecheap!

### Step 4: Wait for DNS Propagation

- **Typical time**: 30 minutes to 2 hours
- **Maximum time**: 24-48 hours
- Namecheap usually propagates quickly (faster than Route 53)

**Check propagation:**
```bash
# Check nameservers
dig NS nolift.training

# Check A records (after propagation)
dig app.nolift.training
dig api.nolift.training
dig auth.nolift.training
```

### Step 5: Configure Azure Communication Services Email

#### 5a. Create Communication Services Resource

```bash
# In Azure Portal:
# 1. Search "Communication Services"
# 2. Create new resource
# 3. Use same resource group as your AKS cluster
# 4. Region: East US (or your cluster region)
# 5. Create
```

#### 5b. Add Custom Email Domain

1. In Communication Services resource → **Email** → **Domains**
2. Click **"Add domain"** → **"Custom domain"**
3. Enter: `nolift.training`
4. Click **"Add"**

#### 5c. Get Verification Records

Azure will show you verification records:
- **Domain verification code** (TXT record starting with `MS=`)
- **DKIM selector 1** and value
- **DKIM selector 2** and value

#### 5d. Update terraform.tfvars

Add these lines to `infrastructure/terraform.tfvars`:

```hcl
# Email Domain Verification (from Azure portal)
azure_email_domain_verification_code = "MS=ms12345678"  # Your actual code

azure_email_dkim_selector1 = "selector1._domainkey"
azure_email_dkim_value1    = "v=DKIM1; k=rsa; p=MIGfMA0GCS..."  # Your actual value

azure_email_dkim_selector2 = "selector2._domainkey"
azure_email_dkim_value2    = "v=DKIM1; k=rsa; p=MIIBIjANB..."  # Your actual value

# Email SMTP Configuration (from Azure portal → Email → Domains → Provision)
azure_email_smtp_username = "your-smtp-username"
azure_email_smtp_password = "your-smtp-password-or-connection-string"
azure_email_from_email    = "noreply@nolift.training"
```

#### 5e. Apply Terraform Again

```bash
cd infrastructure
terraform apply
```

This will create all the email DNS records (TXT, DKIM, SPF, DMARC).

#### 5f. Verify Domain in Azure

1. Go back to Azure Communication Services → Email → Domains
2. Find `nolift.training`
3. Click **"Verify"**
4. Wait 5-10 minutes for verification to complete
5. Status should change to **"Verified"** ✓

### Step 6: Update Google OAuth (if using)

If you're using Google social login:

1. Go to **Google Cloud Console**: https://console.cloud.google.com/
2. Navigate to: **APIs & Services** → **Credentials**
3. Select your **OAuth 2.0 Client ID**
4. Under **"Authorized redirect URIs"**, add:
   ```
   https://auth.nolift.training/realms/powerlifting-coach/broker/google/endpoint
   ```
5. **Save**

### Step 7: Enable HTTPS with Let's Encrypt

#### Install cert-manager

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

#### Create Let's Encrypt Issuer

```bash
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

#### Update Ingress for TLS

Update your ingress resources (in `k8s/base/ingress.yaml`) to add TLS:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-ingress
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - app.nolift.training
    - api.nolift.training
    - auth.nolift.training
    - grafana.nolift.training
    - argocd.nolift.training
    secretName: nolift-training-tls
  rules:
  - host: app.nolift.training
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
  # ... add other rules for api, auth, etc.
```

Apply the changes:
```bash
kubectl apply -f k8s/base/ingress.yaml
```

#### Update Keycloak SSL Settings

Update Keycloak realm config to require external SSL:

In `k8s/base/keycloak-realm.yaml`, change:
```json
"sslRequired": "external"
```

Apply:
```bash
kubectl apply -f k8s/base/keycloak-realm.yaml
kubectl rollout restart deployment/keycloak -n app
```

## Your New URLs

After setup completes:

| Service | URL |
|---------|-----|
| Frontend | https://app.nolift.training |
| API | https://api.nolift.training |
| Auth/Keycloak | https://auth.nolift.training |
| Grafana | https://grafana.nolift.training |
| ArgoCD | https://argocd.nolift.training |
| Prometheus | https://prometheus.nolift.training |
| RabbitMQ | https://rabbitmq.nolift.training |
| OpenWebUI | https://openwebui.nolift.training |

## Cost Breakdown

| Item | Cost |
|------|------|
| Namecheap domain (.training TLD) | ~$30/year |
| Azure DNS zone | ~$0.50/month ($6/year) |
| Azure Communication Services | Free (100 emails/day) |
| cert-manager (HTTPS) | Free |
| **Total** | **~$36/year** |

Note: `.training` domains are more expensive than `.com` (~$12/year) but are great for fitness/coaching apps!

## Troubleshooting

### Nameservers Not Updating

```bash
# Check current nameservers
dig NS nolift.training

# If still showing Namecheap nameservers:
# 1. Double-check you saved the changes in Namecheap
# 2. Wait 30 minutes and try again
# 3. Try clearing DNS cache:
sudo systemd-resolve --flush-caches  # Linux
```

### DNS Records Not Resolving

```bash
# First, verify nameservers changed
dig NS nolift.training
# Should show Azure nameservers

# Then check A records
dig app.nolift.training
# Should show your LoadBalancer IP

# If NS records correct but A records don't resolve:
# - Wait a bit longer (DNS propagation in progress)
# - Verify terraform apply succeeded
# - Check terraform output: terraform output domain_urls
```

### Email Domain Verification Fails

1. **Check DNS records exist:**
   ```bash
   dig TXT nolift.training
   dig TXT selector1._domainkey.nolift.training
   dig TXT selector2._domainkey.nolift.training
   ```

2. **Verify values match Azure portal**
   - The TXT values must match exactly
   - Check for typos in terraform.tfvars

3. **Wait and retry**
   - DNS changes can take 10-15 minutes
   - Click "Verify" button again in Azure portal

### SSL Certificates Not Issuing

```bash
# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager

# Check certificate status
kubectl get certificate -A
kubectl describe certificate nolift-training-tls -n app

# Check ACME challenge
kubectl get challenges -A
# If stuck, delete and retry:
kubectl delete certificate nolift-training-tls -n app
kubectl delete ingress app-ingress -n app
kubectl apply -f k8s/base/ingress.yaml
```

### Still Using nip.io URLs?

If your app still shows nip.io URLs:

1. **Check ingress configuration**:
   ```bash
   kubectl get ingress -A
   ```

2. **Update Keycloak redirect URIs**:
   - Log into Keycloak admin console
   - Go to Clients → powerlifting-coach-app
   - Update Valid Redirect URIs to: `https://app.nolift.training/*`
   - Update Web Origins to: `https://app.nolift.training`

3. **Update environment variables** (if your services use them):
   - Check ConfigMaps and Secrets for hardcoded nip.io URLs
   - Update to use `nolift.training`

## Namecheap-Specific Tips

1. **Nameserver Changes**: Namecheap typically propagates DNS changes within 30 minutes (faster than most registrars)

2. **WHOIS Privacy**: Namecheap includes free WHOIS protection - make sure it's enabled:
   - Domain List → Manage → WHOIS Guard → ON

3. **Auto-Renew**: Enable auto-renew to avoid losing your domain:
   - Domain List → Manage → Auto-Renew → Enable

4. **Domain Lock**: Enable registrar lock for security:
   - Domain List → Manage → Registrar Lock → Enabled

5. **Email Forwarding**: Namecheap offers free email forwarding if you want to receive emails at `you@nolift.training`:
   - Domain List → Manage → Mail Settings → Email Forwarding

## Next Steps

1. **Set up monitoring**:
   - Monitor DNS resolution
   - SSL certificate expiration alerts
   - Email deliverability

2. **Configure email reputation**:
   - Set up DMARC reporting
   - Monitor spam reports
   - Start with low email volume, increase gradually

3. **Production hardening**:
   - Enable Azure DDoS protection
   - Configure rate limiting
   - Set up web application firewall (WAF)

4. **Backup strategy**:
   - Export DNS zone regularly
   - Backup Keycloak configuration
   - Document infrastructure

## Reference

- **Namecheap DNS**: https://www.namecheap.com/support/knowledgebase/article.aspx/767/10/how-to-change-dns-for-a-domain/
- **Azure DNS**: https://docs.microsoft.com/en-us/azure/dns/
- **cert-manager**: https://cert-manager.io/docs/
- **Keycloak**: https://www.keycloak.org/documentation

## Need Help?

- Check main docs: [docs/DOMAIN_SETUP.md](docs/DOMAIN_SETUP.md)
- Review Terraform logs: `terraform apply 2>&1 | tee terraform.log`
- Check Kubernetes events: `kubectl get events -A --sort-by='.lastTimestamp'`
- View service logs: `kubectl logs -n app deployment/keycloak --tail=100`
