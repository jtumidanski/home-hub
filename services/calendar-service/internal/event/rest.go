package event

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id              uuid.UUID  `json:"-"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	StartTime       time.Time  `json:"startTime"`
	EndTime         time.Time  `json:"endTime"`
	AllDay          bool       `json:"allDay"`
	Location        *string    `json:"location"`
	Visibility      string     `json:"visibility"`
	UserDisplayName string     `json:"userDisplayName"`
	UserColor       string     `json:"userColor"`
	IsOwner         bool       `json:"isOwner"`
}

func (r RestModel) GetName() string { return "calendar-events" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func TransformWithPrivacy(m Model, requesterUserID uuid.UUID) (RestModel, error) {
	isOwner := m.UserID() == requesterUserID
	rm := RestModel{
		Id:              m.Id(),
		StartTime:       m.StartTime(),
		EndTime:         m.EndTime(),
		AllDay:          m.AllDay(),
		Visibility:      m.Visibility(),
		UserDisplayName: m.UserDisplayName(),
		UserColor:       m.UserColor(),
		IsOwner:         isOwner,
	}

	if m.IsPrivate() && !isOwner {
		rm.Title = "Busy"
		rm.Description = nil
		rm.Location = nil
	} else {
		rm.Title = m.Title()
		desc := m.Description()
		if desc != "" {
			rm.Description = &desc
		}
		loc := m.Location()
		if loc != "" {
			rm.Location = &loc
		}
	}
	return rm, nil
}

func TransformSliceWithPrivacy(models []Model, requesterUserID uuid.UUID) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := TransformWithPrivacy(m, requesterUserID)
		if err != nil {
			return nil, err
		}
		result[i] = rm
	}
	return result, nil
}
