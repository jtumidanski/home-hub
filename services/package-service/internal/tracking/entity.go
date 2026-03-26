package tracking

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                 uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId           uuid.UUID  `gorm:"type:uuid;not null;index:idx_pkg_household_status"`
	HouseholdId        uuid.UUID  `gorm:"type:uuid;not null;index:idx_pkg_household_status"`
	UserId             uuid.UUID  `gorm:"type:uuid;not null"`
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

func (Entity) TableName() string { return "packages" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_pkg_household_tracking ON packages (tenant_id, household_id, tracking_number)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_pkg_polling ON packages (status, last_polled_at) WHERE status IN ('pre_transit', 'in_transit', 'out_for_delivery')").Error; err != nil {
		return err
	}
	return db.Exec("CREATE INDEX IF NOT EXISTS idx_pkg_cleanup ON packages (status, archived_at) WHERE status IN ('delivered', 'archived')").Error
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:                 m.id,
		TenantId:           m.tenantID,
		HouseholdId:        m.householdID,
		UserId:             m.userID,
		TrackingNumber:     m.trackingNumber,
		Carrier:            m.carrier,
		Label:              m.label,
		Notes:              m.notes,
		Status:             m.status,
		Private:            m.private,
		EstimatedDelivery:  m.estimatedDelivery,
		ActualDelivery:     m.actualDelivery,
		LastPolledAt:       m.lastPolledAt,
		LastStatusChangeAt: m.lastStatusChangeAt,
		ArchivedAt:         m.archivedAt,
		CreatedAt:          m.createdAt,
		UpdatedAt:          m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetUserID(e.UserId).
		SetTrackingNumber(e.TrackingNumber).
		SetCarrier(e.Carrier).
		SetLabel(e.Label).
		SetNotes(e.Notes).
		SetStatus(e.Status).
		SetPrivate(e.Private).
		SetEstimatedDelivery(e.EstimatedDelivery).
		SetActualDelivery(e.ActualDelivery).
		SetLastPolledAt(e.LastPolledAt).
		SetLastStatusChangeAt(e.LastStatusChangeAt).
		SetArchivedAt(e.ArchivedAt).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
