package reminder

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id               uuid.UUID  `json:"-"`
	Title            string     `json:"title"`
	Notes            string     `json:"notes,omitempty"`
	ScheduledFor     time.Time  `json:"scheduledFor"`
	OwnerUserId      *string    `json:"ownerUserId"`
	Active           bool       `json:"active"`
	LastDismissedAt  *time.Time `json:"lastDismissedAt,omitempty"`
	LastSnoozedUntil *time.Time `json:"lastSnoozedUntil,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (r RestModel) GetName() string       { return "reminders" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model) (RestModel, error) {
	var ownerUserId *string
	if m.OwnerUserID() != nil {
		s := m.OwnerUserID().String()
		ownerUserId = &s
	}
	return RestModel{
		Id: m.Id(), Title: m.Title(), Notes: m.Notes(),
		ScheduledFor: m.ScheduledFor(), OwnerUserId: ownerUserId, Active: m.IsActive(),
		LastDismissedAt: m.LastDismissedAt(), LastSnoozedUntil: m.LastSnoozedUntil(),
		CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt(),
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
	Id           uuid.UUID `json:"-"`
	Title        string    `json:"title"`
	Notes        string    `json:"notes"`
	ScheduledFor string    `json:"scheduledFor"`
	OwnerUserId  *string   `json:"ownerUserId,omitempty"`
}

func (r CreateRequest) GetName() string       { return "reminders" }
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
	Id           uuid.UUID `json:"-"`
	Title        string    `json:"title"`
	Notes        string    `json:"notes"`
	ScheduledFor string    `json:"scheduledFor"`
	OwnerUserId  *string   `json:"ownerUserId,omitempty"`
}

func (r UpdateRequest) GetName() string       { return "reminders" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }
