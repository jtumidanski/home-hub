# REST API

All endpoints are prefixed with `/api/v1` and require JWT authentication. Request and response bodies use JSON:API format.

## Endpoints

### GET /api/v1/tenants

Returns the current tenant (derived from auth context) as a single-element array.

**Parameters:** None

**Response:** JSON:API array of `tenants` resources.

| Attribute   | Type   |
|-------------|--------|
| name        | string |
| createdAt   | string |
| updatedAt   | string |

**Error Conditions:**

| Status | Condition          |
|--------|--------------------|
| 404    | Tenant not found   |

---

### POST /api/v1/tenants

Creates a new tenant.

**Parameters:** None

**Request Model:** JSON:API `tenants` resource.

| Attribute | Type   | Required |
|-----------|--------|----------|
| name      | string | yes      |

**Response:** JSON:API `tenants` resource. Status 201.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 500    | Creation failed |

---

### GET /api/v1/tenants/{id}

Returns a tenant by ID.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** JSON:API `tenants` resource.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 404    | Not found      |

---

### GET /api/v1/households

Lists all households for the current tenant.

**Parameters:** None

**Response:** JSON:API array of `households` resources.

| Attribute    | Type    |
|--------------|---------|
| name         | string  |
| timezone     | string  |
| units        | string  |
| latitude     | number  |
| longitude    | number  |
| locationName | string  |
| createdAt    | string  |
| updatedAt    | string  |

**Error Conditions:**

| Status | Condition    |
|--------|--------------|
| 500    | Query failed |

---

### POST /api/v1/households

Creates a new household. Automatically creates an owner membership for the requesting user.

**Parameters:** None

**Request Model:** JSON:API `households` resource.

| Attribute | Type   | Required |
|-----------|--------|----------|
| name      | string | yes      |
| timezone  | string | yes      |
| units     | string | yes      |

**Response:** JSON:API `households` resource. Status 201.

**Error Conditions:**

| Status | Condition        |
|--------|------------------|
| 400    | Validation failed |
| 500    | Creation failed  |

---

### GET /api/v1/households/{id}

Returns a household by ID.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** JSON:API `households` resource.

**Error Conditions:**

| Status | Condition  |
|--------|------------|
| 400    | Invalid ID |
| 404    | Not found  |

---

### PATCH /api/v1/households/{id}

Updates a household.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Request Model:** JSON:API `households` resource.

| Attribute    | Type   | Required |
|--------------|--------|----------|
| name         | string | yes      |
| timezone     | string | yes      |
| units        | string | yes      |
| latitude     | number | no       |
| longitude    | number | no       |
| locationName | string | no       |

**Response:** JSON:API `households` resource.

**Error Conditions:**

| Status | Condition         |
|--------|-------------------|
| 400    | Invalid ID        |
| 400    | Validation failed |
| 404    | Not found         |
| 500    | Update failed     |

---

### GET /api/v1/households/{id}/members

Lists household members with display names and roles. Display names are resolved from the auth database.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** JSON:API array of `members` resources.

| Attribute   | Type   |
|-------------|--------|
| displayName | string |
| role        | string |

**Relationships:**

| Name      | Type   | Resource Type |
|-----------|--------|---------------|
| user      | to-one | users         |
| household | to-one | households    |

**Error Conditions:**

| Status | Condition    |
|--------|--------------|
| 400    | Invalid ID   |
| 500    | Query failed |

---

### GET /api/v1/memberships

Lists memberships. By default returns the current user's memberships. Supports filtering by household.

**Parameters:**

| Name                 | In    | Type | Required |
|----------------------|-------|------|----------|
| filter[householdId]  | query | UUID | no       |

When `filter[householdId]` is provided, returns all memberships for that household. Each membership includes an `isLastOwner` attribute (true when the member is an owner and is the only owner in the household).

**Response:** JSON:API array of `memberships` resources.

| Attribute   | Type    |
|-------------|---------|
| role        | string  |
| isLastOwner | boolean |
| createdAt   | string  |
| updatedAt   | string  |

**Relationships:**

| Name      | Type   | Resource Type |
|-----------|--------|---------------|
| household | to-one | households    |
| user      | to-one | users         |

**Error Conditions:**

| Status | Condition       |
|--------|-----------------|
| 400    | Invalid filter  |
| 500    | Query failed    |

---

### POST /api/v1/memberships

Creates a membership. Requires `household` and `user` relationships in the request body.

**Parameters:** None

**Request Model:** JSON:API `memberships` resource.

| Attribute | Type   | Required |
|-----------|--------|----------|
| role      | string | yes      |

**Relationships:**

| Name      | Type       | Required |
|-----------|------------|----------|
| household | to-one     | yes      |
| user      | to-one     | yes      |

**Response:** JSON:API `memberships` resource. Status 201.

**Error Conditions:**

| Status | Condition              |
|--------|------------------------|
| 400    | Invalid request body   |
| 400    | Missing relationships  |
| 500    | Creation failed        |

---

### PATCH /api/v1/memberships/{id}

Updates a membership's role. Authorization-checked: requester must be owner or admin.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Request Model:** JSON:API `memberships` resource.

| Attribute | Type   | Required |
|-----------|--------|----------|
| role      | string | yes      |

**Response:** JSON:API `memberships` resource.

**Error Conditions:**

| Status | Condition                                  |
|--------|--------------------------------------------|
| 400    | Invalid ID                                 |
| 400    | Invalid body or missing role               |
| 403    | Not authorized, cannot modify self or owner|
| 404    | Not found                                  |
| 500    | Update failed                              |

---

### DELETE /api/v1/memberships/{id}

Deletes a membership. Authorization-checked. Self-deletion clears active household if it matches the left household.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition                          |
|--------|------------------------------------|
| 400    | Invalid ID                         |
| 403    | Not authorized or cannot remove owner |
| 404    | Not found                          |
| 422    | Last owner cannot leave            |
| 500    | Delete failed                      |

---

### GET /api/v1/preferences

Returns the current user's preference for the current tenant. Auto-creates a preference with default theme "light" if none exists.

**Parameters:** None

**Response:** JSON:API array of `preferences` resources.

| Attribute  | Type   |
|------------|--------|
| theme      | string |
| createdAt  | string |
| updatedAt  | string |

**Error Conditions:**

| Status | Condition    |
|--------|--------------|
| 500    | Query failed |

---

### PATCH /api/v1/preferences/{id}

Updates a preference. Supports updating theme via attributes and active household via the `activeHousehold` relationship.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Request Model:** JSON:API `preferences` resource.

| Attribute | Type   | Required |
|-----------|--------|----------|
| theme     | string | no       |

**Relationships:**

| Name             | Type   | Required |
|------------------|--------|----------|
| activeHousehold  | to-one | no       |

**Response:** JSON:API `preferences` resource.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 400    | Invalid body   |
| 500    | Update failed  |

---

### GET /api/v1/invitations

Lists pending non-expired invitations for a household. Requires the `filter[householdId]` query parameter.

**Parameters:**

| Name                 | In    | Type | Required |
|----------------------|-------|------|----------|
| filter[householdId]  | query | UUID | yes      |

**Response:** JSON:API array of `invitations` resources.

| Attribute  | Type   |
|------------|--------|
| email      | string |
| role       | string |
| status     | string |
| expiresAt  | string |
| createdAt  | string |
| updatedAt  | string |

**Relationships:**

| Name      | Type   | Resource Type |
|-----------|--------|---------------|
| household | to-one | households    |
| invitedBy | to-one | users         |

**Error Conditions:**

| Status | Condition       |
|--------|-----------------|
| 400    | Missing or invalid filter |
| 500    | Query failed    |

---

### GET /api/v1/invitations/mine

Returns pending non-expired invitations for the current user's email. Bypasses tenant filtering. Includes associated household information.

**Parameters:** None

**Response:** JSON:API array of `invitations` resources with included `households` resources.

| Attribute  | Type   |
|------------|--------|
| email      | string |
| role       | string |
| status     | string |
| expiresAt  | string |
| createdAt  | string |
| updatedAt  | string |

**Relationships:**

| Name      | Type   | Resource Type |
|-----------|--------|---------------|
| household | to-one | households    |
| invitedBy | to-one | users         |

**Error Conditions:**

| Status | Condition         |
|--------|-------------------|
| 401    | Email not in token |
| 500    | Query failed      |

---

### POST /api/v1/invitations

Creates an invitation. Requires owner or admin role in the target household.

**Parameters:** None

**Request Model:** JSON:API `invitations` resource.

| Attribute | Type   | Required |
|-----------|--------|----------|
| email     | string | yes      |
| role      | string | no       |

**Relationships:**

| Name      | Type   | Required |
|-----------|--------|----------|
| household | to-one | yes      |

**Response:** JSON:API `invitations` resource. Status 201.

**Error Conditions:**

| Status | Condition                       |
|--------|---------------------------------|
| 400    | Missing email or household      |
| 403    | Not authorized (not owner/admin)|
| 409    | Pending invitation already exists |
| 422    | Invalid role or already a member|
| 500    | Creation failed                 |

---

### DELETE /api/v1/invitations/{id}

Revokes a pending invitation. Requires owner or admin role in the household.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition            |
|--------|----------------------|
| 403    | Not authorized       |
| 404    | Not found or not pending |
| 500    | Revoke failed        |

---

### POST /api/v1/invitations/{id}/accept

Accepts a pending invitation. Creates membership and preference. Email validated against token.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** JSON:API `invitations` resource (status=accepted).

**Error Conditions:**

| Status | Condition                           |
|--------|-------------------------------------|
| 401    | Email not in token                  |
| 403    | Email does not match invitation     |
| 404    | Not found or not pending            |
| 410    | Invitation has expired              |
| 422    | Cross-tenant or duplicate membership|
| 500    | Accept failed                       |

---

### POST /api/v1/invitations/{id}/decline

Declines a pending invitation. Email validated against token.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** JSON:API `invitations` resource (status=declined).

**Error Conditions:**

| Status | Condition                       |
|--------|---------------------------------|
| 401    | Email not in token              |
| 403    | Email does not match invitation |
| 404    | Not found or not pending        |
| 500    | Decline failed                  |

---

### GET /api/v1/contexts/current

Returns the fully resolved application context for the current user and tenant.

**Parameters:** None

**Response:** JSON:API `contexts` resource with ID "current".

| Attribute              | Type    |
|------------------------|---------|
| resolvedTheme          | string  |
| resolvedRole           | string  |
| canCreateHousehold     | boolean |
| pendingInvitationCount | number  |

**Relationships:**

| Name             | Type    | Resource Type |
|------------------|---------|---------------|
| tenant           | to-one  | tenants       |
| preference       | to-one  | preferences   |
| activeHousehold  | to-one  | households    |
| memberships      | to-many | memberships   |

**Error Conditions:**

| Status | Condition             |
|--------|-----------------------|
| 500    | Resolution failed     |
