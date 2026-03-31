package planitem

import (
	"time"

	"github.com/google/uuid"
)

// RestModel is the JSON:API representation for a plan item.
type RestModel struct {
	Id                uuid.UUID `json:"-"`
	Day               string    `json:"day"`
	Slot              string    `json:"slot"`
	RecipeID          uuid.UUID `json:"recipe_id"`
	ServingMultiplier *float64  `json:"serving_multiplier"`
	PlannedServings   *int      `json:"planned_servings"`
	Notes             *string   `json:"notes"`
	Position          int       `json:"position"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (r RestModel) GetName() string       { return "plan-items" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func TransformItem(m Model) RestModel {
	return RestModel{
		Id:                m.Id(),
		Day:               m.Day().Format("2006-01-02"),
		Slot:              m.Slot(),
		RecipeID:          m.RecipeID(),
		ServingMultiplier: m.ServingMultiplier(),
		PlannedServings:   m.PlannedServings(),
		Notes:             m.Notes(),
		Position:          m.Position(),
		CreatedAt:         m.CreatedAt(),
		UpdatedAt:         m.UpdatedAt(),
	}
}

func TransformItemSlice(models []Model) []RestModel {
	result := make([]RestModel, len(models))
	for i, m := range models {
		result[i] = TransformItem(m)
	}
	return result
}

// CreateItemRequest is the JSON:API request body for adding a plan item.
type CreateItemRequest struct {
	Id                uuid.UUID `json:"-"`
	Day               string    `json:"day"`
	Slot              string    `json:"slot"`
	RecipeID          string    `json:"recipe_id"`
	ServingMultiplier *float64  `json:"serving_multiplier,omitempty"`
	PlannedServings   *int      `json:"planned_servings,omitempty"`
	Notes             *string   `json:"notes,omitempty"`
	Position          *int      `json:"position,omitempty"`
}

func (r CreateItemRequest) GetName() string       { return "plan-items" }
func (r CreateItemRequest) GetID() string          { return r.Id.String() }
func (r *CreateItemRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// UpdateItemRequest is the JSON:API request body for updating a plan item.
type UpdateItemRequest struct {
	Id                uuid.UUID `json:"-"`
	Day               string    `json:"day,omitempty"`
	Slot              string    `json:"slot,omitempty"`
	ServingMultiplier *float64  `json:"serving_multiplier,omitempty"`
	PlannedServings   *int      `json:"planned_servings,omitempty"`
	Notes             *string   `json:"notes,omitempty"`
	Position          *int      `json:"position,omitempty"`
}

func (r UpdateItemRequest) GetName() string       { return "plan-items" }
func (r UpdateItemRequest) GetID() string          { return r.Id.String() }
func (r *UpdateItemRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }
