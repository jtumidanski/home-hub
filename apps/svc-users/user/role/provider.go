package role

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Provider represents a lazy data access function for user roles
type Provider func(*gorm.DB) ([]Model, error)

// GetByUserId returns a provider that fetches all roles for a given user
func GetByUserId(userId uuid.UUID) Provider {
	return func(db *gorm.DB) ([]Model, error) {
		var entities []Entity
		if err := db.Where("user_id = ?", userId).Find(&entities).Error; err != nil {
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

// GetByUserIdAndRole returns a provider that fetches a specific role for a user
func GetByUserIdAndRole(userId uuid.UUID, roleName string) Provider {
	return func(db *gorm.DB) ([]Model, error) {
		var entity Entity
		err := db.Where("user_id = ? AND role = ?", userId, roleName).First(&entity).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return []Model{}, nil
			}
			return nil, err
		}

		model, err := Make(entity)
		if err != nil {
			return nil, err
		}

		return []Model{model}, nil
	}
}

// GetAll returns a provider that fetches all user roles (admin use only)
func GetAll() Provider {
	return func(db *gorm.DB) ([]Model, error) {
		var entities []Entity
		if err := db.Find(&entities).Error; err != nil {
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

// HasRole checks if a user has a specific role
func HasRole(userId uuid.UUID, roleName string) func(*gorm.DB) (bool, error) {
	return func(db *gorm.DB) (bool, error) {
		var count int64
		err := db.Model(&Entity{}).
			Where("user_id = ? AND role = ?", userId, roleName).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}
}

// HasAnyRole checks if a user has any of the specified roles
func HasAnyRole(userId uuid.UUID, roleNames []string) func(*gorm.DB) (bool, error) {
	return func(db *gorm.DB) (bool, error) {
		if len(roleNames) == 0 {
			return false, nil
		}

		var count int64
		err := db.Model(&Entity{}).
			Where("user_id = ? AND role IN ?", userId, roleNames).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}
}
