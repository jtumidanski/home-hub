# Home Hub Authentication Architecture

**Last Updated:** 2025-11-10

---

## Overview

Home Hub uses **oauth2-proxy** with **NGINX Ingress** to provide secure, multi-provider authentication for the platform. This architecture keeps OAuth complexity out of application code while providing flexible identity management through Google and GitHub as identity providers.

### Key Features

- **Multi-provider support:** Google and GitHub OAuth
- **Header-based identity:** No session management in application code
- **Application-managed roles:** Fine-grained RBAC stored in PostgreSQL
- **Zero-trust pattern:** All auth enforcement at ingress layer
- **Stateless services:** Auth state in cookies, identity in headers
- **Future-ready:** Designed for easy migration to Keycloak broker pattern

---

## Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                         User Browser                              │
│  Cookies: _hub_google, _hub_github                                │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                             │ HTTPS Request
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────┐
│                        NGINX Ingress                              │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  auth_request to /oauth2/{provider}/auth                   │  │
│  │  ├─ Authenticated? → Forward with headers                  │  │
│  │  └─ Not auth? → Redirect to /oauth2/{provider}/start       │  │
│  └────────────────────────────────────────────────────────────┘  │
└────────────────────────────┬─────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌────────────────┐  ┌────────────────┐  ┌────────────────┐
│ oauth2-proxy   │  │ oauth2-proxy   │  │ Backend        │
│ (Google)       │  │ (GitHub)       │  │ Services       │
│                │  │                │  │                │
│ Port: 4180     │  │ Port: 4181     │  │ (svc-users,    │
│                │  │                │  │  etc.)         │
└────────┬───────┘  └────────┬───────┘  └────────┬───────┘
         │                   │                   │
         │                   │                   │
         ▼                   ▼                   ▼
┌────────────────────────────────────────────────────────┐
│                  Google / GitHub OAuth                  │
└────────────────────────────────────────────────────────┘
```

---

## Authentication Flow

### Initial Authentication (Google Example)

1. **User visits protected route:** `https://homehub.example.com/admin`

2. **NGINX checks authentication:**
   - Sends subrequest to `/oauth2/google/auth`
   - oauth2-proxy checks for valid `_hub_google` cookie

3. **No valid session → Redirect to login:**
   ```
   302 Redirect → https://homehub.example.com/oauth2/google/start?rd=/admin
   ```

4. **oauth2-proxy redirects to Google:**
   ```
   302 Redirect → https://accounts.google.com/o/oauth2/auth?client_id=...
   ```

5. **User authenticates with Google**

6. **Google redirects back with code:**
   ```
   302 Redirect → https://homehub.example.com/oauth2/google/callback?code=...
   ```

7. **oauth2-proxy exchanges code for token:**
   - Validates token with Google
   - Sets `_hub_google` cookie (secure, httpOnly)
   - Redirects back to original URL: `/admin`

8. **Subsequent requests include cookie:**
   - NGINX auth_request succeeds
   - Headers added to request:
     - `X-Auth-Request-Email: user@example.com`
     - `X-Auth-Request-User: User Name`
     - `X-Forwarded-By: nginx-ingress`

9. **Backend receives authenticated request:**
   - Middleware extracts headers
   - Gets or creates user in database
   - Loads roles from `user_roles` table
   - Attaches auth context to request

---

## Header Propagation

### Headers from oauth2-proxy to NGINX

| Header | Description | Example |
|--------|-------------|---------|
| `X-Auth-Request-Email` | User's email address | `user@example.com` |
| `X-Auth-Request-User` | User's display name | `John Doe` |
| `X-Auth-Request-Groups` | Groups (if applicable) | `github:my-org` |
| `X-Auth-Request-Access-Token` | OAuth access token | `ya29.a0...` |

### Headers from NGINX to Backend

| Header | Description | Set By |
|--------|-------------|--------|
| `X-Auth-Request-Email` | User's email (primary identity) | NGINX (from oauth2-proxy) |
| `X-Auth-Request-User` | User's display name | NGINX (from oauth2-proxy) |
| `X-Forwarded-By` | Request source validation | NGINX |
| `X-Real-IP` | Original client IP | NGINX |
| `X-Forwarded-For` | Proxy chain IPs | NGINX |

### Backend Auth Context

After middleware processes headers, request context contains:

```go
type Context struct {
    UserId   uuid.UUID  // From database
    Email    string     // From X-Auth-Request-Email
    Name     string     // From X-Auth-Request-User
    Provider string     // Inferred: "google" or "github"
    Roles    []string   // From user_roles table
}
```

---

## Role-Based Access Control

### Role Management

Roles are stored in PostgreSQL and managed through the application API:

```sql
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(100) NOT NULL,
    PRIMARY KEY (user_id, role)
);
```

### Standard Roles

| Role | Description | Default |
|------|-------------|---------|
| `user` | Standard authenticated user | ✓ (auto-assigned) |
| `admin` | Full system access | ✗ |
| `household_admin` | Manage household settings | ✗ |
| `device_manager` | Manage devices | ✗ |

### Role Checking in Backend

```go
// In handler
authCtx := auth.MustFromContext(r.Context())

// Check single role
if authCtx.IsAdmin() {
    // Admin-only logic
}

// Check any role
if authCtx.HasAnyRole([]string{"admin", "household_admin"}) {
    // Multi-role logic
}

// Middleware to enforce role
router.Handle("/admin/users",
    auth.RequireRole(logger, "admin")(handler))
```

### Role Assignment API

**Assign Role:**
```http
POST /api/users/{userId}/roles
Content-Type: application/json
X-Auth-Request-Email: admin@example.com

{
  "role": "admin"
}
```

**Remove Role:**
```http
DELETE /api/users/{userId}/roles/{role}
X-Auth-Request-Email: admin@example.com
```

**List User Roles:**
```http
GET /api/users/{userId}/roles
```

---

## Security Model

### Request Source Validation

All backend services MUST validate that requests come from trusted ingress:

```go
func ValidateSource(r *http.Request) error {
    forwardedBy := r.Header.Get("X-Forwarded-By")
    if forwardedBy != "nginx-ingress" {
        return ErrInvalidSource
    }
    return nil
}
```

**Why:** Prevents clients from bypassing authentication by sending `X-Auth-Request-Email` directly to backend services.

### Cookie Security

**Production (HTTPS):**
```yaml
--cookie-secure=true
--cookie-samesite=lax
--cookie-httponly=true
```

**Development (HTTP):**
```yaml
--cookie-secure=false
--cookie-samesite=lax
--cookie-httponly=true
```

### Session Management

- **Expiration:** 168 hours (7 days)
- **Refresh:** Every 1 hour (oauth2-proxy refreshes token automatically)
- **Revocation:** Clear cookie via `/oauth2/{provider}/sign_out`

### Email Domain Restrictions

Restrict to organization domain (Google only):

```yaml
--email-domain=example.com
```

For multiple domains:
```yaml
--email-domain=example.com,example.org
```

### GitHub Organization Restrictions

```yaml
--github-org=my-organization
```

Users must be members of the specified organization.

---

## Deployment Configuration

### Environment Variables

See `.env.example` for complete list. Key variables:

```bash
# Google
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_COOKIE_SECRET=<32-byte base64>
GOOGLE_EMAIL_DOMAIN=*  # Or restrict

# GitHub
GITHUB_CLIENT_ID=...
GITHUB_CLIENT_SECRET=...
GITHUB_COOKIE_SECRET=<32-byte base64>  # Different from Google!
GITHUB_ORG=  # Optional

# General
COOKIE_DOMAIN=homehub.example.com
COOKIE_SECURE=true  # false for local dev
```

### Local Development

1. **DNS:** Use `homehub.localtest.me` (resolves to 127.0.0.1)
2. **OAuth Redirect URIs:**
   ```
   http://homehub.localtest.me/oauth2/google/callback
   http://homehub.localtest.me/oauth2/github/callback
   ```
3. **Start services:**
   ```bash
   docker-compose up -d
   ```
4. **Access:** `http://homehub.localtest.me:3000`

### Production Deployment

1. **Create OAuth apps** (see `docs/auth-setup-google.md` and `docs/auth-setup-github.md`)
2. **Generate secrets:**
   ```bash
   python -c 'import os,base64; print(base64.urlsafe_b64encode(os.urandom(32)).decode())'
   ```
3. **Create Kubernetes secrets:**
   ```bash
   kubectl create secret generic oauth2-google \
     --from-literal=client-id=YOUR_ID \
     --from-literal=client-secret=YOUR_SECRET \
     --from-literal=cookie-secret=YOUR_COOKIE_SECRET \
     --namespace=home-hub
   ```
4. **Apply manifests:**
   ```bash
   kubectl apply -f deploy/k8s/
   ```
5. **Verify:**
   ```bash
   kubectl get pods -n home-hub
   kubectl logs -n home-hub oauth2-proxy-google-xxx
   ```

---

## API Integration

### Frontend: Fetching User Info

```typescript
// Fetch current user
const response = await fetch('/api/me');
const data = await response.json();

// Response structure:
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "displayName": "User Name",
    "provider": "google",
    "householdId": "uuid",  // nullable
    "createdAt": "2025-11-10T12:00:00Z",
    "updatedAt": "2025-11-10T12:00:00Z"
  },
  "roles": ["user", "admin"]
}
```

### Frontend: Logout

Redirect user to:
```
window.location.href = '/oauth2/google/sign_out';
```

Or for GitHub:
```
window.location.href = '/oauth2/github/sign_out';
```

---

## Troubleshooting

### Error: "redirect_uri_mismatch"

**Cause:** Redirect URI doesn't match OAuth app configuration.

**Solution:**
1. Check OAuth app settings in Google Cloud Console or GitHub
2. Ensure exact match: protocol (`http`/`https`), domain, path
3. Update redirect URI in OAuth app if needed

### Error: "missing X-Auth-Request-Email header"

**Cause:** Request not authenticated or not from ingress.

**Solution:**
1. Verify cookie is present in browser (DevTools → Application → Cookies)
2. Check oauth2-proxy logs for errors
3. Verify NGINX auth_request configuration

### Cookie not being set

**Cause:** Cookie domain mismatch or SameSite issues.

**Solution:**
1. Verify `--cookie-domain` matches your domain
2. For local dev, use `homehub.localtest.me` instead of `localhost`
3. Check `--cookie-secure` is `false` for HTTP (local)

### User authenticated but 401 from API

**Cause:** Backend not receiving or validating headers correctly.

**Solution:**
1. Check `X-Forwarded-By` header is set to `nginx-ingress`
2. Verify auth middleware is applied to route
3. Check backend logs for validation errors

### Session expires too quickly

**Cause:** Cookie refresh not configured.

**Solution:**
1. Ensure `--cookie-refresh=1h` is set
2. Check oauth2-proxy can refresh OAuth tokens
3. Verify provider hasn't revoked tokens

---

## Monitoring and Observability

### Metrics

oauth2-proxy exposes Prometheus metrics on `/metrics`:

- `oauth2_proxy_requests_total` - Total requests
- `oauth2_proxy_authentication_attempts_total` - Auth attempts
- `oauth2_proxy_authentication_failures_total` - Auth failures

### Logs

**oauth2-proxy logs:**
```bash
kubectl logs -n home-hub oauth2-proxy-google-xxx --follow
```

**Backend auth logs:**
```bash
kubectl logs -n home-hub svc-users-xxx --follow | grep auth
```

### Key Log Patterns

**Successful auth:**
```
User authenticated: user_id=xxx email=user@example.com provider=google roles=[user]
```

**Auth failure:**
```
Failed to extract auth headers: missing X-Auth-Request-Email header
Request from untrusted source: missing X-Forwarded-By header
```

---

## Future: Keycloak Migration

### Why Keycloak?

- Single identity broker for all providers
- Advanced user management UI
- Federation with LDAP/AD
- Fine-grained permissions
- Built-in 2FA support

### Migration Path

1. **Deploy Keycloak** instance
2. **Configure providers** in Keycloak (Google, GitHub as identity providers)
3. **Update oauth2-proxy** to use Keycloak as OIDC provider:
   ```yaml
   --provider=oidc
   --oidc-issuer-url=https://keycloak.example.com/realms/home-hub
   --client-id=home-hub
   ```
4. **Migrate users** from application database to Keycloak
5. **Update ingress** to use single oauth2-proxy instance
6. **Decommission** per-provider oauth2-proxy instances

### Impact

- **Backend code:** No changes required (same headers)
- **Frontend:** No changes required
- **Users:** Must re-authenticate once
- **Downtime:** Can be done with zero downtime using parallel deployment

---

## References

- [oauth2-proxy Documentation](https://oauth2-proxy.github.io/oauth2-proxy/)
- [NGINX Ingress Auth Annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#external-authentication)
- [Google OAuth Setup](./auth-setup-google.md)
- [GitHub OAuth Setup](./auth-setup-github.md)
- [Troubleshooting Guide](./auth-troubleshooting.md)
- [Home Hub Architecture](./PROJECT_KNOWLEDGE.md)

---

## Quick Reference

### Redirect URIs

**Local:**
```
http://homehub.localtest.me/oauth2/google/callback
http://homehub.localtest.me/oauth2/github/callback
```

**Production:**
```
https://homehub.example.com/oauth2/google/callback
https://homehub.example.com/oauth2/github/callback
```

### Common Commands

**Generate cookie secret:**
```bash
python -c 'import os,base64; print(base64.urlsafe_b64encode(os.urandom(32)).decode())'
```

**Test auth flow:**
```bash
# Should return 401 or redirect
curl -v http://homehub.localtest.me:3000/api/me

# With cookie
curl -v -H "Cookie: _hub_google=xxx" http://homehub.localtest.me:3000/api/me
```

**Check oauth2-proxy health:**
```bash
curl http://localhost:4180/ping
```

---

**Document End**
