package membership

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "memberships" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{Id: m.Id(), Role: m.Role(), CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt()}, nil
}

type CreateRequest struct {
	Id   uuid.UUID `json:"-"`
	Role string    `json:"role"`
}

func (r CreateRequest) GetName() string { return "memberships" }
func (r CreateRequest) GetID() string   { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
