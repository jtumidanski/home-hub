package household

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a household
type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func (r *RestModel) GetName() string {
	return "households"
}

func (r RestModel) GetID() string {
	return r.Id.String()
}

func (r *RestModel) SetID(idStr string) error {
	// Handle empty string gracefully
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
		Name:      m.Name(),
		CreatedAt: m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CreateRequest represents a JSON:API request to create a household
type CreateRequest struct {
	Id   uuid.UUID `json:"-"`
	Name string    `json:"name"`
}

func (r *CreateRequest) GetName() string {
	return "households"
}

func (r CreateRequest) GetID() string {
	return r.Id.String()
}

func (r *CreateRequest) SetID(idStr string) error {
	// For create requests, ID is optional (server generates it)
	// Empty string is valid and will result in uuid.Nil
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

// UpdateRequest represents a JSON:API request to update a household
type UpdateRequest struct {
	Id   uuid.UUID `json:"-"`
	Name *string   `json:"name,omitempty"`
}

func (r *UpdateRequest) GetName() string {
	return "households"
}

func (r UpdateRequest) GetID() string {
	return r.Id.String()
}

func (r *UpdateRequest) SetID(idStr string) error {
	// For update requests, ID typically comes from URL path
	// Empty string is valid and will result in uuid.Nil
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
