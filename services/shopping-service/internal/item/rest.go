package item

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id                uuid.UUID  `json:"-"`
	Name              string     `json:"name"`
	Quantity          *string    `json:"quantity"`
	CategoryId        *uuid.UUID `json:"category_id"`
	CategoryName      *string    `json:"category_name"`
	CategorySortOrder *int       `json:"category_sort_order"`
	Checked           bool       `json:"checked"`
	Position          int        `json:"position"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (r RestModel) GetName() string       { return "shopping-items" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

// NestedRestModel mirrors RestModel for embedding inside parent resources
// (e.g. shopping list detail). It serializes Id as a normal JSON field so the
// frontend can read item.id from the embedded array. RestModel keeps json:"-"
// because api2go places the id at the resource root for top-level item
// responses.
type NestedRestModel struct {
	Id                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Quantity          *string    `json:"quantity"`
	CategoryId        *uuid.UUID `json:"category_id"`
	CategoryName      *string    `json:"category_name"`
	CategorySortOrder *int       `json:"category_sort_order"`
	Checked           bool       `json:"checked"`
	Position          int        `json:"position"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateRequest struct {
	Id         uuid.UUID  `json:"-"`
	Name       string     `json:"name"`
	Quantity   *string    `json:"quantity,omitempty"`
	CategoryId *uuid.UUID `json:"category_id,omitempty"`
	Position   *int       `json:"position,omitempty"`
}

func (r CreateRequest) GetName() string       { return "shopping-items" }
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
	Id         uuid.UUID  `json:"-"`
	Name       *string    `json:"name,omitempty"`
	Quantity   *string    `json:"quantity,omitempty"`
	CategoryId *uuid.UUID `json:"category_id,omitempty"`
	Position   *int       `json:"position,omitempty"`
}

func (r UpdateRequest) GetName() string       { return "shopping-items" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type CheckRequest struct {
	Id      uuid.UUID `json:"-"`
	Checked bool      `json:"checked"`
}

func (r CheckRequest) GetName() string       { return "shopping-items" }
func (r CheckRequest) GetID() string          { return r.Id.String() }
func (r *CheckRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:                m.Id(),
		Name:              m.Name(),
		Quantity:          m.Quantity(),
		CategoryId:        m.CategoryID(),
		CategoryName:      m.CategoryName(),
		CategorySortOrder: m.CategorySortOrder(),
		Checked:           m.Checked(),
		Position:          m.Position(),
		CreatedAt:         m.CreatedAt(),
		UpdatedAt:         m.UpdatedAt(),
	}, nil
}

func TransformNested(m Model) NestedRestModel {
	return NestedRestModel{
		Id:                m.Id(),
		Name:              m.Name(),
		Quantity:          m.Quantity(),
		CategoryId:        m.CategoryID(),
		CategoryName:      m.CategoryName(),
		CategorySortOrder: m.CategorySortOrder(),
		Checked:           m.Checked(),
		Position:          m.Position(),
		CreatedAt:         m.CreatedAt(),
		UpdatedAt:         m.UpdatedAt(),
	}
}

func TransformNestedSlice(models []Model) []NestedRestModel {
	result := make([]NestedRestModel, len(models))
	for i, m := range models {
		result[i] = TransformNested(m)
	}
	return result
}

func TransformSlice(models []Model) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		r, err := Transform(m)
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}
