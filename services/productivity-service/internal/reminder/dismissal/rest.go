package dismissal

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

func (r RestModel) GetName() string       { return "reminder-dismissals" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(e Entity) (RestModel, error) {
	return RestModel{
		Id:        e.Id,
		CreatedAt: e.CreatedAt,
	}, nil
}

type CreateRequest struct {
	Id         uuid.UUID `json:"-"`
	ReminderId string    `json:"reminderId"`
}

func (r CreateRequest) GetName() string       { return "reminder-dismissals" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }
