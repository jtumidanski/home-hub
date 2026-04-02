# Domain

## User

### Responsibility

Represents an authenticated user in the system. Users are identified by email and carry profile information sourced from the external identity provider.

### Core Models

**Model** (`user.Model`)

| Field             | Type      |
|-------------------|-----------|
| id                | uuid.UUID |
| email             | string    |
| displayName       | string    |
| givenName         | string    |
| familyName        | string    |
| avatarURL         | string    |
| providerAvatarURL | string    |
| createdAt         | time.Time |
| updatedAt         | time.Time |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Email is unique across all users.
- Users are created via FindOrCreate: if a user with the given email exists, it is returned; otherwise a new user is created.
- `avatarURL` stores user-selected avatar descriptors in `dicebear:{style}:{seed}` format or empty string.
- `providerAvatarURL` stores the OIDC provider's picture URL.
- Avatar format validation accepts `dicebear:(adventurer|bottts|fun-emoji):{alphanumeric seed, 1-64 chars}` or empty string.

### Processors

**Processor** (`user.Processor`)

| Method                                                              | Description                                                  |
|---------------------------------------------------------------------|--------------------------------------------------------------|
| `ByIDProvider(id)`                                                  | Lazy lookup by ID                                            |
| `ByEmailProvider(email)`                                            | Lazy lookup by email                                         |
| `ByIDsProvider(ids)`                                                | Lazy batch lookup by IDs                                     |
| `FindOrCreate(email, displayName, givenName, familyName, avatarURL)` | Returns existing user by email or creates new                |
| `UpdateProviderAvatar(userID, url)`                                 | Updates provider_avatar_url without touching user avatar      |
| `UpdateAvatar(userID, avatarURL)`                                   | Validates format and persists user-selected avatar, returns updated model |

**ValidateAvatarFormat** (`user.ValidateAvatarFormat`)

Package-level function. Accepts empty string or `dicebear:{style}:{seed}` format. Returns `ErrInvalidAvatarFormat` otherwise.

---

## External Identity

### Responsibility

Links an external OIDC provider identity (provider name + subject) to an internal user.

### Core Models

**Model** (`externalidentity.Model`)

| Field           | Type      |
|-----------------|-----------|
| id              | uuid.UUID |
| userId          | uuid.UUID |
| provider        | string    |
| providerSubject | string    |
| createdAt       | time.Time |
| updatedAt       | time.Time |

### Invariants

- The (provider, providerSubject) pair is unique.
- Each external identity is linked to exactly one user.

### Processors

**Processor** (`externalidentity.Processor`)

| Method                                 | Description                                |
|----------------------------------------|--------------------------------------------|
| `FindByProviderSubject(provider, sub)` | Lazy lookup by provider and subject        |
| `Create(userID, provider, subject)`    | Links an external identity to a user       |

---

## Refresh Token

### Responsibility

Manages long-lived refresh tokens for session continuity. Tokens are stored as SHA256 hashes and support rotation and revocation.

### Core Models

**Model** (`refreshtoken.Model`)

| Field     | Type      |
|-----------|-----------|
| id        | uuid.UUID |
| userId    | uuid.UUID |
| tokenHash | string    |
| expiresAt | time.Time |
| revoked   | bool      |
| createdAt | time.Time |
| updatedAt | time.Time |

### Invariants

- Raw tokens are 32 random bytes encoded as 64 hex characters.
- Only the SHA256 hash is stored; raw tokens are never persisted.
- Token TTL is 7 days from creation.
- Validation checks hash match, revoked flag, and expiration.
- Rotation revokes the old token before issuing a new one.

### Processors

**Processor** (`refreshtoken.Processor`)

| Method                    | Description                                                |
|---------------------------|------------------------------------------------------------|
| `Create(userID)`          | Generates a new token, stores its hash, returns raw string |
| `Validate(raw)`           | Verifies token validity, returns user ID                   |
| `Rotate(oldRaw)`          | Validates and revokes old token, issues new one            |
| `RevokeAllForUser(userID)` | Revokes all tokens for a user                             |

---

## OIDC Provider

### Responsibility

Represents a configured external OIDC identity provider.

### Core Models

**Model** (`oidcprovider.Model`)

| Field     | Type      |
|-----------|-----------|
| id        | uuid.UUID |
| name      | string    |
| issuerURL | string    |
| clientID  | string    |
| enabled   | bool      |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Name is required.
- Provider list is currently derived from configuration, not from the database.
- A Google provider entry is returned when `OIDC_CLIENT_ID` is configured.
- Google provider uses a fixed UUID (`00000000-0000-0000-0000-000000000001`).

### Processors

**Processor** (`oidcprovider.Processor`)

| Method          | Description                                     |
|-----------------|-------------------------------------------------|
| `ListEnabled()` | Returns all enabled OIDC providers from config  |

---

## Auth Flow

### Responsibility

Orchestrates the full authentication lifecycle: OIDC callback handling, token refresh, and logout. Coordinates across the user, external identity, refresh token, and JWT domains.

### Core Models

**CallbackResult** (`authflow.CallbackResult`)

| Field        | Type   |
|--------------|--------|
| AccessToken  | string |
| RefreshToken | string |

**RefreshResult** (`authflow.RefreshResult`)

| Field        | Type   |
|--------------|--------|
| AccessToken  | string |
| RefreshToken | string |

### Processors

**Processor** (`authflow.Processor`)

| Method                              | Description                                                                                          |
|-------------------------------------|------------------------------------------------------------------------------------------------------|
| `HandleCallback(oidcCfg, code)`     | Exchanges authorization code, fetches user info, finds or creates user, links identity, issues tokens |
| `HandleCallbackWithUserInfo(info)`  | Performs user/identity/token steps given resolved user info                                           |
| `HandleRefresh(oldRefreshToken)`    | Rotates refresh token and issues a new access token                                                  |
| `HandleLogout(userID)`              | Revokes all refresh tokens for the user                                                              |

---

## JWT

### Responsibility

Issues RS256-signed JWT access tokens and exposes the public key via JWKS.

### Core Models

**Claims** (`jwt.Claims`)

| Field       | Type      |
|-------------|-----------|
| (standard)  | jwtgo.RegisteredClaims |
| UserID      | uuid.UUID |
| Email       | string    |
| TenantID    | uuid.UUID |
| HouseholdID | uuid.UUID |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Access token TTL is 15 minutes.
- Issuer claim is `home-hub-auth`.
- Subject claim is the user UUID.
- Signing algorithm is RS256.
- JWT header includes `kid`.
- Private key accepts PKCS#1 or PKCS#8 PEM formats.

### Processors

**Issuer** (`jwt.Issuer`)

| Method                              | Description                              |
|-------------------------------------|------------------------------------------|
| `NewIssuer(pemKey, kid)`            | Creates issuer from PEM private key      |
| `Issue(userID, email, tenantID, hhID)` | Signs and returns a JWT string        |
| `PublicKey()`                       | Returns the RSA public key               |
| `Kid()`                             | Returns the key ID                       |

**BuildJWKS** (`jwt.BuildJWKS`)

Constructs a JWKS structure from the issuer's public key for the `.well-known/jwks.json` endpoint.

---

## OIDC Protocol

### Responsibility

Implements the OpenID Connect authorization code flow: discovery, authorization URL construction, code exchange, and user info retrieval.

### Core Models

**ProviderConfig** (`oidc.ProviderConfig`)

| Field        | Type   |
|--------------|--------|
| Name         | string |
| IssuerURL    | string |
| ClientID     | string |
| ClientSecret | string |
| RedirectURL  | string |

**UserInfo** (`oidc.UserInfo`)

| Field       | Type   |
|-------------|--------|
| Subject     | string |
| Email       | string |
| DisplayName | string |
| GivenName   | string |
| FamilyName  | string |
| AvatarURL   | string |

**Discovery** (`oidc.Discovery`)

| Field                 | Type   |
|-----------------------|--------|
| AuthorizationEndpoint | string |
| TokenEndpoint         | string |
| UserinfoEndpoint      | string |

### Invariants

- Discovery fetches `/.well-known/openid-configuration` from the issuer URL.
- Authorization URL uses scope `openid email profile`.
- Code exchange sends `grant_type=authorization_code` with client credentials.
- User info fetched via Bearer token from the userinfo endpoint.
