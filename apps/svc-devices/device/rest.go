package device

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a device
type RestModel struct {
	Id          uuid.UUID `json:"-"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	HouseholdId uuid.UUID `json:"householdId"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

func (r RestModel) GetName() string {
	return "devices"
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
		Id:          m.Id(),
		Name:        m.Name(),
		Type:        m.Type(),
		HouseholdId: m.HouseholdId(),
		CreatedAt:   m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CreateRequest represents a JSON:API request to create a device
// Note: HouseholdId is NOT included in the request - it comes from auth context
type CreateRequest struct {
	Id   uuid.UUID `json:"-"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

func (r CreateRequest) GetName() string {
	return "devices"
}

func (r CreateRequest) GetID() string {
	return r.Id.String()
}

func (r *CreateRequest) SetID(idStr string) error {
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

// UpdateRequest represents a JSON:API request to update a device
type UpdateRequest struct {
	Id   uuid.UUID `json:"-"`
	Name *string   `json:"name,omitempty"`
}

func (r UpdateRequest) GetName() string {
	return "devices"
}

func (r UpdateRequest) GetID() string {
	return r.Id.String()
}

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
