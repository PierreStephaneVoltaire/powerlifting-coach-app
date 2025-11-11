# Domain Setup Guide: AWS Route 53 + Azure DNS

This guide walks you through setting up a custom domain registered on AWS Route 53 with DNS managed by Azure DNS.

## Architecture Overview

```
┌─────────────────────┐
│  AWS Route 53       │
│  (Domain Registry)  │
│                     │
│  - Domain Purchase  │
│  - Nameserver Only  │
└──────────┬──────────┘
           │
           │ NS Records Point To
           │
           ▼
┌─────────────────────┐
│  Azure DNS          │
│  (DNS Management)   │
│                     │
│  - A Records        │
│  - TXT Records      │
│  - MX Records       │
│  - DKIM Records     │
└──────────┬──────────┘
           │
           │ A Records Point To
           │
           ▼
┌─────────────────────┐
│  AKS Cluster        │
│  (Your App)         │
│                     │
│  - LoadBalancer IP  │
└─────────────────────┘
```

## Why This Setup?

1. **Domain Registration on AWS Route 53**: Centralized domain management if you already have domains there
2. **DNS Management on Azure**: Keep DNS in the same cloud as your infrastructure
3. **Cost**: ~$12/year (domain) + ~$0.50/month (Azure DNS zone)

## Step-by-Step Setup

### Phase 1: Purchase Domain on Route 53

1. **Log into AWS Console**
   - Navigate to Route 53
   - Go to "Registered domains" > "Register domain"

2. **Choose and Purchase Domain**
   - Search for available domain (e.g., `coachpotato.app`)
   - Complete purchase (~$12-15/year)
   - Auto-renew enabled by default (recommended)

3. **Wait for Registration**
   - Usually takes 10-15 minutes
   - You'll receive email confirmation

### Phase 2: Deploy Infrastructure with Domain

1. **Update Terraform Variables**

   Create or update `infrastructure/terraform.tfvars`:

   ```hcl
   # Domain Configuration
   domain_name = "yourdomain.com"  # Replace with your actual domain

   # Other existing variables...
   azure_subscription_id = "your-subscription-id"
   project_name         = "coachpotato"
   environment          = "dev"
   ```

2. **Run Terraform Apply**

   ```bash
   cd infrastructure
   terraform init
   terraform apply
   ```

3. **Get Azure Nameservers**

   After apply completes, get the Azure nameservers:

   ```bash
   terraform output azure_nameservers
   ```

   You should see output like:
   ```
   [
     "ns1-01.azure-dns.com.",
     "ns2-01.azure-dns.net.",
     "ns3-01.azure-dns.org.",
     "ns4-01.azure-dns.info."
   ]
   ```

   **IMPORTANT**: Copy these nameservers - you'll need them in the next step!

### Phase 3: Point Route 53 to Azure DNS

1. **Update Route 53 Nameservers**
   - Go back to AWS Route 53 Console
   - Click on "Registered domains"
   - Select your domain
   - Click "Add or edit name servers"
   - Replace the existing nameservers with the Azure nameservers from Step 2.3
   - Save changes

2. **Wait for DNS Propagation**
   - DNS changes can take 24-48 hours to fully propagate
   - Usually propagates within 1-2 hours

3. **Verify DNS Propagation**

   ```bash
   # Check nameservers
   dig NS yourdomain.com

   # Check A records (after propagation)
   dig app.yourdomain.com
   dig api.yourdomain.com
   dig auth.yourdomain.com
   ```

   You should see the Azure nameservers and LoadBalancer IP addresses.

### Phase 4: Configure Azure Communication Services Email

1. **Create Azure Communication Services Resource** (if not already created)

   ```bash
   # Via Azure Portal
   - Search for "Communication Services"
   - Create new resource
   - Select your resource group (same as AKS cluster)
   - Choose pricing tier (Free tier available)
   ```

2. **Add Email Domain**
   - In your Communication Services resource, go to "Email" > "Domains"
   - Click "Add domain"
   - Choose "Custom domain"
   - Enter your domain name (e.g., `coachpotato.app`)

3. **Get Verification Records**

   Azure will provide verification records. You'll see:
   - TXT record for domain verification
   - DKIM records (2 selectors)
   - Recommended SPF/DMARC records

4. **Update Terraform with Email Records**

   Add to `infrastructure/terraform.tfvars`:

   ```hcl
   # Email Domain Verification
   azure_email_domain_verification_code = "MS=ms12345678"  # From Azure portal

   # DKIM Configuration (from Azure portal)
   azure_email_dkim_selector1 = "selector1._domainkey"
   azure_email_dkim_value1    = "v=DKIM1; k=rsa; p=MIGfMA0GCS..."
   azure_email_dkim_selector2 = "selector2._domainkey"
   azure_email_dkim_value2    = "v=DKIM1; k=rsa; p=MIIBIjANB..."

   # Email Configuration (already set)
   azure_email_smtp_host     = "smtp.azurecomm.net"
   azure_email_smtp_username = "your-username"
   azure_email_smtp_password = "your-password"
   azure_email_from_email    = "noreply@yourdomain.com"
   ```

5. **Apply Terraform to Add DNS Records**

   ```bash
   cd infrastructure
   terraform apply
   ```

   This will create:
   - TXT record for domain verification
   - DKIM TXT records (2)
   - SPF TXT record
   - DMARC TXT record

6. **Verify Domain in Azure**
   - Go back to Azure Communication Services
   - Click "Verify" on your domain
   - Should show as "Verified" (may take a few minutes)

7. **Test Email Sending**

   ```bash
   # Port-forward to Keycloak to trigger password reset email
   kubectl port-forward -n app svc/keycloak-service 8080:8080

   # In browser, go to: http://localhost:8080
   # Try "Forgot Password" flow to test email
   ```

### Phase 5: Update OAuth Redirect URIs

1. **Update Google OAuth Console**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Navigate to "APIs & Services" > "Credentials"
   - Select your OAuth 2.0 Client ID
   - Add authorized redirect URIs:
     - `https://auth.yourdomain.com/realms/powerlifting-coach/broker/google/endpoint`
   - Save

2. **Update Keycloak Configuration** (if needed)
   - Most redirect URIs should be automatically updated via Terraform
   - Verify in Keycloak Admin Console (http://auth.yourdomain.com)

### Phase 6: Enable HTTPS (Recommended)

Currently, the setup uses HTTP with nip.io. For production with a custom domain, enable HTTPS:

1. **Install cert-manager** (if not already installed)

   ```bash
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
   ```

2. **Create Let's Encrypt ClusterIssuer**

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: cert-manager.io/v1
   kind: ClusterIssuer
   metadata:
     name: letsencrypt-prod
   spec:
     acme:
       server: https://acme-v02.api.letsencrypt.org/directory
       email: your-email@example.com
       privateKeySecretRef:
         name: letsencrypt-prod
       solvers:
       - http01:
           ingress:
             class: nginx
   EOF
   ```

3. **Update Ingress Resources**

   Add TLS configuration to your ingress resources in `k8s/base/ingress.yaml`:

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
       - app.yourdomain.com
       - api.yourdomain.com
       - auth.yourdomain.com
       secretName: app-tls-cert
     rules:
     - host: app.yourdomain.com
       # ... rest of ingress config
   ```

4. **Update Keycloak SSL Settings**

   In `k8s/base/keycloak-realm.yaml`, update:
   ```json
   "sslRequired": "external"
   ```

## Cost Breakdown

| Service | Cost | Notes |
|---------|------|-------|
| Route 53 Domain Registration | ~$12-15/year | Varies by TLD (.com, .app, etc.) |
| Route 53 Hosted Zone | $0/month | Deleted after moving to Azure DNS |
| Azure DNS Zone | ~$0.50/month | First 25 zones, 1M queries |
| Azure Communication Services | Free | First 100 emails/day free |
| cert-manager (HTTPS) | Free | Open source |

**Total**: ~$12-15/year + $6/year = ~$18-21/year

## Troubleshooting

### DNS Not Resolving

```bash
# Check if nameservers updated
dig NS yourdomain.com

# Check if A records created
dig app.yourdomain.com

# Force DNS cache flush
sudo systemd-resolve --flush-caches  # Linux
dscacheutil -flushcache              # macOS
ipconfig /flushdns                   # Windows
```

### Email Not Sending

1. **Check Domain Verification**
   ```bash
   # Verify TXT record exists
   dig TXT yourdomain.com

   # Check DKIM records
   dig TXT selector1._domainkey.yourdomain.com
   dig TXT selector2._domainkey.yourdomain.com
   ```

2. **Check Keycloak Email Settings**
   ```bash
   kubectl logs -n app deployment/keycloak | grep -i email
   ```

3. **Verify SMTP Credentials**
   ```bash
   kubectl get secret -n app azure-email-secret -o yaml
   ```

### SSL Certificate Issues

```bash
# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager

# Check certificate status
kubectl get certificate -A
kubectl describe certificate app-tls-cert -n app

# Check challenge status
kubectl get challenges -A
```

### Google OAuth Not Working

1. **Verify Redirect URI Matches Exactly**
   - Google Console: `https://auth.yourdomain.com/realms/powerlifting-coach/broker/google/endpoint`
   - Keycloak: Check "Identity Providers" > "Google" > "Redirect URI"

2. **Check CORS Settings**
   - In Keycloak Admin Console
   - Client settings > "Web Origins"
   - Should include `https://app.yourdomain.com`

## Switching Back to nip.io (Rollback)

If you need to roll back:

1. **Update terraform.tfvars**
   ```hcl
   domain_name = "localhost"
   ```

2. **Apply Terraform**
   ```bash
   terraform apply
   ```

3. **Update Google OAuth Redirect URIs**
   - Change back to: `http://auth.{your-lb-ip}.nip.io/realms/powerlifting-coach/broker/google/endpoint`

## Next Steps

1. Set up monitoring for your domain:
   - DNS monitoring (uptime checks)
   - SSL certificate expiration alerts
   - Email deliverability monitoring

2. Configure email reputation:
   - Set up DMARC reporting
   - Monitor spam reports
   - Gradually increase sending volume

3. Production hardening:
   - Enable DDoS protection
   - Set up WAF rules
   - Configure rate limiting

## Reference Links

- [AWS Route 53 Documentation](https://docs.aws.amazon.com/route53/)
- [Azure DNS Documentation](https://docs.microsoft.com/en-us/azure/dns/)
- [Azure Communication Services Email](https://docs.microsoft.com/en-us/azure/communication-services/concepts/email/email-overview)
- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Keycloak Documentation](https://www.keycloak.org/documentation)

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Review Terraform logs: `terraform apply -auto-approve 2>&1 | tee terraform.log`
3. Check Kubernetes events: `kubectl get events -A --sort-by='.lastTimestamp'`
4. Review application logs: `kubectl logs -n app -l app=keycloak --tail=100`
