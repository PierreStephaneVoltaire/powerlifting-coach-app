# Keycloak Email and Social Login Configuration

This document describes the configuration for AWS SES (Simple Email Service) email integration and Google social login in Keycloak.

## Overview

The Keycloak instance has been configured to support:
1. **AWS SES SMTP** - For sending password reset emails and email verification
2. **Google OAuth 2.0** - For social login (Sign in with Google)

## Prerequisites

### 1. AWS SES Account Setup

1. Create an AWS account at https://aws.amazon.com/
2. Navigate to AWS SES (Simple Email Service)
3. **Verify your email address or domain**:
   - Go to "Verified identities"
   - Click "Create identity"
   - Choose either email address or domain
   - For domain: Follow DNS verification steps (add TXT and CNAME records)
   - For email: Click the verification link sent to your email

4. **Move out of SES Sandbox** (for production):
   - By default, SES is in sandbox mode (can only send to verified addresses)
   - Request production access: SES → Account dashboard → "Request production access"
   - Provide use case details and expected sending volume

5. **Create SMTP Credentials**:
   - Go to SES → SMTP settings
   - Click "Create SMTP credentials"
   - AWS will generate:
     - **SMTP Username** (IAM access key)
     - **SMTP Password** (IAM secret key)
   - **Important**: Save these credentials immediately - the password is only shown once!
   - Note your **SMTP endpoint** (e.g., `email-smtp.us-east-1.amazonaws.com`)

6. **Configure sending limits and reputation** (optional but recommended):
   - Set up a configuration set for tracking bounces and complaints
   - Configure email feedback forwarding
   - Monitor your sending reputation in the SES dashboard

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
export TF_VAR_ses_smtp_host="email-smtp.us-east-1.amazonaws.com"
export TF_VAR_ses_smtp_username="your-ses-smtp-username"
export TF_VAR_ses_smtp_password="your-ses-smtp-password"
export TF_VAR_ses_from_email="noreply@powerliftingcoach.app"
export TF_VAR_google_oauth_client_id="your-google-client-id.apps.googleusercontent.com"
export TF_VAR_google_oauth_client_secret="your-google-client-secret"
```

Or create a `terraform.tfvars` file in the `infrastructure/` directory:

```hcl
ses_smtp_host              = "email-smtp.us-east-1.amazonaws.com"
ses_smtp_username          = "your-ses-smtp-username"
ses_smtp_password          = "your-ses-smtp-password"
ses_from_email             = "noreply@powerliftingcoach.app"
google_oauth_client_id     = "your-google-client-id.apps.googleusercontent.com"
google_oauth_client_secret = "your-google-client-secret"
```

**Important**: Never commit `terraform.tfvars` to version control if it contains sensitive values!

### AWS SES Regions

Common AWS SES SMTP endpoints by region:
- **US East (N. Virginia)**: `email-smtp.us-east-1.amazonaws.com`
- **US West (Oregon)**: `email-smtp.us-west-2.amazonaws.com`
- **EU (Ireland)**: `email-smtp.eu-west-1.amazonaws.com`
- **EU (Frankfurt)**: `email-smtp.eu-central-1.amazonaws.com`
- **Asia Pacific (Tokyo)**: `email-smtp.ap-northeast-1.amazonaws.com`
- **Asia Pacific (Sydney)**: `email-smtp.ap-southeast-2.amazonaws.com`

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
   - `ses-secret` - Contains AWS SES SMTP credentials
   - `google-oauth-secret` - Contains Google OAuth credentials

## Features Enabled

### Email Features

With AWS SES configured, the following features are now available:

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

### Testing AWS SES Integration

1. Access Keycloak admin console at `https://auth.your-domain.com/admin`
2. Login with admin credentials
3. Navigate to: Realm Settings → Email
4. Click "Test connection" to verify SMTP settings
5. Try the password reset flow:
   - Go to login page
   - Click "Forgot Password?"
   - Enter your email
   - Check for reset email
6. If in SES sandbox mode, ensure the recipient email is verified in AWS SES

### Testing Google Login

1. Go to your application login page
2. You should see a "Sign in with Google" button
3. Click it and authenticate with a Google account
4. You should be redirected back and logged in
5. Check the Keycloak admin console to verify the user was created

## Troubleshooting

### Email Not Sending

1. **Check SES Sandbox Status**:
   - If in sandbox mode, you can only send to verified email addresses
   - Request production access in AWS SES console

2. **Verify SMTP Credentials**:
   - Ensure the SMTP username and password are correct
   - If you lost the password, create new SMTP credentials in AWS IAM

3. **Check Email Verification**:
   - Ensure your from email address or domain is verified in AWS SES
   - Check SES → Verified identities

4. **View Keycloak logs**:
   ```bash
   kubectl logs -n app deployment/keycloak
   ```

5. **Ensure the SES secret is properly created**:
   ```bash
   kubectl get secret -n app ses-secret
   kubectl describe secret -n app ses-secret
   ```

6. **Check AWS SES Sending Statistics**:
   - Go to SES dashboard → Account dashboard
   - Check for bounces, complaints, or sending quota issues

7. **Verify SMTP Endpoint**:
   - Ensure you're using the correct regional endpoint
   - The endpoint must match the region where your identities are verified

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
- `${env.SES_SMTP_HOST}`
- `${env.SES_SMTP_PASSWORD}`
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

### AWS SES Rate Limiting

If emails are failing due to rate limits:
1. Check your SES sending limits: SES → Account dashboard
2. Request a sending rate increase if needed
3. Implement exponential backoff in your application
4. Monitor CloudWatch metrics for throttling

## Security Considerations

1. **SMTP Credentials**: Stored as Kubernetes secrets, never in plain text
2. **IAM Permissions**: SES SMTP credentials should have minimal permissions (ses:SendEmail, ses:SendRawEmail)
3. **OAuth Secrets**: Stored as Kubernetes secrets, never committed to git
4. **STARTTLS**: Enabled for AWS SES SMTP connection (port 587)
5. **HTTPS Only**: In production, ensure `sslRequired: "external"` in realm config
6. **Email Verification**: Enabled to prevent fake account creation
7. **Trusted Email**: Google emails are marked as trusted since Google verifies them
8. **SES Reputation**: Monitor bounce and complaint rates to maintain good sending reputation

## Cost Considerations

AWS SES Pricing (as of 2024):
- First 62,000 emails per month: **Free** (when sent from EC2)
- After free tier: $0.10 per 1,000 emails
- Data transfer costs may apply
- Dedicated IP addresses: Additional cost if needed

This is significantly more cost-effective than many email service providers for high-volume sending.

## File References

- Realm Configuration: `k8s/base/keycloak-realm.yaml`
- Keycloak Deployment: `k8s/base/keycloak.yaml`
- Terraform Secrets: `infrastructure/kubernetes-resources.tf`
- Variables: `infrastructure/variables.tf`

## Additional Resources

- [Keycloak Email Configuration](https://www.keycloak.org/docs/latest/server_admin/#_email)
- [Keycloak Identity Providers](https://www.keycloak.org/docs/latest/server_admin/#_identity_broker)
- [AWS SES SMTP Documentation](https://docs.aws.amazon.com/ses/latest/dg/send-email-smtp.html)
- [AWS SES Getting Started](https://docs.aws.amazon.com/ses/latest/dg/getting-started.html)
- [Moving out of SES Sandbox](https://docs.aws.amazon.com/ses/latest/dg/request-production-access.html)
- [Google OAuth 2.0 Setup](https://developers.google.com/identity/protocols/oauth2)
