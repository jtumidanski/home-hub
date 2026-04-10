package planneditem

import (
	"time"

	"github.com/google/uuid"
)

// PlannedRest is the kind-shaped planned-values projection. All fields are
// optional and may be `null`. The frontend selects the meaningful subset
// based on the parent exercise's `kind`.
type PlannedRest struct {
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	Weight          *float64 `json:"weight,omitempty"`
	WeightUnit      *string  `json:"weightUnit,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	Distance        *float64 `json:"distance,omitempty"`
	DistanceUnit    *string  `json:"distanceUnit,omitempty"`
}

// RestModel is the JSON:API resource for a planned item. The composite
// week document in `weekview` embeds a richer `ItemRest` that joins the
// exercise catalog; this type covers only the planneditem domain fields
// so the package is independently serializable.
type RestModel struct {
	Id         uuid.UUID   `json:"-"`
	WeekID     uuid.UUID   `json:"weekId"`
	ExerciseID uuid.UUID   `json:"exerciseId"`
	DayOfWeek  int         `json:"dayOfWeek"`
	Position   int         `json:"position"`
	Planned    PlannedRest `json:"planned"`
	Notes      *string     `json:"notes,omitempty"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}

func (r RestModel) GetName() string       { return "planned-items" }
func (r RestModel) GetID() string         { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) RestModel {
	return RestModel{
		Id:         m.Id(),
		WeekID:     m.WeekID(),
		ExerciseID: m.ExerciseID(),
		DayOfWeek:  m.DayOfWeek(),
		Position:   m.Position(),
		Planned: PlannedRest{
			Sets:            m.PlannedSets(),
			Reps:            m.PlannedReps(),
			Weight:          m.PlannedWeight(),
			WeightUnit:      m.PlannedWeightUnit(),
			DurationSeconds: m.PlannedDurationSeconds(),
			Distance:        m.PlannedDistance(),
			DistanceUnit:    m.PlannedDistanceUnit(),
		},
		Notes:     m.Notes(),
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
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
