package trackingevent

import (
	"time"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByPackageID(packageID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("package_id = ?", packageID).Order("timestamp DESC")
	})
}

func Create(db *gorm.DB, packageID uuid.UUID, timestamp time.Time, status, description string, location, rawStatus *string) error {
	// Deduplicate: skip if an event with the same package, timestamp, and description exists
	var count int64
	db.Model(&Entity{}).
		Where("package_id = ? AND timestamp = ? AND description = ?", packageID, timestamp, description).
		Count(&count)
	if count > 0 {
		return nil
	}

	e := Entity{
		Id:          uuid.New(),
		PackageId:   packageID,
		Timestamp:   timestamp,
		Status:      status,
		Description: description,
		Location:    location,
		RawStatus:   rawStatus,
		CreatedAt:   time.Now().UTC(),
	}
	return db.Create(&e).Error
}
