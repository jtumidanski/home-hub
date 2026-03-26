package trackingevent

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	PackageId   uuid.UUID `gorm:"type:uuid;not null;index:idx_te_package_time"`
	Timestamp   time.Time `gorm:"not null;index:idx_te_package_time,sort:desc"`
	Status      string    `gorm:"type:varchar(24);not null"`
	Description string    `gorm:"type:varchar(512);not null"`
	Location    *string   `gorm:"type:varchar(255)"`
	RawStatus   *string   `gorm:"type:varchar(128)"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "tracking_events" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	// Remove existing duplicates before adding the unique index
	db.Exec(`DELETE FROM tracking_events a USING tracking_events b
		WHERE a.id > b.id
		AND a.package_id = b.package_id
		AND a.timestamp = b.timestamp
		AND a.description = b.description`)
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_te_dedup ON tracking_events (package_id, timestamp, description)").Error
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		PackageId:   m.packageID,
		Timestamp:   m.timestamp,
		Status:      m.status,
		Description: m.description,
		Location:    m.location,
		RawStatus:   m.rawStatus,
		CreatedAt:   m.createdAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetPackageID(e.PackageId).
		SetTimestamp(e.Timestamp).
		SetStatus(e.Status).
		SetDescription(e.Description).
		SetLocation(e.Location).
		SetRawStatus(e.RawStatus).
		SetCreatedAt(e.CreatedAt).
		Build()
}
