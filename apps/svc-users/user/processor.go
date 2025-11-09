package user

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrHouseholdNotFound = errors.New("household not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// CreateInput contains the data needed to create a new user
type CreateInput struct {
	Email       string
	DisplayName string
	HouseholdId *uuid.UUID
}

// UpdateInput contains the data to update an existing user
type UpdateInput struct {
	Email       *string
	DisplayName *string
}

// Processor handles business logic for user operations
type Processor struct {
	log logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

// NewProcessor creates a new user processor with dependencies
func NewProcessor(log logrus.FieldLogger, ctx context.Context, db *gorm.DB) Processor {
	return Processor{
		log: log,
		ctx: ctx,
		db:  db,
	}
}

// Create creates a new user with the provided input
func (p Processor) Create(input CreateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithFields(logrus.Fields{
			"email":       input.Email,
			"displayName": input.DisplayName,
		}).Info("Creating new user")

		// Build the model
		builder := NewBuilder().
			SetEmail(input.Email).
			SetDisplayName(input.DisplayName)

		if input.HouseholdId != nil {
			// Verify household exists
			var count int64
			if err := p.db.Table("households").Where("id = ?", input.HouseholdId).Count(&count).Error; err != nil {
				p.log.WithError(err).Error("Failed to check household existence")
				return Model{}, err
			}
			if count == 0 {
				p.log.WithField("householdId", input.HouseholdId).Warn("Household not found")
				return Model{}, ErrHouseholdNotFound
			}
			builder.SetHouseholdId(*input.HouseholdId)
		}

		model, err := builder.Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build user model")
			return Model{}, err
		}

		// Check if email already exists
		var existingCount int64
		if err := p.db.Model(&Entity{}).Where("email = ?", model.Email()).Count(&existingCount).Error; err != nil {
			p.log.WithError(err).Error("Failed to check email uniqueness")
			return Model{}, err
		}
		if existingCount > 0 {
			p.log.WithField("email", model.Email()).Warn("Email already exists")
			return Model{}, ErrEmailAlreadyExists
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Create(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to create user in database")
			return Model{}, err
		}

		p.log.WithField("userId", model.Id()).Info("User created successfully")
		return Make(entity)
	}
}

// GetById retrieves a user by ID
func (p Processor) GetById(id uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		model, err := GetById(p.db)(id)()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Model{}, ErrUserNotFound
			}
			return Model{}, err
		}
		return model, nil
	}
}

// Update updates an existing user
func (p Processor) Update(id uuid.UUID, input UpdateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("userId", id).Info("Updating user")

		// Fetch existing user
		existing, err := p.GetById(id)()
		if err != nil {
			return Model{}, err
		}

		// Build updated model
		builder := existing.Builder().
			SetUpdatedAt(time.Now())

		if input.Email != nil {
			// Check if new email already exists (for a different user)
			if *input.Email != existing.Email() {
				var existingCount int64
				if err := p.db.Model(&Entity{}).
					Where("email = ? AND id != ?", *input.Email, id).
					Count(&existingCount).Error; err != nil {
					p.log.WithError(err).Error("Failed to check email uniqueness")
					return Model{}, err
				}
				if existingCount > 0 {
					p.log.WithField("email", *input.Email).Warn("Email already exists")
					return Model{}, ErrEmailAlreadyExists
				}
			}
			builder.SetEmail(*input.Email)
		}

		if input.DisplayName != nil {
			builder.SetDisplayName(*input.DisplayName)
		}

		model, err := builder.Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build updated user model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to update user in database")
			return Model{}, err
		}

		p.log.WithField("userId", id).Info("User updated successfully")
		return Make(entity)
	}
}

// Delete removes a user by ID
func (p Processor) Delete(id uuid.UUID) error {
	p.log.WithField("userId", id).Info("Deleting user")

	result := p.db.Delete(&Entity{}, "id = ?", id)
	if result.Error != nil {
		p.log.WithError(result.Error).Error("Failed to delete user")
		return result.Error
	}

	if result.RowsAffected == 0 {
		p.log.WithField("userId", id).Warn("User not found for deletion")
		return ErrUserNotFound
	}

	p.log.WithField("userId", id).Info("User deleted successfully")
	return nil
}

// AssociateHousehold associates a user with a household
func (p Processor) AssociateHousehold(userId, householdId uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithFields(logrus.Fields{
			"userId":      userId,
			"householdId": householdId,
		}).Info("Associating user with household")

		// Verify household exists
		var count int64
		if err := p.db.Table("households").Where("id = ?", householdId).Count(&count).Error; err != nil {
			p.log.WithError(err).Error("Failed to check household existence")
			return Model{}, err
		}
		if count == 0 {
			p.log.WithField("householdId", householdId).Warn("Household not found")
			return Model{}, ErrHouseholdNotFound
		}

		// Fetch existing user
		existing, err := p.GetById(userId)()
		if err != nil {
			return Model{}, err
		}

		// Update with household association
		model, err := existing.Builder().
			SetHouseholdId(householdId).
			SetUpdatedAt(time.Now()).
			Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build user model with household")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to associate user with household")
			return Model{}, err
		}

		p.log.WithFields(logrus.Fields{
			"userId":      userId,
			"householdId": householdId,
		}).Info("User associated with household successfully")
		return Make(entity)
	}
}

// DisassociateHousehold removes the household association from a user
func (p Processor) DisassociateHousehold(userId uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("userId", userId).Info("Disassociating user from household")

		// Fetch existing user
		existing, err := p.GetById(userId)()
		if err != nil {
			return Model{}, err
		}

		// Update without household
		model, err := existing.Builder().
			ClearHouseholdId().
			SetUpdatedAt(time.Now()).
			Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build user model without household")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to disassociate user from household")
			return Model{}, err
		}

		p.log.WithField("userId", userId).Info("User disassociated from household successfully")
		return Make(entity)
	}
}
