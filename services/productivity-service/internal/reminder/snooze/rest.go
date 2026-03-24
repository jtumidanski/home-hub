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

func Transform(e Entity) (RestModel, error) {
	return RestModel{
		Id:              e.Id,
		DurationMinutes: e.DurationMinutes,
		SnoozedUntil:    e.SnoozedUntil,
		CreatedAt:       e.CreatedAt,
	}, nil
}

type CreateRequest struct {
	Id              uuid.UUID `json:"-"`
	ReminderId      string    `json:"reminderId"`
	DurationMinutes int       `json:"durationMinutes"`
}

func (r CreateRequest) GetName() string       { return "reminder-snoozes" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }
