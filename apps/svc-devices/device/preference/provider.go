package preference

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/packages/shared-go/database"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"gorm.io/gorm"
)

// GetByDeviceId returns a provider that fetches preferences for a specific device
func GetByDeviceId(db *gorm.DB) func(deviceId uuid.UUID) ops.Provider[Model] {
	return func(deviceId uuid.UUID) ops.Provider[Model] {
		return ops.Map(Make)(database.Query[Entity](db, Entity{DeviceId: deviceId}))
	}
}

// GetById returns a provider that fetches preferences by ID
func GetById(db *gorm.DB) func(id uuid.UUID) ops.Provider[Model] {
	return func(id uuid.UUID) ops.Provider[Model] {
		return ops.Map(Make)(database.Query[Entity](db, Entity{Id: id}))
	}
}
