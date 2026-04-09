package theme

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string         { return "themes" }
func (r RestModel) GetID() string            { return r.Id.String() }
func (r *RestModel) SetID(id string) error   { var err error; r.Id, err = uuid.Parse(id); return err }

type CreateRequest struct {
	Id        uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sortOrder"`
}

func (r CreateRequest) GetName() string       { return "themes" }
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
	SortOrder *int      `json:"sortOrder,omitempty"`
}

func (r UpdateRequest) GetName() string       { return "themes" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model) RestModel {
	return RestModel{
		Id:        m.Id(),
		Name:      m.Name(),
		SortOrder: m.SortOrder(),
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
	}
}

func TransformSlice(models []Model) []*RestModel {
	out := make([]*RestModel, len(models))
	for i, m := range models {
		rm := Transform(m)
		out[i] = &rm
	}
	return out
}
