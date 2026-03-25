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

| Attribute  | Type   |
|------------|--------|
| name       | string |
| timezone   | string |
| units      | string |
| createdAt  | string |
| updatedAt  | string |

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

| Status | Condition      |
|--------|----------------|
| 500    | Creation failed |

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

| Attribute | Type   | Required |
|-----------|--------|----------|
| name      | string | yes      |
| timezone  | string | yes      |
| units     | string | yes      |

**Response:** JSON:API `households` resource.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 500    | Update failed  |

---

### GET /api/v1/memberships

Lists the current user's memberships in the current tenant.

**Parameters:** None

**Response:** JSON:API array of `memberships` resources.

| Attribute  | Type   |
|------------|--------|
| role       | string |
| createdAt  | string |
| updatedAt  | string |

**Error Conditions:**

| Status | Condition    |
|--------|--------------|
| 500    | Query failed |

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

Updates a membership's role.

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

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 400    | Invalid body   |
| 500    | Update failed  |

---

### DELETE /api/v1/memberships/{id}

Deletes a membership.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 500    | Delete failed  |

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

### GET /api/v1/contexts/current

Returns the fully resolved application context for the current user and tenant.

**Parameters:** None

**Response:** JSON:API `contexts` resource with ID "current".

| Attribute          | Type    |
|--------------------|---------|
| resolvedTheme      | string  |
| resolvedRole       | string  |
| canCreateHousehold | boolean |

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
