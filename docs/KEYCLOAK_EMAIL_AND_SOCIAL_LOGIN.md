# Keycloak Email and Social Login Configuration

This document describes the configuration for Mailgun email integration and Google social login in Keycloak.

## Overview

The Keycloak instance has been configured to support:
1. **Mailgun SMTP** - For sending password reset emails and email verification
2. **Google OAuth 2.0** - For social login (Sign in with Google)

## Prerequisites

### 1. Mailgun Account Setup

1. Create a Mailgun account at https://www.mailgun.com/
2. Add and verify your domain (e.g., `powerliftingcoach.app`)
3. Obtain your SMTP credentials:
   - **SMTP Username**: Found in Mailgun dashboard under "Domain Settings" → "SMTP credentials"
     - Format: `postmaster@your-domain.mailgun.org`
   - **SMTP Password**: Your Mailgun SMTP password
   - **From Email**: The email address to send from (e.g., `noreply@powerliftingcoach.app`)

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
export TF_VAR_mailgun_smtp_username="postmaster@your-domain.mailgun.org"
export TF_VAR_mailgun_smtp_password="your-mailgun-smtp-password"
export TF_VAR_mailgun_from_email="noreply@powerliftingcoach.app"
export TF_VAR_google_oauth_client_id="your-google-client-id.apps.googleusercontent.com"
export TF_VAR_google_oauth_client_secret="your-google-client-secret"
```

Or create a `terraform.tfvars` file in the `infrastructure/` directory:

```hcl
mailgun_smtp_username      = "postmaster@your-domain.mailgun.org"
mailgun_smtp_password      = "your-mailgun-smtp-password"
mailgun_from_email         = "noreply@powerliftingcoach.app"
google_oauth_client_id     = "your-google-client-id.apps.googleusercontent.com"
google_oauth_client_secret = "your-google-client-secret"
```

**Important**: Never commit `terraform.tfvars` to version control if it contains sensitive values!

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
   - `mailgun-secret` - Contains SMTP credentials
   - `google-oauth-secret` - Contains Google OAuth credentials

## Features Enabled

### Email Features

With Mailgun configured, the following features are now available:

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

### Testing Mailgun Integration

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

1. Verify SMTP credentials in Mailgun dashboard
2. Check that your domain is verified in Mailgun
3. View Keycloak logs:
   ```bash
   kubectl logs -n app deployment/keycloak
   ```
4. Ensure the Mailgun secret is properly created:
   ```bash
   kubectl get secret -n app mailgun-secret
   ```

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
- `${env.MAILGUN_SMTP_PASSWORD}`
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

## Security Considerations

1. **SMTP Credentials**: Stored as Kubernetes secrets, never in plain text
2. **OAuth Secrets**: Stored as Kubernetes secrets, never committed to git
3. **STARTTLS**: Enabled for Mailgun SMTP connection
4. **HTTPS Only**: In production, ensure `sslRequired: "external"` in realm config
5. **Email Verification**: Enabled to prevent fake account creation
6. **Trusted Email**: Google emails are marked as trusted since Google verifies them

## File References

- Realm Configuration: `k8s/base/keycloak-realm.yaml`
- Keycloak Deployment: `k8s/base/keycloak.yaml`
- Terraform Secrets: `infrastructure/kubernetes-resources.tf`
- Variables: `infrastructure/variables.tf`

## Additional Resources

- [Keycloak Email Configuration](https://www.keycloak.org/docs/latest/server_admin/#_email)
- [Keycloak Identity Providers](https://www.keycloak.org/docs/latest/server_admin/#_identity_broker)
- [Mailgun SMTP Documentation](https://documentation.mailgun.com/en/latest/user_manual.html#sending-via-smtp)
- [Google OAuth 2.0 Setup](https://developers.google.com/identity/protocols/oauth2)
