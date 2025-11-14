package device

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/apps/svc-devices/device/preference"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrDeviceNotFound = errors.New("device not found")
)

// CreateInput contains the data needed to create a new device
type CreateInput struct {
	Name        string
	Type        string
	HouseholdId uuid.UUID
}

// UpdateInput contains the data to update an existing device
type UpdateInput struct {
	Name *string
}

// Processor handles business logic for device operations
type Processor struct {
	log logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

// NewProcessor creates a new device processor with dependencies
func NewProcessor(log logrus.FieldLogger, ctx context.Context, db *gorm.DB) Processor {
	return Processor{
		log: log,
		ctx: ctx,
		db:  db,
	}
}

// Create creates a new device with default preferences
func (p Processor) Create(input CreateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("name", input.Name).Info("Creating new device")

		// Build the device model
		model, err := NewBuilder().
			SetName(input.Name).
			SetType(input.Type).
			SetHouseholdId(input.HouseholdId).
			Build()

		if err != nil {
			p.log.WithError(err).Error("Failed to build device model")
			return Model{}, err
		}

		// Save device to database
		entity := model.ToEntity()
		if err := p.db.Create(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to create device in database")
			return Model{}, err
		}

		// Create default preferences for the device
		prefModel, err := preference.NewBuilder().
			SetDeviceId(model.Id()).
			Build() // Uses defaults: theme=dark, tempUnit=household

		if err != nil {
			p.log.WithError(err).Error("Failed to build default preferences")
			// Note: device was created, but preferences failed
			// We could rollback here or continue
			return Make(entity)
		}

		prefEntity := prefModel.ToEntity()
		if err := p.db.Create(&prefEntity).Error; err != nil {
			p.log.WithError(err).Error("Failed to create default preferences")
			// Note: device exists without preferences - could be fixed later
		}

		p.log.WithField("deviceId", model.Id()).Info("Device created successfully")
		return Make(entity)
	}
}

// GetById retrieves a device by ID
func (p Processor) GetById(id uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		model, err := GetById(p.db)(id)()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Model{}, ErrDeviceNotFound
			}
			return Model{}, err
		}
		return model, nil
	}
}

// GetByHousehold retrieves all devices for a household
func (p Processor) GetByHousehold(householdId uuid.UUID) ops.Provider[[]Model] {
	return GetByHousehold(p.db)(householdId)
}

// Update updates an existing device
func (p Processor) Update(id uuid.UUID, input UpdateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("deviceId", id).Info("Updating device")

		// Fetch existing device
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

		model, err := builder.Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build updated device model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to update device in database")
			return Model{}, err
		}

		p.log.WithField("deviceId", id).Info("Device updated successfully")
		return Make(entity)
	}
}

// Delete removes a device (preferences will cascade delete via DB constraint)
func (p Processor) Delete(id uuid.UUID) ops.Provider[bool] {
	return func() (bool, error) {
		p.log.WithField("deviceId", id).Info("Deleting device")

		// Verify device exists
		_, err := p.GetById(id)()
		if err != nil {
			return false, err
		}

		// Delete the device (preferences cascade automatically)
		if err := p.db.Delete(&Entity{}, "id = ?", id).Error; err != nil {
			p.log.WithError(err).Error("Failed to delete device from database")
			return false, err
		}

		p.log.WithField("deviceId", id).Info("Device deleted successfully")
		return true, nil
	}
}
