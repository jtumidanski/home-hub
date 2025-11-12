package preference

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Processor encapsulates business logic for preference operations
type Processor struct {
	provider Provider
}

// NewProcessor creates a new preference processor
func NewProcessor(provider Provider) *Processor {
	return &Processor{provider: provider}
}

// GetUserPreference retrieves a specific preference for a user
// Returns ErrNotFound if the preference doesn't exist
func (p *Processor) GetUserPreference(userId uuid.UUID, key string) (Model, error) {
	if userId == uuid.Nil {
		return Model{}, fmt.Errorf("userId is required")
	}

	if key == "" {
		return Model{}, fmt.Errorf("key is required")
	}

	preference, err := p.provider.FindByUserIdAndKey(userId, key)
	if err != nil {
		return Model{}, err
	}

	return preference, nil
}

// GetAllUserPreferences retrieves all preferences for a user
func (p *Processor) GetAllUserPreferences(userId uuid.UUID) ([]Model, error) {
	if userId == uuid.Nil {
		return nil, fmt.Errorf("userId is required")
	}

	preferences, err := p.provider.FindAllByUserId(userId)
	if err != nil {
		return nil, err
	}

	return preferences, nil
}

// GetUserPreferencesMap retrieves all preferences for a user as a map[key]value
// This is useful for including preferences in the /me endpoint
func (p *Processor) GetUserPreferencesMap(userId uuid.UUID) (map[string]string, error) {
	preferences, err := p.GetAllUserPreferences(userId)
	if err != nil {
		return nil, err
	}

	// Convert to map
	result := make(map[string]string, len(preferences))
	for _, pref := range preferences {
		result[pref.Key()] = pref.Value()
	}

	return result, nil
}

// SetUserPreference creates or updates a preference for a user
// Returns the created/updated preference
func (p *Processor) SetUserPreference(userId uuid.UUID, key string, value string) (Model, error) {
	if userId == uuid.Nil {
		return Model{}, fmt.Errorf("userId is required")
	}

	// Try to find existing preference
	existing, err := p.provider.FindByUserIdAndKey(userId, key)

	var preference Model
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Create new preference
			preference, err = NewBuilder().
				ForUser(userId).
				WithKey(key).
				WithValue(value).
				Build()
			if err != nil {
				return Model{}, fmt.Errorf("failed to build preference: %w", err)
			}
		} else {
			return Model{}, fmt.Errorf("failed to check existing preference: %w", err)
		}
	} else {
		// Update existing preference
		preference, err = UpdateValue(existing, value)
		if err != nil {
			return Model{}, fmt.Errorf("failed to update preference: %w", err)
		}
	}

	// Save to database
	if err := p.provider.Save(preference); err != nil {
		return Model{}, fmt.Errorf("failed to save preference: %w", err)
	}

	return preference, nil
}

// DeleteUserPreference removes a preference for a user
func (p *Processor) DeleteUserPreference(userId uuid.UUID, key string) error {
	if userId == uuid.Nil {
		return fmt.Errorf("userId is required")
	}

	if key == "" {
		return fmt.Errorf("key is required")
	}

	if err := p.provider.Delete(userId, key); err != nil {
		return err
	}

	return nil
}

// GetThemePreference is a convenience method to get the theme preference
// Returns the default theme (system) if not set
func (p *Processor) GetThemePreference(userId uuid.UUID) (string, error) {
	preference, err := p.GetUserPreference(userId, THEME_KEY)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Return default theme if not set
			return THEME_SYSTEM, nil
		}
		return "", err
	}

	return preference.Value(), nil
}

// SetThemePreference is a convenience method to set the theme preference
func (p *Processor) SetThemePreference(userId uuid.UUID, theme string) (Model, error) {
	return p.SetUserPreference(userId, THEME_KEY, theme)
}
