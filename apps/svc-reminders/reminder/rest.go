package reminder

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a reminder
type RestModel struct {
	Id          uuid.UUID `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	UserId      string    `json:"userId"`
	HouseholdId string    `json:"householdId"`
	CreatedAt   string    `json:"createdAt"`
	RemindAt    string    `json:"remindAt"`
	SnoozeCount int       `json:"snoozeCount"`
	Status      string    `json:"status"`
	DismissedAt *string   `json:"dismissedAt,omitempty"`
	UpdatedAt   string    `json:"updatedAt"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r RestModel) GetName() string {
	return "reminders"
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
	var dismissedAt *string
	if m.DismissedAt() != nil {
		da := m.DismissedAt().Format("2006-01-02T15:04:05Z07:00")
		dismissedAt = &da
	}

	return RestModel{
		Id:          m.Id(),
		Name:        m.Name(),
		Description: m.Description(),
		UserId:      m.UserId().String(),
		HouseholdId: m.HouseholdId().String(),
		CreatedAt:   m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		RemindAt:    m.RemindAt().Format("2006-01-02T15:04:05Z07:00"),
		SnoozeCount: m.SnoozeCount(),
		Status:      m.Status().String(),
		DismissedAt: dismissedAt,
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

// CreateRequest represents a JSON:API request to create a reminder
type CreateRequest struct {
	Id          uuid.UUID `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	RemindAt    string    `json:"remindAt"` // ISO 8601 format
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r CreateRequest) GetName() string {
	return "reminders"
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

// UpdateRequest represents a JSON:API request to update a reminder
type UpdateRequest struct {
	Id          uuid.UUID `json:"-"`
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	RemindAt    *string   `json:"remindAt,omitempty"` // ISO 8601 format
	Status      *string   `json:"status,omitempty"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r UpdateRequest) GetName() string {
	return "reminders"
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

// SnoozeRequest represents a JSON:API request to snooze a reminder
type SnoozeRequest struct {
	Id       uuid.UUID `json:"-"`
	RemindAt string    `json:"remindAt"` // ISO 8601 format
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r SnoozeRequest) GetName() string {
	return "reminders"
}

func (r SnoozeRequest) GetID() string {
	return r.Id.String()
}

func (r *SnoozeRequest) SetID(idStr string) error {
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
