# GitHub OAuth App Setup Guide

This guide walks you through creating a GitHub OAuth application for Home Hub authentication.

---

## Prerequisites

- A GitHub account
- (Optional) Admin access to a GitHub organization, if you want to restrict access to org members

---

## Step 1: Access GitHub Developer Settings

1. Log in to [GitHub](https://github.com/)
2. Click your profile picture in the top-right corner
3. Click **"Settings"**
4. In the left sidebar, scroll down and click **"Developer settings"**
5. Click **"OAuth Apps"** in the left sidebar

---

## Step 2: Create a New OAuth App

1. Click **"New OAuth App"** button
2. You'll see the "Register a new OAuth application" form

---

## Step 3: Fill in Application Details

### Application Information:

**Application name:**
```
Home Hub
```

**Homepage URL:**

**For Local Development:**
```
http://homehub.localtest.me
```

**For Production:**
```
https://homehub.example.com
```

**Application description:** (Optional)
```
Home Hub - Multi-tenant household information platform
```

**Authorization callback URL:**

This is the most important field. Enter the callback URL for your environment:

**For Local Development:**
```
http://homehub.localtest.me/oauth2/github/callback
```

**For Staging:**
```
https://staging.homehub.example.com/oauth2/github/callback
```

**For Production:**
```
https://homehub.example.com/oauth2/github/callback
```

**Note:** GitHub OAuth Apps can only have ONE callback URL. For multiple environments, you'll need to create separate OAuth Apps for local, staging, and production.

### Enable Device Flow: (Optional)
- Leave unchecked unless you need device flow authentication

4. Click **"Register application"**

---

## Step 4: Copy Client ID

After registration, you'll see your new OAuth App's details page.

1. **Copy the Client ID:**
   - Looks like: `Iv1.abc123def456ghi789`
   - Save this for your `.env` file

---

## Step 5: Generate Client Secret

1. Click **"Generate a new client secret"**
2. You may be prompted to confirm your password
3. **Copy the Client Secret immediately:**
   - Looks like: `abc123def456ghi789jkl012mno345pqr678stu901`
   - **Important:** You won't be able to see this secret again!
4. Save this for your `.env` file

---

## Step 6: Configure Environment Variables

Add the following to your `.env` file:

```bash
# GitHub OAuth
GITHUB_CLIENT_ID=Iv1.abc123def456ghi789
GITHUB_CLIENT_SECRET=abc123def456ghi789jkl012mno345pqr678stu901
GITHUB_COOKIE_SECRET=<generate with command below>
```

### Generate Cookie Secret:
```bash
python -c 'import os,base64; print(base64.urlsafe_b64encode(os.urandom(32)).decode())'
```

Copy the output and use it as `GITHUB_COOKIE_SECRET`.

**Important:** Use a DIFFERENT cookie secret than your Google OAuth configuration.

---

## Step 7: (Optional) Create Organization-Restricted OAuth

If you want to restrict access to members of a specific GitHub organization:

### Option A: Use GitHub Organization OAuth App

1. Navigate to your organization page: `https://github.com/orgs/<your-org>/settings`
2. In the left sidebar, click **"Developer settings"**
3. Click **"OAuth Apps"**
4. Click **"New OAuth App"**
5. Follow the same steps as above (Step 3)
6. In your oauth2-proxy configuration, add:
   ```bash
   --github-org=<your-org>
   ```

### Option B: Use Personal OAuth App with Org Restriction

1. Use the OAuth app created in Steps 2-5
2. In your oauth2-proxy configuration, add:
   ```bash
   --github-org=<your-org>
   ```
3. oauth2-proxy will verify the user is a member of the specified organization after authentication

---

## Verification

To verify your GitHub OAuth app is configured correctly:

1. Start your local development environment
2. Navigate to a GitHub-protected route (e.g., `/api/users` with GitHub auth configured)
3. You should be redirected to GitHub's authorization page
4. Click **"Authorize [Your App Name]"**
5. After authorization, you should be redirected back to your app

If you see an error, check the troubleshooting guide: [auth-troubleshooting.md](./auth-troubleshooting.md)

---

## Common Issues

### Error: "The redirect_uri MUST match the registered callback URL for this application."

**Cause:** The callback URL in your oauth2-proxy configuration doesn't match the registered callback URL.

**Solution:**
1. Go to your OAuth App settings in GitHub
2. Verify the "Authorization callback URL" exactly matches your configuration
3. Remember: `http` vs `https` matters, trailing slashes matter
4. GitHub OAuth Apps support only ONE callback URL per app

### Error: "Application suspended"

**Cause:** GitHub has suspended your OAuth app (rare, usually due to policy violations).

**Solution:**
1. Check your email for notifications from GitHub
2. Review GitHub's OAuth App policies
3. Contact GitHub support if you believe this is an error

### Error: "Bad verification code"

**Cause:** The OAuth flow was interrupted or the code expired.

**Solution:**
1. Clear your cookies
2. Try the authentication flow again
3. Ensure your system clock is accurate (OAuth relies on timestamps)

### User authenticated but denied access

**Cause:** User is not a member of the required GitHub organization.

**Solution:**
1. Verify the user is a member of the org specified in `--github-org`
2. Check if the user's organization membership is public or private
3. If membership is private, the user must make it public or the org must grant oauth2-proxy access

---

## Managing Multiple Environments

Since GitHub OAuth Apps support only one callback URL, you have two options for managing multiple environments:

### Option 1: Separate OAuth Apps (Recommended)

Create a separate OAuth App for each environment:
- `Home Hub - Local`
- `Home Hub - Staging`
- `Home Hub - Production`

**Pros:** Clean separation, easy to manage
**Cons:** Must manage multiple sets of credentials

### Option 2: Update Callback URL Per Environment

Use a single OAuth App and update the callback URL when switching environments.

**Pros:** Single set of credentials
**Cons:** Manual updates required, risk of forgetting to update

**Recommendation:** Use Option 1 (separate apps) for cleaner management.

---

## Updating Callback URL

If you need to change the callback URL:

1. Navigate to your OAuth App settings:
   - Personal app: Settings > Developer settings > OAuth Apps > [Your App]
   - Org app: Your org > Settings > Developer settings > OAuth Apps > [Your App]
2. Update the **"Authorization callback URL"**
3. Click **"Update application"**
4. Changes take effect immediately

---

## Organization-Specific Configuration

### Check Organization Membership

To verify a user's organization membership:

1. Navigate to your organization: `https://github.com/orgs/<your-org>/people`
2. Search for the user
3. Check their membership status (Owner, Member)

### Make Organization Membership Public

Users can make their org membership public to avoid issues:

1. User navigates to: `https://github.com/orgs/<your-org>/people`
2. Find their name in the list
3. Click the visibility dropdown next to their name
4. Select **"Public"**

### Grant OAuth App Access to Organization

For private organization membership checks:

1. Navigate to: `https://github.com/orgs/<your-org>/settings/oauth_application_policy`
2. Find your OAuth app in the list
3. Click **"Grant"** to allow the app to see private organization membership

---

## Security Best Practices

1. **Never commit credentials to git:** Always use environment variables or secrets management
2. **Rotate secrets regularly:** Generate a new client secret every 90 days
   - Go to OAuth App settings > "Regenerate client secret"
3. **Monitor authorized apps:** Regularly review which users have authorized your app
4. **Use organization restriction:** If possible, limit access to organization members with `--github-org`
5. **Review app permissions:** Ensure your app only requests necessary scopes (email, profile)
6. **Use HTTPS in production:** Never use HTTP for OAuth in production environments
7. **Set up security notifications:** Enable GitHub security alerts for your OAuth app

---

## Revoking Client Secrets

If you suspect your client secret has been compromised:

1. Navigate to your OAuth App settings
2. Click **"Generate a new client secret"**
3. Update your environment variables with the new secret
4. Deploy the updated configuration
5. The old secret will continue to work until you explicitly revoke it
6. Once deployed, you can delete the old secret from GitHub

---

## Viewing OAuth App Activity

To see who has authorized your app:

1. Navigate to your OAuth App settings
2. Click the **"Insights"** tab (if available)
3. You can see authorization counts and user activity

**Note:** Detailed user lists are not directly available. For comprehensive audit logs, consider implementing your own tracking in the backend.

---

## Next Steps

After completing this guide:

1. Configure Google OAuth (if using): [auth-setup-google.md](./auth-setup-google.md)
2. Set up oauth2-proxy services: [../deploy/README.md](../deploy/README.md)
3. Test the authentication flow

---

## Comparison: GitHub vs GitHub Apps

This guide covers **GitHub OAuth Apps**. There's also **GitHub Apps**, which offer:

**GitHub Apps Advantages:**
- More granular permissions
- Can act as a bot
- Better for automation and integrations

**GitHub OAuth Apps Advantages:**
- Simpler setup for authentication
- Better for user identity verification
- Works well with oauth2-proxy

**Recommendation for Home Hub:** Use GitHub OAuth Apps (this guide) for authentication. GitHub Apps are overkill for our use case.

---

## Resources

- [GitHub OAuth Apps Documentation](https://docs.github.com/en/developers/apps/building-oauth-apps)
- [GitHub OAuth Scopes](https://docs.github.com/en/developers/apps/building-oauth-apps/scopes-for-oauth-apps)
- [oauth2-proxy GitHub Provider](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#github-auth-provider)
- [Authorizing OAuth Apps](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/authorizing-oauth-apps)
