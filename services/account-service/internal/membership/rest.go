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
	Id          uuid.UUID `json:"-"`
	HouseholdId uuid.UUID `json:"household_id"`
	UserId      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
}

func (r CreateRequest) GetName() string { return "memberships" }
func (r CreateRequest) GetID() string   { return r.Id.String() }
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
	Role string    `json:"role"`
}

func (r UpdateRequest) GetName() string { return "memberships" }
func (r UpdateRequest) GetID() string   { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
