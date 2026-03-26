# Household Management — API Contracts

## Invitation Endpoints

All endpoints are under `/api/v1` and require JWT authentication unless otherwise noted.

---

### List Invitations (by household)

```
GET /api/v1/invitations?filter[householdId]={householdId}
```

**Authorization**: Any member of the household.

**Response** `200 OK`:
```json
{
  "data": [
    {
      "type": "invitations",
      "id": "uuid",
      "attributes": {
        "email": "invitee@example.com",
        "role": "editor",
        "status": "pending",
        "expiresAt": "2026-04-02T12:00:00Z",
        "createdAt": "2026-03-26T12:00:00Z",
        "updatedAt": "2026-03-26T12:00:00Z"
      },
      "relationships": {
        "household": {
          "data": { "type": "households", "id": "uuid" }
        },
        "invitedBy": {
          "data": { "type": "users", "id": "uuid" }
        }
      }
    }
  ]
}
```

**Errors**:
- `403 Forbidden` — User is not a member of the household.

---

### List Invitations (for current user)

```
GET /api/v1/invitations/mine?filter[status]=pending
```

**Authorization**: Authenticated user. Bypasses tenant filtering. Email derived from JWT — never a query parameter.

**Response** `200 OK`:
```json
{
  "data": [
    {
      "type": "invitations",
      "id": "uuid",
      "attributes": {
        "email": "user@example.com",
        "role": "editor",
        "status": "pending",
        "expiresAt": "2026-04-02T12:00:00Z",
        "createdAt": "2026-03-26T12:00:00Z",
        "updatedAt": "2026-03-26T12:00:00Z"
      },
      "relationships": {
        "household": {
          "data": { "type": "households", "id": "uuid" }
        },
        "invitedBy": {
          "data": { "type": "users", "id": "uuid" }
        }
      }
    }
  ],
  "included": [
    {
      "type": "households",
      "id": "uuid",
      "attributes": {
        "name": "The Smith Home"
      }
    }
  ]
}
```

**Notes**:
- Includes household resources so the join flow can display household names.
- Expired invitations are excluded even if `status=pending`.

---

### Create Invitation

```
POST /api/v1/invitations
```

**Authorization**: Owner or admin of the target household.

**Request**:
```json
{
  "data": {
    "type": "invitations",
    "attributes": {
      "email": "invitee@example.com",
      "role": "editor"
    },
    "relationships": {
      "household": {
        "data": { "type": "households", "id": "{householdId}" }
      }
    }
  }
}
```

| Field | Required | Default | Constraints |
|-------|----------|---------|-------------|
| email | Yes | — | Valid email format |
| role | No | `viewer` | One of: `admin`, `editor`, `viewer` |
| household | Yes | — | Must be a household the user has privileged access to |

**Response** `201 Created`:
```json
{
  "data": {
    "type": "invitations",
    "id": "uuid",
    "attributes": {
      "email": "invitee@example.com",
      "role": "editor",
      "status": "pending",
      "expiresAt": "2026-04-02T12:00:00Z",
      "createdAt": "2026-03-26T12:00:00Z",
      "updatedAt": "2026-03-26T12:00:00Z"
    },
    "relationships": {
      "household": {
        "data": { "type": "households", "id": "{householdId}" }
      },
      "invitedBy": {
        "data": { "type": "users", "id": "{currentUserId}" }
      }
    }
  }
}
```

**Errors**:
- `403 Forbidden` — User is not owner or admin of the household.
- `409 Conflict` — A pending invitation already exists for this email + household.
- `422 Unprocessable Entity` — Email already has a membership in this household, or role is `owner`.

---

### Revoke Invitation

```
DELETE /api/v1/invitations/{id}
```

**Authorization**: Owner or admin of the invitation's household.

**Response** `204 No Content`

**Errors**:
- `403 Forbidden` — User is not owner or admin.
- `404 Not Found` — Invitation not found or not in `pending` status.

---

### Accept Invitation

```
POST /api/v1/invitations/{id}/accept
```

**Authorization**: Authenticated user whose email matches the invitation's email.

**Side effects**:
1. Sets invitation status to `accepted`.
2. Creates a membership in the invitation's household with the specified role.
3. If the user has no tenant, assigns the invitation's tenant to the user.
4. If the user already has a tenant that differs from the invitation's tenant, returns 422 (cross-tenant blocked).
5. Creates a preference record if none exists.
6. Sets `activeHouseholdId` on the user's preference to the invitation's household (auto-switch).

**Response** `200 OK`:
```json
{
  "data": {
    "type": "invitations",
    "id": "uuid",
    "attributes": {
      "email": "user@example.com",
      "role": "editor",
      "status": "accepted",
      "expiresAt": "2026-04-02T12:00:00Z",
      "createdAt": "2026-03-26T12:00:00Z",
      "updatedAt": "2026-03-26T12:00:00Z"
    },
    "relationships": {
      "household": {
        "data": { "type": "households", "id": "uuid" }
      },
      "invitedBy": {
        "data": { "type": "users", "id": "uuid" }
      }
    }
  }
}
```

**Errors**:
- `403 Forbidden` — User's email does not match the invitation email.
- `404 Not Found` — Invitation not found or not in `pending` status.
- `410 Gone` — Invitation has expired.
- `422 Unprocessable Entity` — User already has a membership in this household, or user belongs to a different tenant than the invitation's tenant.

---

### Decline Invitation

```
POST /api/v1/invitations/{id}/decline
```

**Authorization**: Authenticated user whose email matches the invitation's email.

**Response** `200 OK`:
```json
{
  "data": {
    "type": "invitations",
    "id": "uuid",
    "attributes": {
      "email": "user@example.com",
      "role": "editor",
      "status": "declined",
      "expiresAt": "2026-04-02T12:00:00Z",
      "createdAt": "2026-03-26T12:00:00Z",
      "updatedAt": "2026-03-26T12:00:00Z"
    },
    "relationships": {
      "household": {
        "data": { "type": "households", "id": "uuid" }
      },
      "invitedBy": {
        "data": { "type": "users", "id": "uuid" }
      }
    }
  }
}
```

**Errors**:
- `403 Forbidden` — User's email does not match the invitation email.
- `404 Not Found` — Invitation not found or not in `pending` status.

---

## Modified Endpoints

### Delete Membership (Leave Household)

```
DELETE /api/v1/memberships/{id}
```

**Authorization changes**:
- A user may delete their **own** membership (leave).
- Owner or admin may delete another user's membership (existing behavior, with new restriction: admin cannot remove an owner).

**New error cases**:
- `422 Unprocessable Entity` — Cannot leave: user is the last owner of the household.

**Side effects** (self-removal):
- If the deleted membership's household was the user's `activeHouseholdId`, set `activeHouseholdId` to null in preferences. The existing appcontext resolution logic falls back to the first remaining membership's household on the next context fetch.

---

### Update Membership Role

```
PATCH /api/v1/memberships/{id}
```

**Authorization changes** (tightened):
- Requires owner or admin role in the membership's household.
- Admin cannot change an owner's role.
- A user cannot change their own role.

**Request** (unchanged format):
```json
{
  "data": {
    "type": "memberships",
    "id": "{id}",
    "attributes": {
      "role": "editor"
    }
  }
}
```

**New error cases**:
- `403 Forbidden` — Admin attempting to modify an owner, or user attempting to modify self.

---

### List Household Members

```
GET /api/v1/memberships?filter[householdId]={householdId}
```

**Authorization**: Any member of the household.

**Response**: Returns all memberships for the household. Each membership includes a computed `isLastOwner` boolean attribute (true when the member has role `owner` and is the only owner in the household). The frontend resolves user details (display name, email, avatar) via the batch user lookup endpoint below.

**Response** `200 OK`:
```json
{
  "data": [
    {
      "type": "memberships",
      "id": "uuid",
      "attributes": {
        "role": "owner",
        "isLastOwner": true,
        "createdAt": "2026-01-15T10:00:00Z",
        "updatedAt": "2026-01-15T10:00:00Z"
      },
      "relationships": {
        "household": {
          "data": { "type": "households", "id": "uuid" }
        },
        "user": {
          "data": { "type": "users", "id": "uuid" }
        }
      }
    }
  ]
}
```

---

## New Auth-Service Endpoint

### Batch User Lookup

```
GET /api/v1/users?filter[ids]=uuid1,uuid2,uuid3
```

**Authorization**: Authenticated user (JWT required). Only returns users who share at least one household membership with the requester (prevents arbitrary profile enumeration).

**Query Parameters**:

| Parameter | Required | Description |
|-----------|----------|-------------|
| filter[ids] | Yes | Comma-separated list of user UUIDs (max 50) |

**Response** `200 OK`:
```json
{
  "data": [
    {
      "type": "users",
      "id": "uuid",
      "attributes": {
        "email": "jane@example.com",
        "displayName": "Jane Smith",
        "givenName": "Jane",
        "familyName": "Smith",
        "avatarUrl": "https://...",
        "createdAt": "2026-01-15T10:00:00Z",
        "updatedAt": "2026-01-15T10:00:00Z"
      }
    }
  ]
}
```

**Notes**:
- Called by the frontend to enrich membership data with user details.
- The account-service does NOT call this endpoint — service boundaries are preserved.
- Returns only users matching the provided IDs. Unknown IDs are silently omitted.
- Limit: max 50 IDs per request.

---

## App Context Changes

### GET /api/v1/contexts/current

**New attribute**:
```json
{
  "data": {
    "type": "contexts",
    "id": "current",
    "attributes": {
      "resolvedTheme": "light",
      "resolvedRole": "owner",
      "canCreateHousehold": true,
      "pendingInvitationCount": 2
    }
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| pendingInvitationCount | integer | Number of non-expired pending invitations for the current user's email |
