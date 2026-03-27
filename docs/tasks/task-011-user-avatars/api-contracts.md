# User Avatars — API Contracts

## Auth-Service Endpoints

### PATCH `/api/v1/users/me` (new)

Update the authenticated user's profile. Currently only supports avatar changes.

**Request:**
```json
{
  "data": {
    "type": "users",
    "attributes": {
      "avatarUrl": "dicebear:adventurer:seed123"
    }
  }
}
```

**Validation rules for `avatarUrl`:**
- `"dicebear:{style}:{seed}"` — style must be one of: `adventurer`, `bottts`, `fun-emoji`. Seed is a non-empty alphanumeric string (max 64 chars).
- `""` — clears the user-selected avatar (falls back to provider image or initials)

Any other format returns `422 Unprocessable Entity`.

**Success response:** `200 OK`
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "users",
    "attributes": {
      "email": "user@example.com",
      "displayName": "Jane Doe",
      "givenName": "Jane",
      "familyName": "Doe",
      "avatarUrl": "dicebear:adventurer:seed123",
      "providerAvatarUrl": "https://lh3.googleusercontent.com/...",
      "createdAt": "2026-01-15T10:30:00Z",
      "updatedAt": "2026-03-26T14:00:00Z"
    }
  }
}
```

**Error responses:**
- `401 Unauthorized` — missing or invalid access token
- `422 Unprocessable Entity` — invalid avatar format

---

### GET `/api/v1/users/me` (updated)

No request changes. Response adds `providerAvatarUrl` field.

**Response:** `200 OK`
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "users",
    "attributes": {
      "email": "user@example.com",
      "displayName": "Jane Doe",
      "givenName": "Jane",
      "familyName": "Doe",
      "avatarUrl": "dicebear:adventurer:seed123",
      "providerAvatarUrl": "https://lh3.googleusercontent.com/...",
      "createdAt": "2026-01-15T10:30:00Z",
      "updatedAt": "2026-03-26T14:00:00Z"
    }
  }
}
```

Field semantics:
- `avatarUrl` — the **effective** avatar the frontend should display. Either a `dicebear:` descriptor (user-selected) or empty string (use `providerAvatarUrl` or initials).
- `providerAvatarUrl` — the OIDC provider's picture URL, refreshed on each login. May be empty if the provider doesn't supply one.

---

### GET `/api/v1/users?filter[ids]=...` (updated)

Same field additions as `/users/me`. Each user resource in the response array includes both `avatarUrl` and `providerAvatarUrl`.

---

## DiceBear Avatar URL Resolution (Frontend)

The frontend resolves `dicebear:{style}:{seed}` descriptors to image URLs client-side:

```
https://api.dicebear.com/9.x/{style}/svg?seed={seed}
```

Examples:
- `dicebear:adventurer:abc123` → `https://api.dicebear.com/9.x/adventurer/svg?seed=abc123`
- `dicebear:bottts:xyz789` → `https://api.dicebear.com/9.x/bottts/svg?seed=xyz789`
- `dicebear:fun-emoji:hello` → `https://api.dicebear.com/9.x/fun-emoji/svg?seed=hello`

The frontend may also use the `@dicebear/core` npm package for offline generation if preferred.
