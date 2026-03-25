# REST API

All endpoints are prefixed with `/api/v1`. Auth endpoints do not require JWT authentication. The `/users/me` endpoint reads claims from the `access_token` cookie.

## Endpoints

### GET /api/v1/auth/providers

Returns the list of enabled OIDC providers.

**Parameters:** None

**Response:** JSON:API array of `auth-providers` resources.

| Attribute   | Type   |
|-------------|--------|
| displayName | string |

Provider list is currently derived from the `OIDC_CLIENT_ID` environment variable. If set, a Google provider entry is returned.

**Error Conditions:** None

---

### GET /api/v1/auth/login/{provider}

Initiates the OIDC authorization code flow. Sets an `oauth_state` cookie and redirects to the provider's authorization endpoint.

**Parameters:**

| Name     | In   | Type   | Required |
|----------|------|--------|----------|
| provider | path | string | yes      |

Supported values: `google`

**Response:** 302 redirect to provider authorization URL.

**Cookies Set:**

| Name        | MaxAge | HttpOnly | Secure | SameSite |
|-------------|--------|----------|--------|----------|
| oauth_state | 300s   | true     | true   | Lax      |

**Error Conditions:**

| Status | Condition              |
|--------|------------------------|
| 400    | Unknown provider       |
| 500    | OIDC discovery failed  |

---

### GET /api/v1/auth/callback/{provider}

OIDC callback endpoint. Validates state, exchanges authorization code for tokens, creates or finds the user, links external identity, issues JWT and refresh token, and redirects.

**Parameters:**

| Name     | In    | Type   | Required |
|----------|-------|--------|----------|
| provider | path  | string | yes      |
| code     | query | string | yes      |
| state    | query | string | yes      |
| redirect | query | string | no       |

**Response:** 302 redirect to `/app` (or the `redirect` query parameter if provided).

**Cookies Set:**

| Name          | MaxAge  | Path          | HttpOnly | Secure | SameSite |
|---------------|---------|---------------|----------|--------|----------|
| access_token  | 900s    | /             | true     | true   | Strict   |
| refresh_token | 604800s | /api/v1/auth  | true     | true   | Strict   |

The `oauth_state` cookie is cleared after callback.

**Error Conditions:**

| Status | Condition              |
|--------|------------------------|
| 400    | State mismatch         |
| 400    | Missing code           |
| 500    | OIDC discovery failed  |
| 500    | Code exchange failed   |
| 500    | UserInfo fetch failed  |
| 500    | User creation failed   |
| 500    | JWT issuance failed    |
| 500    | Refresh token failed   |

---

### POST /api/v1/auth/token/refresh

Rotates the refresh token. Validates the existing refresh token from the cookie, revokes it, issues a new refresh token and a new JWT access token.

**Parameters:** None

**Cookies Required:**

| Name          | Description       |
|---------------|-------------------|
| refresh_token | Current raw token |

**Response:** 204 No Content.

**Cookies Set:**

| Name          | MaxAge  | Path          | HttpOnly | Secure | SameSite |
|---------------|---------|---------------|----------|--------|----------|
| access_token  | 900s    | /             | true     | true   | Strict   |
| refresh_token | 604800s | /api/v1/auth  | true     | true   | Strict   |

**Error Conditions:**

| Status | Condition                    |
|--------|------------------------------|
| 401    | Missing refresh token cookie |
| 401    | Invalid or expired token     |
| 500    | JWT issuance failed          |

---

### POST /api/v1/auth/logout

Clears auth cookies. Attempts a best-effort revocation of refresh tokens.

**Parameters:** None

**Response:** 204 No Content.

**Cookies Cleared:** `access_token`, `refresh_token`

**Error Conditions:** None

---

### GET /api/v1/auth/.well-known/jwks.json

Returns the JSON Web Key Set containing the service's RSA public key for JWT verification.

**Parameters:** None

**Response:** `application/json`

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "<key-id>",
      "alg": "RS256",
      "n": "<base64url-modulus>",
      "e": "<base64url-exponent>"
    }
  ]
}
```

**Headers:**
- `Cache-Control: public, max-age=3600`

**Error Conditions:** None

---

### GET /api/v1/users/me

Returns the authenticated user's profile. Extracts user ID from the `access_token` cookie claims.

**Parameters:** None

**Cookies Required:**

| Name         | Description |
|--------------|-------------|
| access_token | JWT         |

**Response:** JSON:API `users` resource.

| Attribute   | Type   |
|-------------|--------|
| email       | string |
| displayName | string |
| givenName   | string |
| familyName  | string |
| avatarUrl   | string |
| createdAt   | string |
| updatedAt   | string |

**Error Conditions:**

| Status | Condition                    |
|--------|------------------------------|
| 401    | Missing or invalid token     |
| 404    | User not found               |
