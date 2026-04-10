package week

import (
	"time"

	"github.com/google/uuid"
)

// RestModel is the JSON:API resource for a single week. The composite
// "week with embedded items" document lives in the `weekview` package; this
// type covers only the week domain fields so the package is independently
// serializable.
type RestModel struct {
	Id            uuid.UUID `json:"-"`
	WeekStartDate string    `json:"weekStartDate"`
	RestDayFlags  []int     `json:"restDayFlags"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string         { return "weeks" }
func (r RestModel) GetID() string           { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) RestModel {
	flags := m.RestDayFlags()
	if flags == nil {
		flags = []int{}
	}
	return RestModel{
		Id:            m.Id(),
		WeekStartDate: m.WeekStartDate().Format("2006-01-02"),
		RestDayFlags:  flags,
		CreatedAt:     m.CreatedAt(),
		UpdatedAt:     m.UpdatedAt(),
	}
}

func TransformSlice(models []Model) []*RestModel {
	out := make([]*RestModel, len(models))
	for i, m := range models {
		rm := Transform(m)
		out[i] = &rm
	}
	return out
}
