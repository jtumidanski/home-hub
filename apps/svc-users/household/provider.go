package household

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/packages/shared-go/database"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"gorm.io/gorm"
)

// GetById returns a provider that fetches a household by ID
func GetById(db *gorm.DB) func(id uuid.UUID) ops.Provider[Model] {
	return func(id uuid.UUID) ops.Provider[Model] {
		return ops.Map(Make)(database.Query[Entity](db, Entity{Id: id}))
	}
}

// GetAll returns a provider that fetches all households
func GetAll(db *gorm.DB) ops.Provider[[]Model] {
	return ops.SliceMap(Make)(database.SliceQuery[Entity](db, Entity{}))(ops.ParallelMap())
}

// Count returns a provider that counts total households
func Count(db *gorm.DB) ops.Provider[int64] {
	return func() (int64, error) {
		var count int64
		err := db.Model(&Entity{}).Count(&count).Error
		return count, err
	}
}
