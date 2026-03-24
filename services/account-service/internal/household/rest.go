package household

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	Timezone  string    `json:"timezone"`
	Units     string    `json:"units"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "households" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id: m.Id(), Name: m.Name(), Timezone: m.Timezone(),
		Units: m.Units(), CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt(),
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

type CreateRequest struct {
	Id       uuid.UUID `json:"-"`
	Name     string    `json:"name"`
	Timezone string    `json:"timezone"`
	Units    string    `json:"units"`
}

func (r CreateRequest) GetName() string { return "households" }
func (r CreateRequest) GetID() string   { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
