package source

import (
	"github.com/google/uuid"
)

type RestModel struct {
	Id      uuid.UUID `json:"-"`
	Name    string    `json:"name"`
	Primary bool      `json:"primary"`
	Visible bool      `json:"visible"`
	Color   string    `json:"color"`
}

func (r RestModel) GetName() string { return "calendar-sources" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:      m.Id(),
		Name:    m.Name(),
		Primary: m.Primary(),
		Visible: m.Visible(),
		Color:   m.Color(),
	}, nil
}

func TransformSlice(models []Model) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := Transform(m)
		if err != nil {
			return nil, err
		}
		result[i] = rm
	}
	return result, nil
}

type ToggleRequest struct {
	Id      uuid.UUID `json:"-"`
	Visible bool      `json:"visible"`
}

func (r ToggleRequest) GetName() string { return "calendar-sources" }
func (r ToggleRequest) GetID() string   { return r.Id.String() }
func (r *ToggleRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
