package preference

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of device preferences
type RestModel struct {
	Id              uuid.UUID `json:"-"`
	DeviceId        uuid.UUID `json:"deviceId"`
	Theme           string    `json:"theme"`
	TemperatureUnit string    `json:"temperatureUnit"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

func (r RestModel) GetName() string {
	return "device_preferences"
}

func (r RestModel) GetID() string {
	return r.Id.String()
}

func (r *RestModel) SetID(idStr string) error {
	if idStr == "" {
		r.Id = uuid.Nil
		return nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}

// Transform converts a domain Model to a REST representation
func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:              m.Id(),
		DeviceId:        m.DeviceId(),
		Theme:           m.Theme(),
		TemperatureUnit: m.TemperatureUnit(),
		CreatedAt:       m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// UpdatePreferencesRequest represents a JSON:API request to update device preferences
type UpdatePreferencesRequest struct {
	Id              uuid.UUID `json:"-"`
	Theme           *string   `json:"theme,omitempty"`
	TemperatureUnit *string   `json:"temperatureUnit,omitempty"`
}

func (r UpdatePreferencesRequest) GetName() string {
	return "device_preferences"
}

func (r UpdatePreferencesRequest) GetID() string {
	return r.Id.String()
}

func (r *UpdatePreferencesRequest) SetID(idStr string) error {
	if idStr == "" {
		r.Id = uuid.Nil
		return nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}
