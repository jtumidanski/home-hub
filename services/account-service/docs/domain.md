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

| Field       | Type        |
|-------------|-------------|
| id          | uuid.UUID   |
| tenantID    | uuid.UUID   |
| name        | string      |
| timezone    | string      |
| units       | string      |
| createdAt   | time.Time   |
| updatedAt   | time.Time   |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Each household belongs to exactly one tenant.
- Creation requires tenantID, name, timezone, and units.
- Creating a household auto-creates an owner membership for the requesting user.
- Update allows changing name, timezone, and units.

### Processors

**Processor** (`household.Processor`)

| Method                              | Description                                |
|-------------------------------------|--------------------------------------------|
| `ByIDProvider(id)`                  | Lazy lookup by ID                          |
| `AllProvider()`                             | Lazy list of households (tenant-scoped)    |
| `Create(tenantID, name, tz, units)`        | Creates a new household                    |
| `CreateWithOwner(tenantID, userID, ...)`   | Creates household with owner membership    |
| `Update(id, name, tz, units)`              | Updates an existing household              |

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

| Method                                        | Description                                  |
|-----------------------------------------------|----------------------------------------------|
| `ByIDProvider(id)`                            | Lazy lookup by ID                            |
| `ByUserProvider(userID)`                      | Lazy list of user's memberships (tenant-scoped) |
| `ByHouseholdAndUserProvider(hhID, userID)`    | Lazy lookup of specific membership           |
| `Create(tenantID, hhID, userID, role)`        | Creates a membership                         |
| `UpdateRole(id, role)`                        | Updates the role                             |
| `Delete(id)`                                  | Removes a membership                         |

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

---

## Application Context

### Responsibility

Resolves the full application context for a user by combining tenant, preference, memberships, and active household into a single domain object.

### Core Models

**Resolved** (`appcontext.Resolved`)

| Field              | Type               |
|--------------------|--------------------|
| Tenant             | tenant.Model       |
| ActiveHousehold    | *household.Model   |
| Preference         | preference.Model   |
| Memberships        | []membership.Model |
| ResolvedRole       | string             |
| CanCreateHousehold | bool               |

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

Returns a fully resolved `*Resolved` or an error.
