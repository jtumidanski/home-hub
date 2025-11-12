package preference

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a user preference
type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

// GetName returns the resource type name for JSON:API
func (r RestModel) GetName() string {
	return "preferences"
}

// GetID returns the string ID for JSON:API
func (r RestModel) GetID() string {
	return r.Id.String()
}

// SetID sets the ID from a string for JSON:API
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
		Id:        m.Id(),
		Key:       m.Key(),
		Value:     m.Value(),
		CreatedAt: m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// TransformSlice converts a slice of domain Models to REST representations
func TransformSlice(models []Model) ([]RestModel, error) {
	restModels := make([]RestModel, len(models))
	for i, model := range models {
		restModel, err := Transform(model)
		if err != nil {
			return nil, err
		}
		restModels[i] = restModel
	}
	return restModels, nil
}

// UpdateRequest represents a JSON:API request to update a preference
type UpdateRequest struct {
	Id    uuid.UUID `json:"-"`
	Key   string    `json:"key,omitempty"`
	Value string    `json:"value"`
}

// GetName returns the resource type name for JSON:API
func (r UpdateRequest) GetName() string {
	return "preferences"
}

// GetID returns the string ID for JSON:API
func (r UpdateRequest) GetID() string {
	return r.Id.String()
}

// SetID sets the ID from a string for JSON:API
func (r *UpdateRequest) SetID(idStr string) error {
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
