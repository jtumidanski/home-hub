package reminder

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/packages/shared-go/database"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"gorm.io/gorm"
)

// GetById returns a provider that fetches a reminder by ID
func GetById(db *gorm.DB) func(id uuid.UUID) ops.Provider[Model] {
	return func(id uuid.UUID) ops.Provider[Model] {
		return ops.Map(Make)(database.Query[Entity](db, Entity{Id: id}))
	}
}

// GetByUserId returns a provider that fetches all reminders for a user
func GetByUserId(db *gorm.DB) func(userId uuid.UUID) ops.Provider[[]Model] {
	return func(userId uuid.UUID) ops.Provider[[]Model] {
		return ops.SliceMap(Make)(database.SliceQuery[Entity](db, Entity{UserId: userId}))(ops.ParallelMap())
	}
}

// GetByUserIdAndStatus returns a provider that fetches all reminders for a user with a specific status
func GetByUserIdAndStatus(db *gorm.DB) func(userId uuid.UUID, status Status) ops.Provider[[]Model] {
	return func(userId uuid.UUID, status Status) ops.Provider[[]Model] {
		return ops.SliceMap(Make)(database.SliceQuery[Entity](db, Entity{UserId: userId, Status: string(status)}))(ops.ParallelMap())
	}
}

// GetByHouseholdId returns a provider that fetches all reminders for a household
func GetByHouseholdId(db *gorm.DB) func(householdId uuid.UUID) ops.Provider[[]Model] {
	return func(householdId uuid.UUID) ops.Provider[[]Model] {
		return ops.SliceMap(Make)(database.SliceQuery[Entity](db, Entity{HouseholdId: householdId}))(ops.ParallelMap())
	}
}

// GetOverdueForDismissal returns a provider that fetches all overdue reminders (more than 24 hours past remind_at)
// This is used by the sweeper to auto-dismiss overdue reminders
func GetOverdueForDismissal(db *gorm.DB) ops.Provider[[]Model] {
	return func() ([]Model, error) {
		var entities []Entity
		cutoff := time.Now().Add(-24 * time.Hour)

		err := db.Where("(status = ? OR status = ?) AND remind_at < ?",
			string(StatusActive), string(StatusSnoozed), cutoff).
			Order("remind_at ASC").
			Find(&entities).Error

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

// Count returns a provider that counts total reminders for a user
func Count(db *gorm.DB) func(userId uuid.UUID) ops.Provider[int64] {
	return func(userId uuid.UUID) ops.Provider[int64] {
		return func() (int64, error) {
			var count int64
			err := db.Model(&Entity{}).Where("user_id = ?", userId).Count(&count).Error
			return count, err
		}
	}
}

// CountByStatus returns a provider that counts reminders by status for a user
func CountByStatus(db *gorm.DB) func(userId uuid.UUID, status Status) ops.Provider[int64] {
	return func(userId uuid.UUID, status Status) ops.Provider[int64] {
		return func() (int64, error) {
			var count int64
			err := db.Model(&Entity{}).
				Where("user_id = ? AND status = ?", userId, string(status)).
				Count(&count).Error
			return count, err
		}
	}
}
