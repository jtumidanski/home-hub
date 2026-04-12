package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Trigger identifies why a reaper run happened.
type Trigger string

const (
	TriggerScheduled Trigger = "scheduled"
	TriggerManual    Trigger = "manual"
)

// RunEntity is the GORM model for the per-service retention_runs audit table.
// Each reaper-owning service migrates its own copy of this table.
type RunEntity struct {
	Id         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId   uuid.UUID  `gorm:"type:uuid;not null;index:idx_retention_runs_tenant_started"`
	ScopeKind  string     `gorm:"type:text;not null"`
	ScopeId    uuid.UUID  `gorm:"type:uuid;not null"`
	Category   string     `gorm:"type:text;not null;index:idx_retention_runs_category_started"`
	Trigger    string     `gorm:"type:text;not null"`
	DryRun     bool       `gorm:"not null;default:false"`
	Scanned    int        `gorm:"not null;default:0"`
	Deleted    int        `gorm:"not null;default:0"`
	StartedAt  time.Time  `gorm:"not null;index:idx_retention_runs_tenant_started,priority:2,sort:desc;index:idx_retention_runs_category_started,priority:2,sort:desc"`
	FinishedAt *time.Time `gorm:""`
	Error      *string    `gorm:"type:text"`
}

// TableName forces the GORM table name regardless of struct package.
func (RunEntity) TableName() string { return "retention_runs" }

// MigrateRuns runs AutoMigrate for retention_runs in the calling service.
func MigrateRuns(db *gorm.DB) error { return db.AutoMigrate(&RunEntity{}) }

// RunRecord is the application-level value used to write a single audit row.
type RunRecord struct {
	Id         uuid.UUID
	TenantId   uuid.UUID
	ScopeKind  ScopeKind
	ScopeId    uuid.UUID
	Category   Category
	Trigger    Trigger
	DryRun     bool
	Scanned    int
	Deleted    int
	StartedAt  time.Time
	FinishedAt *time.Time
	Error      string
}

// WriteRun persists a RunRecord to retention_runs. The write deliberately
// bypasses the tenant filter because the reaper runs cross-tenant.
func WriteRun(ctx context.Context, db *gorm.DB, r RunRecord) error {
	if r.Id == uuid.Nil {
		r.Id = uuid.New()
	}
	var errPtr *string
	if r.Error != "" {
		errPtr = &r.Error
	}
	e := RunEntity{
		Id:         r.Id,
		TenantId:   r.TenantId,
		ScopeKind:  string(r.ScopeKind),
		ScopeId:    r.ScopeId,
		Category:   string(r.Category),
		Trigger:    string(r.Trigger),
		DryRun:     r.DryRun,
		Scanned:    r.Scanned,
		Deleted:    r.Deleted,
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
		Error:      errPtr,
	}
	return db.WithContext(ctx).Create(&e).Error
}
