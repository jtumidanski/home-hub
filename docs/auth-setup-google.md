# Google OAuth App Setup Guide

This guide walks you through creating a Google OAuth 2.0 application for Home Hub authentication.

---

## Prerequisites

- A Google account
- Access to [Google Cloud Console](https://console.cloud.google.com/)

---

## Step 1: Access Google Cloud Console

1. Navigate to https://console.cloud.google.com/
2. Sign in with your Google account

---

## Step 2: Create or Select a Project

### Option A: Create a New Project
1. Click the project dropdown in the top navigation bar
2. Click **"NEW PROJECT"**
3. Enter project details:
   - **Project name:** `Home Hub` (or your preferred name)
   - **Organization:** Select if applicable, or leave as "No organization"
4. Click **"CREATE"**
5. Wait for the project to be created (usually takes a few seconds)
6. Select the newly created project from the project dropdown

### Option B: Use an Existing Project
1. Click the project dropdown in the top navigation bar
2. Select your existing project

---

## Step 3: Enable Google+ API (Required for OAuth)

1. In the left sidebar, navigate to **APIs & Services > Library**
2. Search for "Google+ API"
3. Click on **"Google+ API"**
4. Click **"ENABLE"**
5. Wait for the API to be enabled

**Note:** While Google+ has been deprecated as a social network, the Google+ API is still required for OAuth 2.0 to access basic profile information.

---

## Step 4: Configure OAuth Consent Screen

1. In the left sidebar, navigate to **APIs & Services > OAuth consent screen**
2. Choose user type:
   - **Internal:** Only users in your Google Workspace organization can sign in (recommended for private household use)
   - **External:** Anyone with a Google account can sign in (must go through verification for production)
3. Click **"CREATE"**

### Fill in App Information:
- **App name:** `Home Hub`
- **User support email:** Your email address
- **App logo:** (Optional) Upload a logo
- **Application home page:** `http://homehub.localtest.me` (for local) or your production domain
- **Authorized domains:** Add your production domain (e.g., `example.com`)
- **Developer contact information:** Your email address

4. Click **"SAVE AND CONTINUE"**

### Scopes:
5. Click **"ADD OR REMOVE SCOPES"**
6. Select the following scopes:
   - `openid`
   - `email`
   - `profile`
   - `.../auth/userinfo.email`
   - `.../auth/userinfo.profile`
7. Click **"UPDATE"**
8. Click **"SAVE AND CONTINUE"**

### Test Users (for External apps):
9. If you chose "External" user type:
   - Click **"ADD USERS"**
   - Add email addresses of users who should be able to test the app
   - Click **"ADD"**
10. Click **"SAVE AND CONTINUE"**

### Summary:
11. Review your configuration
12. Click **"BACK TO DASHBOARD"**

---

## Step 5: Create OAuth 2.0 Credentials

1. In the left sidebar, navigate to **APIs & Services > Credentials**
2. Click **"+ CREATE CREDENTIALS"** at the top
3. Select **"OAuth client ID"**
4. Choose application type: **"Web application"**
5. Enter a name: `Home Hub OAuth Client`

### Configure Authorized JavaScript Origins:
6. Under **"Authorized JavaScript origins"**, click **"+ ADD URI"**
7. Add the following URIs:

**For Local Development:**
```
http://homehub.localtest.me
```

**For Staging/Production:**
```
https://staging.homehub.example.com
https://homehub.example.com
```

### Configure Authorized Redirect URIs:
8. Under **"Authorized redirect URIs"**, click **"+ ADD URI"**
9. Add the following URIs:

**For Local Development:**
```
http://homehub.localtest.me/oauth2/google/callback
```

**For Staging:**
```
https://staging.homehub.example.com/oauth2/google/callback
```

**For Production:**
```
https://homehub.example.com/oauth2/google/callback
```

10. Click **"CREATE"**

---

## Step 6: Copy Credentials

1. A modal will appear with your credentials
2. **Copy and save the following:**
   - **Client ID:** Looks like `123456789-abc123def456.apps.googleusercontent.com`
   - **Client Secret:** Looks like `GOCSPX-abc123def456ghi789`
3. Click **"OK"**

**Important:** Store these credentials securely. You'll need them for your `.env` file.

---

## Step 7: Configure Environment Variables

Add the following to your `.env` file:

```bash
# Google OAuth
GOOGLE_CLIENT_ID=123456789-abc123def456.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-abc123def456ghi789
GOOGLE_COOKIE_SECRET=<generate with command below>
```

### Generate Cookie Secret:
```bash
python -c 'import os,base64; print(base64.urlsafe_b64encode(os.urandom(32)).decode())'
```

Copy the output and use it as `GOOGLE_COOKIE_SECRET`.

---

## Verification

To verify your OAuth app is configured correctly:

1. Start your local development environment
2. Navigate to `http://homehub.localtest.me/admin`
3. You should be redirected to Google's sign-in page
4. After signing in, you should be redirected back to your app

If you see an error, check the troubleshooting guide: [auth-troubleshooting.md](./auth-troubleshooting.md)

---

## Common Issues

### Error: "redirect_uri_mismatch"
**Cause:** The redirect URI in your oauth2-proxy configuration doesn't match what's registered in Google Cloud Console.

**Solution:**
1. Verify the redirect URI in your configuration exactly matches one of the URIs in Step 5
2. Remember: `http` vs `https` matters, trailing slashes matter

### Error: "access_denied"
**Cause:** Your Google account is not authorized to use the app.

**Solution:**
- If using "Internal" user type: Ensure you're signing in with a Google Workspace account from your organization
- If using "External" user type in testing: Add your email to the test users list (Step 4)

### Error: "invalid_client"
**Cause:** The client ID or secret is incorrect.

**Solution:**
1. Verify you copied the entire client ID and secret
2. Check for extra spaces or line breaks
3. Regenerate the secret if necessary (Credentials page > Edit client > Reset secret)

---

## Additional Configuration

### Restrict to Specific Domain

If you want to restrict sign-ins to users from a specific domain (e.g., `@yourdomain.com`):

1. In your oauth2-proxy configuration, set:
   ```bash
   --email-domain=yourdomain.com
   ```

2. Users with emails outside this domain will be denied access after authentication.

### Production Verification

If you chose "External" user type and want to move to production (more than 100 users):

1. Navigate to **OAuth consent screen**
2. Click **"PUBLISH APP"**
3. Click **"Prepare for verification"**
4. Follow the verification process (may take several days)

**Note:** For private household use, you likely don't need to go through verification. Keep the app in "Testing" mode and add users manually.

---

## Updating Redirect URIs

If you need to add or change redirect URIs later:

1. Navigate to **APIs & Services > Credentials**
2. Click on your OAuth client name (`Home Hub OAuth Client`)
3. Modify **"Authorized redirect URIs"**
4. Click **"SAVE"**
5. Changes take effect immediately

---

## Security Best Practices

1. **Never commit credentials to git:** Always use environment variables or secrets management
2. **Rotate secrets regularly:** Generate a new client secret every 90 days (recommended)
3. **Monitor OAuth consent screen:** Check for unexpected users or sign-in patterns
4. **Use HTTPS in production:** Never use HTTP for OAuth in production environments
5. **Restrict email domains:** If possible, limit to your organization's domain

---

## Next Steps

After completing this guide:

1. Configure GitHub OAuth (if using): [auth-setup-github.md](./auth-setup-github.md)
2. Set up oauth2-proxy services: [../deploy/README.md](../deploy/README.md)
3. Test the authentication flow

---

## Resources

- [Google OAuth 2.0 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [Google Cloud Console](https://console.cloud.google.com/)
- [oauth2-proxy Documentation](https://oauth2-proxy.github.io/oauth2-proxy/)
