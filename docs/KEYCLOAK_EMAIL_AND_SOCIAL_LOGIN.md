# Keycloak Email and Social Login Configuration

This document describes the configuration for Azure Communication Services Email integration and Google social login in Keycloak.

## Overview

The Keycloak instance has been configured to support:
1. **Azure Communication Services Email SMTP** - For sending password reset emails and email verification
2. **Google OAuth 2.0** - For social login (Sign in with Google)

## Prerequisites

### 1. Azure Communication Services Email Setup

1. Create an Azure account at https://portal.azure.com/
2. Create an Azure Communication Services resource:
   - Search for "Communication Services" in the Azure Portal
   - Click "Create"
   - Select your subscription and resource group
   - Choose a region (e.g., East US)
   - Provide a name for your resource

3. **Set up Email Communication Service**:
   - In your Communication Services resource, go to "Email" → "Domains"
   - You can use Azure Managed Domain (free) or connect your own custom domain
   - For custom domain:
     - Add your domain
     - Verify ownership by adding DNS records (TXT, SPF, DKIM)
     - Wait for verification (can take up to 24 hours)

4. **Configure Sender Authentication**:
   - Go to "Email" → "MailFrom addresses"
   - Add your sender email address (e.g., noreply@coachpotato.app)
   - The address must be from a verified domain

5. **Get SMTP Credentials**:
   - Azure Communication Services Email supports SMTP
   - SMTP Host: `smtp.azurecomm.net`
   - SMTP Port: `587` (with STARTTLS)
   - **Username**: Your verified sender email address
   - **Password**: Connection string from your Azure Communication Services resource
     - Go to your Communication Services resource
     - Navigate to "Keys" or "Connection strings"
     - Copy the connection string

6. **Pricing** (as of 2025):
   - Free tier: 500 emails per month
   - After free tier: $0.25 per 1,000 emails for Azure-managed domains
   - Custom domains: $0.25 per 1,000 emails
   - No data transfer costs for email

### 2. Google OAuth 2.0 Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API (or Google Identity)
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client ID"
5. Configure the OAuth consent screen:
   - Add your application name
   - Add authorized domains
   - Add scopes: `email`, `profile`, `openid`
6. Create OAuth 2.0 credentials:
   - **Application type**: Web application
   - **Authorized redirect URIs**:
     - `https://auth.your-domain.com/realms/powerlifting-coach/broker/google/endpoint`
     - `http://localhost:8080/realms/powerlifting-coach/broker/google/endpoint` (for local testing)
7. Copy the **Client ID** and **Client Secret**

## Configuration

### Setting Environment Variables

Set the following environment variables when applying Terraform:

```bash
export TF_VAR_azure_email_smtp_host="smtp.azurecomm.net"
export TF_VAR_azure_email_smtp_username="noreply@coachpotato.app"
export TF_VAR_azure_email_smtp_password="your-azure-connection-string"
export TF_VAR_azure_email_from_email="noreply@coachpotato.app"
export TF_VAR_google_oauth_client_id="your-google-client-id.apps.googleusercontent.com"
export TF_VAR_google_oauth_client_secret="your-google-client-secret"
```

Or create a `terraform.tfvars` file in the `infrastructure/` directory:

```hcl
azure_email_smtp_host      = "smtp.azurecomm.net"
azure_email_smtp_username  = "noreply@coachpotato.app"
azure_email_smtp_password  = "your-azure-connection-string"
azure_email_from_email     = "noreply@coachpotato.app"
google_oauth_client_id     = "your-google-client-id.apps.googleusercontent.com"
google_oauth_client_secret = "your-google-client-secret"
```

**Important**: Never commit `terraform.tfvars` to version control if it contains sensitive values!

### Azure Communication Services Regions

Azure Communication Services is available in multiple regions. Common regions include:
- **United States**: East US, West US 2, Central US
- **Europe**: North Europe, West Europe
- **Asia Pacific**: Southeast Asia, Australia East
- **Others**: Canada Central, UK South, etc.

Choose the region closest to your infrastructure for best performance.

### Applying the Configuration

1. Navigate to the infrastructure directory:
   ```bash
   cd infrastructure
   ```

2. Initialize Terraform (if not already done):
   ```bash
   terraform init
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

4. The configuration will create Kubernetes secrets:
   - `azure-email-secret` - Contains Azure Communication Services Email SMTP credentials
   - `google-oauth-secret` - Contains Google OAuth credentials

## Features Enabled

### Email Features

With Azure Communication Services Email configured, the following features are now available:

1. **Password Reset**: Users can request password reset emails
2. **Email Verification**: New user registrations will send verification emails
3. **Account Notifications**: Admin actions and account changes can trigger email notifications

### Social Login Features

With Google OAuth configured:

1. **Sign in with Google**: Users can log in using their Google account
2. **Sign up with Google**: New users can register using their Google account
3. **Account Linking**: Existing email/password accounts can be linked to Google accounts
4. **Automatic Profile Mapping**: Email and name are automatically synced from Google

## Testing

### Testing Azure Email Integration

1. Access Keycloak admin console at `https://auth.your-domain.com/admin`
2. Login with admin credentials
3. Navigate to: Realm Settings → Email
4. Click "Test connection" to verify SMTP settings
5. Try the password reset flow:
   - Go to login page
   - Click "Forgot Password?"
   - Enter your email
   - Check for reset email

### Testing Google Login

1. Go to your application login page
2. You should see a "Sign in with Google" button
3. Click it and authenticate with a Google account
4. You should be redirected back and logged in
5. Check the Keycloak admin console to verify the user was created

## Troubleshooting

### Email Not Sending

1. **Verify Domain Verification**:
   - Check Azure Portal → Communication Services → Email → Domains
   - Ensure your domain status is "Verified"
   - For custom domains, verify DNS records are correctly configured

2. **Check SMTP Credentials**:
   - Ensure the connection string is correct
   - Verify the sender email address is from a verified domain
   - Check that the username matches your verified sender address

3. **View Keycloak logs**:
   ```bash
   kubectl logs -n app deployment/keycloak
   ```

4. **Ensure the Azure email secret is properly created**:
   ```bash
   kubectl get secret -n app azure-email-secret
   kubectl describe secret -n app azure-email-secret
   ```

5. **Check Azure Email Metrics**:
   - Go to Azure Portal → Communication Services → Email → Metrics
   - Check for delivery failures or authentication errors

6. **Verify SMTP Connection**:
   - SMTP host must be `smtp.azurecomm.net`
   - Port must be `587` with STARTTLS enabled
   - Ensure firewall rules allow outbound connections on port 587

### Google Login Not Working

1. Verify the redirect URIs in Google Cloud Console match exactly
2. Check that the OAuth consent screen is published (not in testing mode)
3. Verify the Google OAuth secret:
   ```bash
   kubectl get secret -n app google-oauth-secret
   ```
4. Check Keycloak logs for authentication errors

### Environment Variables Not Loading

The realm configuration uses environment variable substitution syntax:
- `${env.AZURE_EMAIL_SMTP_HOST}`
- `${env.AZURE_EMAIL_SMTP_PASSWORD}`
- `${env.GOOGLE_CLIENT_ID}`

These are resolved by Keycloak at startup. If they're not working:

1. Verify the environment variables are set in the deployment:
   ```bash
   kubectl describe deployment -n app keycloak
   ```
2. Restart Keycloak to reload configuration:
   ```bash
   kubectl rollout restart deployment/keycloak -n app
   ```

### Azure Email Rate Limiting

Azure Communication Services has rate limits:
1. Check your quota in Azure Portal → Communication Services → Quotas
2. Free tier: 500 emails/month
3. Request a quota increase if needed in Azure Portal
4. Monitor Azure metrics for throttling events

## Security Considerations

1. **SMTP Credentials**: Stored as Kubernetes secrets, never in plain text
2. **Connection String**: Azure connection strings should be rotated regularly
3. **OAuth Secrets**: Stored as Kubernetes secrets, never committed to git
4. **STARTTLS**: Enabled for Azure Email SMTP connection (port 587)
5. **HTTPS Only**: In production, ensure `sslRequired: "external"` in realm config
6. **Email Verification**: Enabled to prevent fake account creation
7. **Trusted Email**: Google emails are marked as trusted since Google verifies them
8. **Domain Verification**: Keep DNS records secure and monitor for changes

## Cost Considerations

Azure Communication Services Email Pricing (as of 2025):
- First 500 emails per month: **Free**
- After free tier: $0.25 per 1,000 emails
- No data transfer costs
- No minimum commitment

This is cost-effective for small to medium applications and integrates well with other Azure services.

## File References

- Realm Configuration: `k8s/base/keycloak-realm.yaml`
- Keycloak Deployment: `k8s/base/keycloak.yaml`
- Terraform Secrets: `infrastructure/kubernetes-resources.tf`
- Variables: `infrastructure/variables.tf`

## Additional Resources

- [Keycloak Email Configuration](https://www.keycloak.org/docs/latest/server_admin/#_email)
- [Keycloak Identity Providers](https://www.keycloak.org/docs/latest/server_admin/#_identity_broker)
- [Azure Communication Services Email Overview](https://learn.microsoft.com/en-us/azure/communication-services/concepts/email/email-overview)
- [Azure Communication Services Email Quickstart](https://learn.microsoft.com/en-us/azure/communication-services/quickstarts/email/send-email)
- [Azure Email SMTP Documentation](https://learn.microsoft.com/en-us/azure/communication-services/quickstarts/email/send-email-smtp)
- [Google OAuth 2.0 Setup](https://developers.google.com/identity/protocols/oauth2)
