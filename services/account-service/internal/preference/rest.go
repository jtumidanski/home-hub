package preference

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Theme     string    `json:"theme"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "preferences" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{Id: m.Id(), Theme: m.Theme(), CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt()}, nil
}

type UpdateRequest struct {
	Id    uuid.UUID `json:"-"`
	Theme *string   `json:"theme,omitempty"`
}

func (r UpdateRequest) GetName() string { return "preferences" }
func (r UpdateRequest) GetID() string   { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
