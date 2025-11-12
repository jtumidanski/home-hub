package household

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
	ErrHouseholdNotFound  = errors.New("household not found")
	ErrHouseholdHasUsers  = errors.New("cannot delete household with associated users")
)

// CreateInput contains the data needed to create a new household
type CreateInput struct {
	Name      string
	Latitude  *float64
	Longitude *float64
	Timezone  *string
}

// UpdateInput contains the data to update an existing household
type UpdateInput struct {
	Name      *string
	Latitude  *float64
	Longitude *float64
	Timezone  *string
}

// Processor handles business logic for household operations
type Processor struct {
	log logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

// NewProcessor creates a new household processor with dependencies
func NewProcessor(log logrus.FieldLogger, ctx context.Context, db *gorm.DB) Processor {
	return Processor{
		log: log,
		ctx: ctx,
		db:  db,
	}
}

// Create creates a new household with the provided input
func (p Processor) Create(input CreateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("name", input.Name).Info("Creating new household")

		// Build the model
		builder := NewBuilder().SetName(input.Name)

		// Add location coordinates if provided
		if input.Latitude != nil {
			builder.SetLatitude(*input.Latitude)
		}
		if input.Longitude != nil {
			builder.SetLongitude(*input.Longitude)
		}
		if input.Timezone != nil {
			builder.SetTimezone(*input.Timezone)
		}

		model, err := builder.Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build household model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Create(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to create household in database")
			return Model{}, err
		}

		p.log.WithField("householdId", model.Id()).Info("Household created successfully")
		return Make(entity)
	}
}

// GetById retrieves a household by ID
func (p Processor) GetById(id uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		model, err := GetById(p.db)(id)()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Model{}, ErrHouseholdNotFound
			}
			return Model{}, err
		}
		return model, nil
	}
}

// Update updates an existing household
func (p Processor) Update(id uuid.UUID, input UpdateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("householdId", id).Info("Updating household")

		// Fetch existing household
		existing, err := p.GetById(id)()
		if err != nil {
			return Model{}, err
		}

		// Build updated model
		builder := existing.Builder().
			SetUpdatedAt(time.Now())

		if input.Name != nil {
			builder.SetName(*input.Name)
		}
		if input.Latitude != nil {
			builder.SetLatitude(*input.Latitude)
		}
		if input.Longitude != nil {
			builder.SetLongitude(*input.Longitude)
		}
		if input.Timezone != nil {
			builder.SetTimezone(*input.Timezone)
		}

		model, err := builder.Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build updated household model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to update household in database")
			return Model{}, err
		}

		p.log.WithField("householdId", id).Info("Household updated successfully")
		return Make(entity)
	}
}

// Delete removes a household by ID
// This will fail if there are any users associated with the household
func (p Processor) Delete(id uuid.UUID) error {
	p.log.WithField("householdId", id).Info("Deleting household")

	// Check if household has any users
	var userCount int64
	if err := p.db.Table("users").Where("household_id = ?", id).Count(&userCount).Error; err != nil {
		p.log.WithError(err).Error("Failed to check for associated users")
		return err
	}

	if userCount > 0 {
		p.log.WithFields(logrus.Fields{
			"householdId": id,
			"userCount":   userCount,
		}).Warn("Cannot delete household with associated users")
		return ErrHouseholdHasUsers
	}

	// Delete the household
	result := p.db.Delete(&Entity{}, "id = ?", id)
	if result.Error != nil {
		p.log.WithError(result.Error).Error("Failed to delete household")
		return result.Error
	}

	if result.RowsAffected == 0 {
		p.log.WithField("householdId", id).Warn("Household not found for deletion")
		return ErrHouseholdNotFound
	}

	p.log.WithField("householdId", id).Info("Household deleted successfully")
	return nil
}
