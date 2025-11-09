package user

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/packages/shared-go/database"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"gorm.io/gorm"
)

// GetById returns a provider that fetches a user by ID
func GetById(db *gorm.DB) func(id uuid.UUID) ops.Provider[Model] {
	return func(id uuid.UUID) ops.Provider[Model] {
		return ops.Map(Make)(database.Query[Entity](db, Entity{Id: id}))
	}
}

// GetAll returns a provider that fetches all users
func GetAll(db *gorm.DB) ops.Provider[[]Model] {
	return ops.SliceMap(Make)(database.SliceQuery[Entity](db, Entity{}))(ops.ParallelMap())
}

// GetByEmail returns a provider that fetches a user by email address
func GetByEmail(db *gorm.DB) func(email string) ops.Provider[Model] {
	return func(email string) ops.Provider[Model] {
		return ops.Map(Make)(database.Query[Entity](db, Entity{Email: email}))
	}
}

// GetByHouseholdId returns a provider that fetches all users in a household
func GetByHouseholdId(db *gorm.DB) func(householdId uuid.UUID) ops.Provider[[]Model] {
	return func(householdId uuid.UUID) ops.Provider[[]Model] {
		return ops.SliceMap(Make)(database.SliceQuery[Entity](db, Entity{HouseholdId: &householdId}))(ops.ParallelMap())
	}
}

// GetUsersWithoutHousehold returns a provider that fetches all users not associated with any household
func GetUsersWithoutHousehold(db *gorm.DB) ops.Provider[[]Model] {
	return func() ([]Model, error) {
		var entities []Entity
		err := db.Where("household_id IS NULL").Find(&entities).Error
		if err != nil {
			return nil, err
		}

		models := make([]Model, len(entities))
		for i, entity := range entities {
			model, err := Make(entity)
			if err != nil {
				return nil, err
			}
			models[i] = model
		}
		return models, nil
	}
}
