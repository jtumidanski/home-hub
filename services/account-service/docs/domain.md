# Domain

## Tenant

### Responsibility

Represents an isolated organizational boundary. All other domain objects belong to a tenant.

### Core Models

**Model** (`tenant.Model`)

| Field       | Type        |
|-------------|-------------|
| id          | uuid.UUID   |
| name        | string      |
| createdAt   | time.Time   |
| updatedAt   | time.Time   |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Tenant creation requires a non-empty name.
- UUIDs are generated at creation time.

### Processors

**Processor** (`tenant.Processor`)

| Method                        | Description                          |
|-------------------------------|--------------------------------------|
| `ByIDProvider(id)`            | Lazy lookup by ID                    |
| `Create(name)`               | Creates a new tenant                 |

---

## Household

### Responsibility

Represents a physical household within a tenant. Households contain members and define locale settings.

### Core Models

**Model** (`household.Model`)

| Field        | Type      |
|--------------|-----------|
| id           | uuid.UUID |
| tenantID     | uuid.UUID |
| name         | string    |
| timezone     | string    |
| units        | string    |
| latitude     | *float64  |
| longitude    | *float64  |
| locationName | *string   |
| createdAt    | time.Time |
| updatedAt    | time.Time |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Each household belongs to exactly one tenant.
- Creation requires tenantID, name, timezone, and units.
- Creating a household auto-creates an owner membership for the requesting user.
- Update allows changing name, timezone, units, latitude, longitude, and locationName.
- Latitude and longitude must be provided together (both or neither).
- Latitude range is [-90, 90]. Longitude range is [-180, 180].

### Processors

**Processor** (`household.Processor`)

| Method                              | Description                                |
|-------------------------------------|--------------------------------------------|
| `ByIDProvider(id)`                  | Lazy lookup by ID                          |
| `AllProvider()`                             | Lazy list of households (tenant-scoped)    |
| `Create(tenantID, name, tz, units)`        | Creates a new household                    |
| `CreateWithOwner(tenantID, userID, ...)`   | Creates household with owner membership    |
| `Update(id, name, tz, units, lat, lon, locationName)` | Updates an existing household   |

---

## Membership

### Responsibility

Represents a user's role within a household. Links a user to a household under a tenant.

### Core Models

**Model** (`membership.Model`)

| Field        | Type        |
|--------------|-------------|
| id           | uuid.UUID   |
| tenantID     | uuid.UUID   |
| householdID  | uuid.UUID   |
| userID       | uuid.UUID   |
| role         | string      |
| createdAt    | time.Time   |
| updatedAt    | time.Time   |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- The (householdID, userID) pair is unique — one role per user per household.
- Creation requires tenantID, householdID, userID, and role.
- Role values used: owner, admin, editor, viewer.

### Processors

**Processor** (`membership.Processor`)

| Method                                        | Description                                     |
|-----------------------------------------------|-------------------------------------------------|
| `ByIDProvider(id)`                            | Lazy lookup by ID                               |
| `ByUserProvider(userID)`                      | Lazy list of user's memberships (tenant-scoped) |
| `ByHouseholdAndUserProvider(hhID, userID)`    | Lazy lookup of specific membership              |
| `ByHouseholdProvider(householdID)`            | Lazy list of household's memberships            |
| `CountOwnersByHousehold(householdID)`         | Count of owners in a household                  |
| `Create(tenantID, hhID, userID, role)`        | Creates a membership                            |
| `UpdateRole(id, role)`                        | Updates the role                                |
| `UpdateRoleAuthorized(id, role, requesterID)` | Authorization-checked role update               |
| `DeleteAuthorized(id, requesterID)`           | Authorization-checked deletion                  |
| `Delete(id)`                                  | Removes a membership                            |

**Authorization Rules**

- `UpdateRoleAuthorized`: Requester must be owner or admin. Cannot modify own role. Admin cannot modify an owner's role.
- `DeleteAuthorized`: Self-deletion allowed for any member, blocked if last owner. Other deletion requires owner or admin. Admin cannot remove an owner.

---

## Preference

### Responsibility

Stores per-user, per-tenant UI preferences including theme and active household selection.

### Core Models

**Model** (`preference.Model`)

| Field              | Type         |
|--------------------|--------------|
| id                 | uuid.UUID    |
| tenantID           | uuid.UUID    |
| userID             | uuid.UUID    |
| theme              | string       |
| activeHouseholdID  | *uuid.UUID   |
| createdAt          | time.Time    |
| updatedAt          | time.Time    |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- The (tenantID, userID) pair is unique — one preference per user per tenant.
- Preferences are auto-created on first access with a default theme of "light".
- activeHouseholdID is nullable.

### Processors

**Processor** (`preference.Processor`)

| Method                                       | Description                                       |
|----------------------------------------------|---------------------------------------------------|
| `ByIDProvider(id)`                           | Lazy lookup by ID                                 |
| `ByUserProvider(userID)`                     | Lazy lookup by user (tenant-scoped)               |
| `FindOrCreate(tenantID, userID)`             | Returns existing or creates with default theme    |
| `UpdateTheme(id, theme)`                     | Updates the theme                                 |
| `SetActiveHousehold(id, householdID)`        | Sets the active household                         |
| `ClearActiveHousehold(id)`                   | Clears the active household                       |

---

## Invitation

### Responsibility

Represents an invitation for an external user to join a household. Manages the invitation lifecycle from creation through acceptance, decline, or revocation.

### Core Models

**Model** (`invitation.Model`)

| Field       | Type      |
|-------------|-----------|
| id          | uuid.UUID |
| tenantID    | uuid.UUID |
| householdID | uuid.UUID |
| email       | string    |
| role        | string    |
| status      | string    |
| invitedBy   | uuid.UUID |
| expiresAt   | time.Time |
| createdAt   | time.Time |
| updatedAt   | time.Time |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Valid roles: admin, editor, viewer. Defaults to viewer.
- Status lifecycle: pending -> (accepted | declined | revoked).
- Only one pending invitation per (householdID, email) pair.
- Invitations expire 7 days after creation.
- Email is normalized to lowercase at creation.
- Inviter must have owner or admin role in the household.

### Processors

**Processor** (`invitation.Processor`)

| Method                                           | Description                                          |
|--------------------------------------------------|------------------------------------------------------|
| `ByIDProvider(id)`                               | Lazy lookup by ID                                    |
| `ByHouseholdPendingProvider(householdID)`        | Pending non-expired invitations for household        |
| `ByEmailPendingProvider(email)`                  | Pending invitations for email (cross-tenant)         |
| `CountByEmailPending(email)`                     | Count of pending invitations for email (cross-tenant)|
| `Create(tenantID, householdID, email, role, inviterID)` | Creates invitation with authorization check   |
| `Revoke(id, revokerID)`                         | Sets pending invitation to revoked                   |
| `Accept(id, userID, userEmail, userTenantID)`    | Accepts invitation, creates membership and preference|
| `Decline(id, userEmail)`                         | Sets pending invitation to declined                  |
| `ByEmailPendingWithHouseholds(email)`            | Returns invitations with associated households       |

**Accept Side Effects**

- Creates membership with the invitation's role in the invitation's household.
- Creates or finds preference for the user, sets active household.
- Updates invitation status to accepted.

**Authorization Rules**

- `Create`: Requires owner or admin role in the household.
- `Revoke`: Requires owner or admin role in the household. Only pending invitations can be revoked.
- `Accept`: User email must match invitation email. User's tenant (if any) must match invitation's tenant.
- `Decline`: User email must match invitation email.

---

## Application Context

### Responsibility

Resolves the full application context for a user by combining tenant, preference, memberships, and active household into a single domain object.

### Core Models

**Resolved** (`appcontext.Resolved`)

| Field                  | Type               |
|------------------------|--------------------|
| Tenant                 | tenant.Model       |
| ActiveHousehold        | *household.Model   |
| Preference             | preference.Model   |
| Memberships            | []membership.Model |
| ResolvedRole           | string             |
| CanCreateHousehold     | bool               |
| PendingInvitationCount | int64              |

### Invariants

- Active household is resolved from the preference's activeHouseholdID. If that is nil or invalid, the first membership's household is used and persisted back to the preference.
- ResolvedRole is the role from the membership matching the active household. Empty string if no active household.
- CanCreateHousehold is true when ResolvedRole is "owner".

### Processors

**Resolve** (`appcontext.Resolve`)

| Parameter  | Type      |
|------------|-----------|
| tenantID   | uuid.UUID |
| userID     | uuid.UUID |
| userEmail  | string    |

Returns a fully resolved `*Resolved` or an error.

- PendingInvitationCount is populated by querying pending invitations matching the user's email.
