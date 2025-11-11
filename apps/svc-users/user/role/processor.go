package role

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRoleAlreadyAssigned = errors.New("role already assigned to user")
	ErrRoleNotAssigned     = errors.New("role not assigned to user")
)

// Processor represents a pure business logic function that produces a Model
type Processor func(*gorm.DB) (Model, error)

// AssignRole assigns a role to a user
// Returns the created role model or error if already assigned
func AssignRole(userId uuid.UUID, roleName string) Processor {
	return func(db *gorm.DB) (Model, error) {
		// Check if role already assigned
		var count int64
		err := db.Model(&Entity{}).
			Where("user_id = ? AND role = ?", userId, roleName).
			Count(&count).Error
		if err != nil {
			return Model{}, err
		}
		if count > 0 {
			return Model{}, ErrRoleAlreadyAssigned
		}

		// Build and validate the model
		model, err := NewBuilder().
			SetUserId(userId).
			SetRole(roleName).
			Build()
		if err != nil {
			return Model{}, err
		}

		// Persist to database
		entity := model.ToEntity()
		if err := db.Create(&entity).Error; err != nil {
			return Model{}, err
		}

		return model, nil
	}
}

// RemoveRole removes a role from a user
// Returns error if role is not currently assigned
func RemoveRole(userId uuid.UUID, roleName string) func(*gorm.DB) error {
	return func(db *gorm.DB) error {
		// Check if role is assigned
		var count int64
		err := db.Model(&Entity{}).
			Where("user_id = ? AND role = ?", userId, roleName).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count == 0 {
			return ErrRoleNotAssigned
		}

		// Delete the role assignment
		return db.Where("user_id = ? AND role = ?", userId, roleName).
			Delete(&Entity{}).Error
	}
}

// RemoveAllRoles removes all role assignments for a user
func RemoveAllRoles(userId uuid.UUID) func(*gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.Where("user_id = ?", userId).Delete(&Entity{}).Error
	}
}

// EnsureDefaultRoles ensures a user has the default "user" role
// This should be called when a user is first created
func EnsureDefaultRoles(userId uuid.UUID) func(*gorm.DB) error {
	return func(db *gorm.DB) error {
		// Check if user already has roles
		var count int64
		err := db.Model(&Entity{}).Where("user_id = ?", userId).Count(&count).Error
		if err != nil {
			return err
		}

		// If no roles, assign default "user" role
		if count == 0 {
			_, err := AssignRole(userId, User)(db)
			return err
		}

		return nil
	}
}
