package performance

import (
	"github.com/google/uuid"
)

// PerformanceActualsAttrs is the summary actuals payload nested inside a
// PATCH request. Pointer fields preserve "omitted vs explicit value".
type PerformanceActualsAttrs struct {
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	Weight          *float64 `json:"weight,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	Distance        *float64 `json:"distance,omitempty"`
	DistanceUnit    *string  `json:"distanceUnit,omitempty"`
}

// PatchPerformanceRequest is the body of
// `PATCH /workouts/weeks/{weekStart}/items/{itemId}/performance`.
type PatchPerformanceRequest struct {
	Id         uuid.UUID                `json:"-"`
	Status     *string                  `json:"status,omitempty"`
	WeightUnit *string                  `json:"weightUnit,omitempty"`
	Actuals    *PerformanceActualsAttrs `json:"actuals,omitempty"`
	Notes      *string                  `json:"notes,omitempty"`
}

func (r PatchPerformanceRequest) GetName() string { return "performances" }
func (r PatchPerformanceRequest) GetID() string   { return r.Id.String() }
func (r *PatchPerformanceRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// PerformanceSetAttrs is one row in a per-set replace payload. The server
// assigns set_number from the array index, so the client never sends it.
type PerformanceSetAttrs struct {
	Reps   int     `json:"reps"`
	Weight float64 `json:"weight"`
}

// PutPerformanceSetsRequest is the body of
// `PUT /workouts/weeks/{weekStart}/items/{itemId}/performance/sets`.
type PutPerformanceSetsRequest struct {
	Id         uuid.UUID             `json:"-"`
	WeightUnit string                `json:"weightUnit"`
	Sets       []PerformanceSetAttrs `json:"sets"`
}

func (r PutPerformanceSetsRequest) GetName() string { return "performances" }
func (r PutPerformanceSetsRequest) GetID() string   { return r.Id.String() }
func (r *PutPerformanceSetsRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
