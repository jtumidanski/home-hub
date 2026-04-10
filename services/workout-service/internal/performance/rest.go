package performance

import (
	"github.com/google/uuid"
)

// ActualsRest is the summary actuals projection sent over the wire. Mirrors
// `PerformanceActualsAttrs` minus the request-side concerns. Pointer fields
// preserve "absent" semantics so the client can render only the meaningful
// shape for the parent exercise's kind.
type ActualsRest struct {
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	Weight          *float64 `json:"weight,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	Distance        *float64 `json:"distance,omitempty"`
	DistanceUnit    *string  `json:"distanceUnit,omitempty"`
}

// SetRest is one row of a per-set performance.
type SetRest struct {
	SetNumber int     `json:"setNumber"`
	Reps      int     `json:"reps"`
	Weight    float64 `json:"weight"`
}

// RestModel is the JSON:API resource for a performance. The `id` is the
// owning planned-item's UUID — the URL path uses `/items/{itemId}/performance`
// so clients identify a performance by its parent item, not by the row's own
// primary key. `Actuals` is populated in summary mode and `Sets` in per-set
// mode; the inverse field is omitted via `omitempty`.
type RestModel struct {
	Id         uuid.UUID    `json:"-"`
	Status     string       `json:"status"`
	Mode       string       `json:"mode"`
	WeightUnit *string      `json:"weightUnit,omitempty"`
	Actuals    *ActualsRest `json:"actuals,omitempty"`
	Sets       []SetRest    `json:"sets,omitempty"`
	Notes      *string      `json:"notes,omitempty"`
}

func (r RestModel) GetName() string { return "performances" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// Transform projects a performance model + its per-set rows into the REST
// resource. The `id` field is set from the model's planned_item_id so the
// resource shape matches the URL identifier.
func Transform(m Model, sets []SetModel) RestModel {
	out := RestModel{
		Id:         m.PlannedItemID(),
		Status:     m.Status(),
		Mode:       m.Mode(),
		WeightUnit: m.WeightUnit(),
		Notes:      m.Notes(),
	}
	if m.Mode() == ModePerSet {
		rows := make([]SetRest, 0, len(sets))
		for _, s := range sets {
			rows = append(rows, SetRest{SetNumber: s.SetNumber(), Reps: s.Reps(), Weight: s.Weight()})
		}
		out.Sets = rows
		return out
	}
	out.Actuals = &ActualsRest{
		Sets:            m.ActualSets(),
		Reps:            m.ActualReps(),
		Weight:          m.ActualWeight(),
		DurationSeconds: m.ActualDurationSeconds(),
		Distance:        m.ActualDistance(),
		DistanceUnit:    m.ActualDistanceUnit(),
	}
	return out
}

// TransformSlice maps a batch of performance models. Per-set rows are looked
// up via the supplied map keyed by performance row id.
func TransformSlice(models []Model, setsByPerf map[uuid.UUID][]SetModel) []*RestModel {
	out := make([]*RestModel, len(models))
	for i, m := range models {
		rm := Transform(m, setsByPerf[m.Id()])
		out[i] = &rm
	}
	return out
}

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
