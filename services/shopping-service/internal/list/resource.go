package list

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/item"
)

type RestModel struct {
	Id           uuid.UUID       `json:"-"`
	Name         string          `json:"name"`
	Status       string          `json:"status"`
	ItemCount    int             `json:"item_count"`
	CheckedCount int             `json:"checked_count"`
	ArchivedAt   *time.Time      `json:"archived_at"`
	Items        []item.RestModel `json:"items,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

func (r RestModel) GetName() string       { return "shopping-lists" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type CreateRequest struct {
	Id   uuid.UUID `json:"-"`
	Name string    `json:"name"`
}

func (r CreateRequest) GetName() string       { return "shopping-lists" }
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
	Id   uuid.UUID `json:"-"`
	Name string    `json:"name"`
}

func (r UpdateRequest) GetName() string       { return "shopping-lists" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type ImportRequest struct {
	Id     uuid.UUID `json:"-"`
	PlanId uuid.UUID `json:"plan_id"`
}

func (r ImportRequest) GetName() string       { return "shopping-list-imports" }
func (r ImportRequest) GetID() string          { return r.Id.String() }
func (r *ImportRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) RestModel {
	return RestModel{
		Id:           m.Id(),
		Name:         m.Name(),
		Status:       m.Status(),
		ItemCount:    m.ItemCount(),
		CheckedCount: m.CheckedCount(),
		ArchivedAt:   m.ArchivedAt(),
		CreatedAt:    m.CreatedAt(),
		UpdatedAt:    m.UpdatedAt(),
	}
}

func TransformWithItems(m Model, items []item.RestModel) RestModel {
	r := Transform(m)
	r.Items = items
	return r
}

func TransformSlice(models []Model) []RestModel {
	result := make([]RestModel, len(models))
	for i, m := range models {
		result[i] = Transform(m)
	}
	return result
}
