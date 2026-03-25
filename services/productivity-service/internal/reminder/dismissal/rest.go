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

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:        m.Id(),
		CreatedAt: m.CreatedAt(),
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
	Id         uuid.UUID `json:"-"`
	ReminderId string    `json:"reminderId"`
}

func (r CreateRequest) GetName() string       { return "reminder-dismissals" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
