package task

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a task
type RestModel struct {
	Id          uuid.UUID `json:"-"`
	UserId      string    `json:"userId"`
	HouseholdId string    `json:"householdId"`
	Day         string    `json:"day"` // Date in YYYY-MM-DD format
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   string    `json:"createdAt"`
	CompletedAt *string   `json:"completedAt,omitempty"`
	UpdatedAt   string    `json:"updatedAt"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r RestModel) GetName() string {
	return "tasks"
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
	var completedAt *string
	if m.CompletedAt() != nil {
		ca := m.CompletedAt().Format("2006-01-02T15:04:05Z07:00")
		completedAt = &ca
	}

	return RestModel{
		Id:          m.Id(),
		UserId:      m.UserId().String(),
		HouseholdId: m.HouseholdId().String(),
		Day:         m.Day().Format("2006-01-02"),
		Title:       m.Title(),
		Description: m.Description(),
		Status:      m.Status().String(),
		CreatedAt:   m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt: completedAt,
		UpdatedAt:   m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
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

// CreateRequest represents a JSON:API request to create a task
type CreateRequest struct {
	Id          uuid.UUID `json:"-"`
	Day         string    `json:"day"`         // Date in YYYY-MM-DD format
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r CreateRequest) GetName() string {
	return "tasks"
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

// UpdateRequest represents a JSON:API request to update a task
type UpdateRequest struct {
	Id          uuid.UUID `json:"-"`
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Day         *string   `json:"day,omitempty"` // Date in YYYY-MM-DD format
	Status      *string   `json:"status,omitempty"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r UpdateRequest) GetName() string {
	return "tasks"
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
