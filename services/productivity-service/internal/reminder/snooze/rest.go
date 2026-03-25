package snooze

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id              uuid.UUID `json:"-"`
	DurationMinutes int       `json:"durationMinutes"`
	SnoozedUntil    time.Time `json:"snoozedUntil"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (r RestModel) GetName() string       { return "reminder-snoozes" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:              m.Id(),
		DurationMinutes: m.DurationMinutes(),
		SnoozedUntil:    m.SnoozedUntil(),
		CreatedAt:       m.CreatedAt(),
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
	Id              uuid.UUID `json:"-"`
	ReminderId      string    `json:"reminderId"`
	DurationMinutes int       `json:"durationMinutes"`
}

func (r CreateRequest) GetName() string       { return "reminder-snoozes" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }
