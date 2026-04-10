package exercise

import (
	"time"

	"github.com/google/uuid"
)

// DefaultsRest is the per-kind defaults projection sent over the wire. All
// fields are optional and may be `null`. The frontend selects the meaningful
// subset based on the exercise's `kind`.
type DefaultsRest struct {
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	Weight          *float64 `json:"weight,omitempty"`
	WeightUnit      *string  `json:"weightUnit,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	Distance        *float64 `json:"distance,omitempty"`
	DistanceUnit    *string  `json:"distanceUnit,omitempty"`
}

type RestModel struct {
	Id                 uuid.UUID    `json:"-"`
	Name               string       `json:"name"`
	Kind               string       `json:"kind"`
	WeightType         string       `json:"weightType"`
	ThemeID            uuid.UUID    `json:"themeId"`
	RegionID           uuid.UUID    `json:"regionId"`
	SecondaryRegionIDs []uuid.UUID  `json:"secondaryRegionIds"`
	Defaults           DefaultsRest `json:"defaults"`
	Notes              *string      `json:"notes,omitempty"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
}

func (r RestModel) GetName() string         { return "exercises" }
func (r RestModel) GetID() string            { return r.Id.String() }
func (r *RestModel) SetID(id string) error   { var err error; r.Id, err = uuid.Parse(id); return err }

type CreateRequest struct {
	Id                 uuid.UUID    `json:"-"`
	Name               string       `json:"name"`
	Kind               string       `json:"kind"`
	WeightType         string       `json:"weightType"`
	ThemeID            uuid.UUID    `json:"themeId"`
	RegionID           uuid.UUID    `json:"regionId"`
	SecondaryRegionIDs []uuid.UUID  `json:"secondaryRegionIds"`
	Defaults           DefaultsRest `json:"defaults"`
	Notes              *string      `json:"notes,omitempty"`
}

func (r CreateRequest) GetName() string       { return "exercises" }
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
	Id                 uuid.UUID     `json:"-"`
	Name               *string       `json:"name,omitempty"`
	Kind               string        `json:"kind,omitempty"`
	WeightType         string        `json:"weightType,omitempty"`
	ThemeID            *uuid.UUID    `json:"themeId,omitempty"`
	RegionID           *uuid.UUID    `json:"regionId,omitempty"`
	SecondaryRegionIDs *[]uuid.UUID  `json:"secondaryRegionIds,omitempty"`
	Defaults           *DefaultsRest `json:"defaults,omitempty"`
	Notes              *string       `json:"notes,omitempty"`
}

func (r UpdateRequest) GetName() string       { return "exercises" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model) RestModel {
	secondary := m.SecondaryRegionIDs()
	if secondary == nil {
		secondary = []uuid.UUID{}
	}
	return RestModel{
		Id:                 m.Id(),
		Name:               m.Name(),
		Kind:               m.Kind(),
		WeightType:         m.WeightType(),
		ThemeID:            m.ThemeID(),
		RegionID:           m.RegionID(),
		SecondaryRegionIDs: secondary,
		Defaults: DefaultsRest{
			Sets:            m.DefaultSets(),
			Reps:            m.DefaultReps(),
			Weight:          m.DefaultWeight(),
			WeightUnit:      m.DefaultWeightUnit(),
			DurationSeconds: m.DefaultDurationSeconds(),
			Distance:        m.DefaultDistance(),
			DistanceUnit:    m.DefaultDistanceUnit(),
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
