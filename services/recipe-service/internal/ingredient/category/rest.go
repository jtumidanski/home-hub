package category

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id              uuid.UUID `json:"-"`
	Name            string    `json:"name"`
	SortOrder       int       `json:"sort_order"`
	IngredientCount int       `json:"ingredient_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (r RestModel) GetName() string       { return "ingredient-categories" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type CreateRequest struct {
	Id   uuid.UUID `json:"-"`
	Name string    `json:"name"`
}

func (r CreateRequest) GetName() string       { return "ingredient-categories" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type UpdateRequest struct {
	Id        uuid.UUID `json:"-"`
	Name      string    `json:"name,omitempty"`
	SortOrder *int      `json:"sort_order,omitempty"`
}

func (r UpdateRequest) GetName() string       { return "ingredient-categories" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model) RestModel {
	return RestModel{
		Id:              m.Id(),
		Name:            m.Name(),
		SortOrder:       m.SortOrder(),
		IngredientCount: m.IngredientCount(),
		CreatedAt:       m.CreatedAt(),
		UpdatedAt:       m.UpdatedAt(),
	}
}
