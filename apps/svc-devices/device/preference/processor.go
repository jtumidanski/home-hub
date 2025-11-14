package preference

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrPreferencesNotFound = errors.New("device preferences not found")
)

// UpdateInput contains the data to update device preferences
type UpdateInput struct {
	Theme           *string
	TemperatureUnit *string
}

// Processor handles business logic for device preference operations
type Processor struct {
	log logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

// NewProcessor creates a new device preferences processor with dependencies
func NewProcessor(log logrus.FieldLogger, ctx context.Context, db *gorm.DB) Processor {
	return Processor{
		log: log,
		ctx: ctx,
		db:  db,
	}
}

// GetByDeviceId retrieves preferences for a specific device
func (p Processor) GetByDeviceId(deviceId uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		model, err := GetByDeviceId(p.db)(deviceId)()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Model{}, ErrPreferencesNotFound
			}
			return Model{}, err
		}
		return model, nil
	}
}

// Upsert creates or updates preferences for a device
func (p Processor) Upsert(deviceId uuid.UUID, input UpdateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("deviceId", deviceId).Info("Upserting device preferences")

		// Try to fetch existing preferences
		existing, err := p.GetByDeviceId(deviceId)()

		var model Model
		if err != nil && !errors.Is(err, ErrPreferencesNotFound) {
			// Unexpected error
			return Model{}, err
		}

		if errors.Is(err, ErrPreferencesNotFound) {
			// Create new preferences
			p.log.WithField("deviceId", deviceId).Info("Creating new preferences")

			builder := NewBuilder().SetDeviceId(deviceId)

			if input.Theme != nil {
				builder.SetTheme(*input.Theme)
			}
			if input.TemperatureUnit != nil {
				builder.SetTemperatureUnit(*input.TemperatureUnit)
			}

			model, err = builder.Build()
			if err != nil {
				p.log.WithError(err).Error("Failed to build preferences model")
				return Model{}, err
			}

			entity := model.ToEntity()
			if err := p.db.Create(&entity).Error; err != nil {
				p.log.WithError(err).Error("Failed to create preferences in database")
				return Model{}, err
			}

			return Make(entity)
		}

		// Update existing preferences
		p.log.WithField("deviceId", deviceId).Info("Updating existing preferences")

		builder := existing.Builder().SetUpdatedAt(time.Now())

		if input.Theme != nil {
			builder.SetTheme(*input.Theme)
		}
		if input.TemperatureUnit != nil {
			builder.SetTemperatureUnit(*input.TemperatureUnit)
		}

		model, err = builder.Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build updated preferences model")
			return Model{}, err
		}

		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to update preferences in database")
			return Model{}, err
		}

		p.log.WithField("deviceId", deviceId).Info("Preferences upserted successfully")
		return Make(entity)
	}
}
