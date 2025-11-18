package meal

import (
	"github.com/google/uuid"
)

// RestModel represents the JSON:API representation of a meal
type RestModel struct {
	Id                uuid.UUID `json:"-"`
	HouseholdId       string    `json:"householdId"`
	UserId            string    `json:"userId"`
	Title             string    `json:"title"`
	Description       string    `json:"description,omitempty"`
	RawIngredientText string    `json:"rawIngredientText"`
	CreatedAt         string    `json:"createdAt"`
	UpdatedAt         string    `json:"updatedAt"`
}

// GetName returns the resource type name for JSON:API
// Note: Value receiver (not pointer) as per api2go interface requirements
func (r RestModel) GetName() string {
	return "meals"
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
		Id:                m.Id(),
		HouseholdId:       m.HouseholdId().String(),
		UserId:            m.UserId().String(),
		Title:             m.Title(),
		Description:       m.Description(),
		RawIngredientText: m.RawIngredientText(),
		CreatedAt:         m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
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

// CreateRequest represents a JSON:API request to create a meal
type CreateRequest struct {
	Id             uuid.UUID `json:"-"`
	Title          string    `json:"title"`
	Description    string    `json:"description,omitempty"`
	IngredientText string    `json:"ingredientText"`
}

// GetName returns the resource type name for JSON:API
func (r CreateRequest) GetName() string {
	return "meals"
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

// UpdateRequest represents a JSON:API request to update a meal
type UpdateRequest struct {
	Id          uuid.UUID `json:"-"`
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
}

// GetName returns the resource type name for JSON:API
func (r UpdateRequest) GetName() string {
	return "meals"
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
