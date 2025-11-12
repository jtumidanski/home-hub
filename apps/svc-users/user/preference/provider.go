package preference

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrNotFound is returned when a preference is not found
	ErrNotFound = errors.New("preference not found")
	// ErrDuplicateKey is returned when trying to create a duplicate preference
	ErrDuplicateKey = errors.New("preference with this key already exists for user")
)

// Provider defines the interface for preference repository operations
type Provider interface {
	// FindByUserIdAndKey finds a preference by user ID and key
	FindByUserIdAndKey(userId uuid.UUID, key string) (Model, error)

	// FindAllByUserId finds all preferences for a user
	FindAllByUserId(userId uuid.UUID) ([]Model, error)

	// Save creates or updates a preference (upsert behavior)
	Save(preference Model) error

	// Delete removes a preference by user ID and key
	Delete(userId uuid.UUID, key string) error
}

// GormProvider implements Provider using GORM
type GormProvider struct {
	db *gorm.DB
}

// NewGormProvider creates a new GORM-based preference provider
func NewGormProvider(db *gorm.DB) Provider {
	return &GormProvider{db: db}
}

// FindByUserIdAndKey finds a preference by user ID and key
func (p *GormProvider) FindByUserIdAndKey(userId uuid.UUID, key string) (Model, error) {
	var entity Entity
	err := p.db.Where("user_id = ? AND key = ?", userId, key).First(&entity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Model{}, ErrNotFound
		}
		return Model{}, fmt.Errorf("failed to find preference: %w", err)
	}

	return Make(entity)
}

// FindAllByUserId finds all preferences for a user
func (p *GormProvider) FindAllByUserId(userId uuid.UUID) ([]Model, error) {
	var entities []Entity
	err := p.db.Where("user_id = ?", userId).Order("key ASC").Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find preferences: %w", err)
	}

	// Convert entities to models
	models := make([]Model, len(entities))
	for i, entity := range entities {
		model, err := Make(entity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity to model: %w", err)
		}
		models[i] = model
	}

	return models, nil
}

// Save creates or updates a preference (upsert behavior)
// Uses GORM's Save method which performs an INSERT or UPDATE based on primary key
func (p *GormProvider) Save(preference Model) error {
	entity := preference.ToEntity()

	// Check if a preference with this user_id and key already exists
	var existing Entity
	err := p.db.Where("user_id = ? AND key = ?", entity.UserId, entity.Key).First(&existing).Error

	if err == nil {
		// Preference exists, update it with the new ID to maintain the same record
		entity.Id = existing.Id
		entity.CreatedAt = existing.CreatedAt
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Unexpected error
		return fmt.Errorf("failed to check existing preference: %w", err)
	}

	// Save (insert or update)
	if err := p.db.Save(&entity).Error; err != nil {
		return fmt.Errorf("failed to save preference: %w", err)
	}

	return nil
}

// Delete removes a preference by user ID and key
func (p *GormProvider) Delete(userId uuid.UUID, key string) error {
	result := p.db.Where("user_id = ? AND key = ?", userId, key).Delete(&Entity{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete preference: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
