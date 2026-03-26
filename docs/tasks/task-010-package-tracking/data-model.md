# Package Tracking — Data Model

## Schema: `package`

All tables live under the `package` PostgreSQL schema, consistent with other services using per-service schemas.

---

## Entity Relationship

```
packages 1───* tracking_events
packages *───1 (user_id, external reference)
carrier_tokens (system-level, not per-tenant)
```

---

## Tables

### packages

Primary entity for tracked packages.

```sql
CREATE TABLE package.packages (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID        NOT NULL,
    household_id    UUID        NOT NULL,
    user_id         UUID        NOT NULL,
    tracking_number VARCHAR(64) NOT NULL,
    carrier         VARCHAR(16) NOT NULL,
    label           VARCHAR(255),
    notes           TEXT,
    status          VARCHAR(24) NOT NULL DEFAULT 'pre_transit',
    private         BOOLEAN     NOT NULL DEFAULT false,
    estimated_delivery DATE,
    actual_delivery TIMESTAMPTZ,
    last_polled_at  TIMESTAMPTZ,
    last_status_change_at TIMESTAMPTZ,
    archived_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_packages_household_tracking
        UNIQUE (tenant_id, household_id, tracking_number),

    CONSTRAINT chk_packages_carrier
        CHECK (carrier IN ('usps', 'ups', 'fedex')),

    CONSTRAINT chk_packages_status
        CHECK (status IN ('pre_transit', 'in_transit', 'out_for_delivery',
                          'delivered', 'exception', 'stale', 'archived'))
);

CREATE INDEX idx_packages_household_status
    ON package.packages (tenant_id, household_id, status);

CREATE INDEX idx_packages_polling
    ON package.packages (status, last_polled_at)
    WHERE status IN ('pre_transit', 'in_transit', 'out_for_delivery');

CREATE INDEX idx_packages_cleanup
    ON package.packages (status, archived_at)
    WHERE status IN ('delivered', 'archived');
```

**Status transitions:**

```
pre_transit ──→ in_transit ──→ out_for_delivery ──→ delivered ──→ archived
     │               │                │                              ↑
     │               │                │                              │
     └───────────────┴────────────────┴──→ exception          (unarchive)
     │               │                │
     └───────────────┴────────────────┴──→ stale
                                                    (any) ──→ (deleted)
```

**Notes:**
- `status` can move backward if carrier corrects (e.g., `out_for_delivery` back to `in_transit` for re-routing)
- `archived` is only reachable from `delivered` (auto or manual)
- `stale` is set by the cleanup job after 14 days with no status change
- Hard deletion is a row delete (no soft delete — archived serves that purpose)

### tracking_events

Individual tracking scan events from carrier APIs.

```sql
CREATE TABLE package.tracking_events (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id  UUID         NOT NULL REFERENCES package.packages(id) ON DELETE CASCADE,
    timestamp   TIMESTAMPTZ  NOT NULL,
    status      VARCHAR(24)  NOT NULL,
    description VARCHAR(512) NOT NULL,
    location    VARCHAR(255),
    raw_status  VARCHAR(128),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_tracking_events_package_time
    ON package.tracking_events (package_id, timestamp DESC);
```

**Notes:**
- Events are append-only — never updated or deleted except via CASCADE when parent package is deleted
- `status` uses the same enum as `packages.status` to represent the package state at that point in time
- `raw_status` preserves the original carrier status code for debugging
- `location` is best-effort — not all carriers provide it for all events

### carrier_tokens

System-level OAuth token cache for carrier APIs. Not tenant-scoped.

```sql
CREATE TABLE package.carrier_tokens (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    carrier      VARCHAR(16) NOT NULL UNIQUE,
    access_token TEXT        NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Notes:**
- Only three rows (one per carrier) — this is essentially a key-value store
- `access_token` is encrypted at the application level before storage (AES-256-GCM, consistent with calendar-service token encryption)
- Tokens are refreshed proactively before `expires_at`
- Alternative: keep tokens in memory only and re-authenticate on service restart. Database persistence avoids unnecessary re-auth in multi-replica scenarios.

---

## GORM Entity Mapping

Per project conventions, GORM entities handle migration via `AutoMigrate`.

```go
// entity.go (sketch)
type PackageEntity struct {
    ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TenantID           uuid.UUID  `gorm:"type:uuid;not null;index:idx_pkg_household_status"`
    HouseholdID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_pkg_household_status"`
    UserID             uuid.UUID  `gorm:"type:uuid;not null"`
    TrackingNumber     string     `gorm:"type:varchar(64);not null"`
    Carrier            string     `gorm:"type:varchar(16);not null"`
    Label              *string    `gorm:"type:varchar(255)"`
    Notes              *string    `gorm:"type:text"`
    Status             string     `gorm:"type:varchar(24);not null;default:'pre_transit';index:idx_pkg_household_status"`
    Private            bool       `gorm:"not null;default:false"`
    EstimatedDelivery  *time.Time `gorm:"type:date"`
    ActualDelivery     *time.Time `gorm:"type:timestamptz"`
    LastPolledAt       *time.Time `gorm:"type:timestamptz"`
    LastStatusChangeAt *time.Time `gorm:"type:timestamptz"`
    ArchivedAt         *time.Time `gorm:"type:timestamptz"`
    CreatedAt          time.Time  `gorm:"type:timestamptz;not null"`
    UpdatedAt          time.Time  `gorm:"type:timestamptz;not null"`
}

type TrackingEventEntity struct {
    ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    PackageID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_te_package_time"`
    Timestamp   time.Time  `gorm:"type:timestamptz;not null;index:idx_te_package_time,sort:desc"`
    Status      string     `gorm:"type:varchar(24);not null"`
    Description string     `gorm:"type:varchar(512);not null"`
    Location    *string    `gorm:"type:varchar(255)"`
    RawStatus   *string    `gorm:"type:varchar(128)"`
    CreatedAt   time.Time  `gorm:"type:timestamptz;not null"`
}
```

---

## Migration Notes

- GORM `AutoMigrate` handles table creation on service startup (per project pattern)
- Composite unique constraint on `(tenant_id, household_id, tracking_number)` ensures no duplicate tracking within a household
- Partial indexes on `status` for polling and cleanup queries improve background job performance
- `ON DELETE CASCADE` on tracking_events ensures clean removal when a package is deleted
