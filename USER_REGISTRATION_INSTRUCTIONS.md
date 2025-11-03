# User Registration Instructions

## Option 1: Use the Frontend Registration Page (RECOMMENDED)

The easiest way to create a user is through the registration page:

1. Start the frontend (if not already running):
   ```bash
   cd frontend
   npm start
   ```

2. Navigate to: `http://localhost:3000/register`

3. Fill in the form:
   - Name
   - Email
   - Password (min 8 characters)
   - User type: Athlete or Coach

4. Click "Create Account"

5. You'll be automatically logged in and redirected to the onboarding page

**Note:** The registration page calls the auth-service `/api/v1/auth/register` endpoint which creates the user in Keycloak automatically.

---

## Option 2: Manual Registration via API

If you want to register via API directly:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-email@example.com",
    "password": "your-password-123",
    "name": "Your Name",
    "user_type": "athlete"
  }'
```

Valid `user_type` values: `athlete` or `coach`

---

## Option 3: Manual User Creation in Keycloak Admin Console

If Keycloak is deployed and you want to create users manually:

### Step 1: Get Keycloak Admin Password

From Terraform:
```bash
cd infrastructure
terraform output -raw keycloak_admin_password
```

Or from Kubernetes secret:
```bash
kubectl get secret app-secrets -n app -o jsonpath='{.data.keycloak-admin-password}' | base64 -d
```

### Step 2: Access Keycloak Admin Console

1. Port forward to Keycloak service:
   ```bash
   kubectl port-forward svc/keycloak 8080:8080 -n default
   ```

2. Open browser to: `http://localhost:8080/admin`

3. Login with:
   - Username: `admin`
   - Password: (from Step 1)

### Step 3: Create User

1. Select the correct realm (should be `powerlifting-coach` or similar - check your terraform config)

2. Go to: **Users** → **Add user**

3. Fill in:
   - Username: user's email
   - Email: user's email
   - First name: user's name
   - Email verified: ON
   - Enabled: ON

4. Click **Create**

5. Go to **Credentials** tab:
   - Set password
   - Temporary: OFF
   - Click **Set Password**

6. Go to **Attributes** tab:
   - Add attribute:
     - Key: `user_type`
     - Value: `athlete` or `coach`
   - Click **Save**

7. Go to **Role Mappings** tab:
   - Assign the appropriate role (`athlete` or `coach`)

---

## Troubleshooting

### Cannot connect to auth-service

Make sure the auth-service is running:
```bash
kubectl get pods -n default | grep auth-service
kubectl logs -n default <auth-service-pod-name>
```

Port forward if needed:
```bash
kubectl port-forward svc/auth-service 8080:80 -n default
```

### Keycloak not accessible

Check if Keycloak is running:
```bash
kubectl get pods -n default | grep keycloak
```

If not deployed, check terraform:
```bash
cd infrastructure
terraform plan
```

### "User already exists" error

The email is already registered. Either:
1. Use a different email
2. Login with existing credentials
3. Delete the user from Keycloak admin console

---

## Quick Test Users

For testing, you can create these users via the registration page:

- **Athlete Test User:**
  - Email: `athlete@test.com`
  - Password: `test12345`
  - Name: Test Athlete
  - Type: Athlete

- **Coach Test User:**
  - Email: `coach@test.com`
  - Password: `test12345`
  - Name: Test Coach
  - Type: Coach
