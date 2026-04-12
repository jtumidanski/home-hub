// Package weekview owns the composite "week with embedded items" REST shape
// plus the HTTP handlers that mutate it. It's a separate package because the
// week and planneditem domain packages would otherwise import each other.
//
// The package is read-only with respect to domain logic — it delegates every
// state change to the underlying domain processors and only handles request
// parsing, projection assembly, and JSON envelope marshaling.
package weekview

import (
	"time"

	"github.com/google/uuid"
)

// PlannedRest is the kind-shaped planned-values block. All fields are optional.
type PlannedRest struct {
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	Weight          *float64 `json:"weight,omitempty"`
	WeightUnit      *string  `json:"weightUnit,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	Distance        *float64 `json:"distance,omitempty"`
	DistanceUnit    *string  `json:"distanceUnit,omitempty"`
}

// PerformanceSetRest is one per-set row in `mode=per_set`.
type PerformanceSetRest struct {
	SetNumber int     `json:"setNumber"`
	Reps      int     `json:"reps"`
	Weight    float64 `json:"weight"`
}

// ActualsRest is the summary actuals block. Mirrors PlannedRest but without
// the planned-only `weightUnit` (the unit lives at the performance level).
type ActualsRest struct {
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	Weight          *float64 `json:"weight,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	Distance        *float64 `json:"distance,omitempty"`
	DistanceUnit    *string  `json:"distanceUnit,omitempty"`
}

// PerformanceRest is the embedded performance projection. `Actuals` is null in
// per_set mode and populated in summary mode; `Sets` is the inverse.
type PerformanceRest struct {
	Status     string                `json:"status"`
	Mode       string                `json:"mode"`
	WeightUnit *string               `json:"weightUnit,omitempty"`
	Actuals    *ActualsRest          `json:"actuals,omitempty"`
	Sets       []PerformanceSetRest  `json:"sets,omitempty"`
	Notes      *string               `json:"notes,omitempty"`
	UpdatedAt  *time.Time            `json:"updatedAt,omitempty"`
}

// ItemRest is one row inside `weeks.attributes.items`. The naming convention
// matches `api-contracts.md` §4.1 verbatim.
type ItemRest struct {
	ID              uuid.UUID        `json:"id"`
	DayOfWeek       int              `json:"dayOfWeek"`
	Position        int              `json:"position"`
	ExerciseID      uuid.UUID        `json:"exerciseId"`
	ExerciseName    string           `json:"exerciseName"`
	ExerciseDeleted bool             `json:"exerciseDeleted"`
	Kind            string           `json:"kind"`
	WeightType      string           `json:"weightType"`
	Planned         PlannedRest      `json:"planned"`
	Performance     *PerformanceRest `json:"performance,omitempty"`
	Notes           *string          `json:"notes,omitempty"`
}

// RestModel is the JSON:API resource for the composite "week with embedded
// items" view. The struct is flat — api2go marshals every json-tagged field
// into the response's `attributes` block automatically.
type RestModel struct {
	Id            uuid.UUID  `json:"-"`
	WeekStartDate string     `json:"weekStartDate"`
	RestDayFlags  []int      `json:"restDayFlags"`
	Items         []ItemRest `json:"items"`
}

func (r RestModel) GetName() string { return "weeks" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// --- request types --------------------------------------------------------
//
// All request structs implement the JSON:API EntityNamer + UnmarshalIdentifier
// interfaces so `server.RegisterInputHandler[T]` can deserialize them
// directly. The api2go layer strips the `{data: {type, id, attributes}}`
// envelope and json-unmarshals the attributes into the struct, so request
// payloads can be modeled as flat structs.

// PatchWeekRequest is the body of `PATCH /workouts/weeks/{weekStart}`. Only
// `restDayFlags` is patchable today; the field is a pointer so we can detect
// "field omitted" vs "explicit empty array" — both meaningful.
type PatchWeekRequest struct {
	Id           uuid.UUID `json:"-"`
	RestDayFlags *[]int    `json:"restDayFlags,omitempty"`
}

func (r PatchWeekRequest) GetName() string { return "weeks" }
func (r PatchWeekRequest) GetID() string   { return r.Id.String() }
func (r *PatchWeekRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// CopyWeekRequest is the body of `POST /workouts/weeks/{weekStart}/copy`.
type CopyWeekRequest struct {
	Id   uuid.UUID `json:"-"`
	Mode string    `json:"mode"`
}

func (r CopyWeekRequest) GetName() string { return "weeks" }
func (r CopyWeekRequest) GetID() string   { return r.Id.String() }
func (r *CopyWeekRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// AddPlannedItemRequest is the body of `POST /weeks/{weekStart}/items`.
type AddPlannedItemRequest struct {
	Id                     uuid.UUID `json:"-"`
	ExerciseID             uuid.UUID `json:"exerciseId"`
	DayOfWeek              int       `json:"dayOfWeek"`
	Position               *int      `json:"position,omitempty"`
	PlannedSets            *int      `json:"plannedSets,omitempty"`
	PlannedReps            *int      `json:"plannedReps,omitempty"`
	PlannedWeight          *float64  `json:"plannedWeight,omitempty"`
	PlannedWeightUnit      *string   `json:"plannedWeightUnit,omitempty"`
	PlannedDurationSeconds *int      `json:"plannedDurationSeconds,omitempty"`
	PlannedDistance        *float64  `json:"plannedDistance,omitempty"`
	PlannedDistanceUnit    *string   `json:"plannedDistanceUnit,omitempty"`
	Notes                  *string   `json:"notes,omitempty"`
}

func (r AddPlannedItemRequest) GetName() string { return "planned-items" }
func (r AddPlannedItemRequest) GetID() string   { return r.Id.String() }
func (r *AddPlannedItemRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// BulkAddPlannedItemsRequest is the body of `POST /weeks/{weekStart}/items/bulk`.
// Each item shares the same shape as the single-add request.
type BulkAddPlannedItemsRequest struct {
	Id    uuid.UUID                 `json:"-"`
	Items []BulkAddPlannedItemEntry `json:"items"`
}

// BulkAddPlannedItemEntry is one row in a bulk-add request.
type BulkAddPlannedItemEntry struct {
	ExerciseID             uuid.UUID `json:"exerciseId"`
	DayOfWeek              int       `json:"dayOfWeek"`
	Position               *int      `json:"position,omitempty"`
	PlannedSets            *int      `json:"plannedSets,omitempty"`
	PlannedReps            *int      `json:"plannedReps,omitempty"`
	PlannedWeight          *float64  `json:"plannedWeight,omitempty"`
	PlannedWeightUnit      *string   `json:"plannedWeightUnit,omitempty"`
	PlannedDurationSeconds *int      `json:"plannedDurationSeconds,omitempty"`
	PlannedDistance        *float64  `json:"plannedDistance,omitempty"`
	PlannedDistanceUnit    *string   `json:"plannedDistanceUnit,omitempty"`
	Notes                  *string   `json:"notes,omitempty"`
}

func (r BulkAddPlannedItemsRequest) GetName() string { return "planned-items" }
func (r BulkAddPlannedItemsRequest) GetID() string   { return r.Id.String() }
func (r *BulkAddPlannedItemsRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// UpdatePlannedItemRequest is the body of `PATCH /weeks/{weekStart}/items/{itemId}`.
type UpdatePlannedItemRequest struct {
	Id                     uuid.UUID `json:"-"`
	DayOfWeek              *int      `json:"dayOfWeek,omitempty"`
	Position               *int      `json:"position,omitempty"`
	PlannedSets            *int      `json:"plannedSets,omitempty"`
	PlannedReps            *int      `json:"plannedReps,omitempty"`
	PlannedWeight          *float64  `json:"plannedWeight,omitempty"`
	PlannedWeightUnit      *string   `json:"plannedWeightUnit,omitempty"`
	PlannedDurationSeconds *int      `json:"plannedDurationSeconds,omitempty"`
	PlannedDistance        *float64  `json:"plannedDistance,omitempty"`
	PlannedDistanceUnit    *string   `json:"plannedDistanceUnit,omitempty"`
	Notes                  *string   `json:"notes,omitempty"`
}

func (r UpdatePlannedItemRequest) GetName() string { return "planned-items" }
func (r UpdatePlannedItemRequest) GetID() string   { return r.Id.String() }
func (r *UpdatePlannedItemRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// ReorderPlannedItemEntry is one row in the reorder request payload.
type ReorderPlannedItemEntry struct {
	ItemID    uuid.UUID `json:"itemId"`
	DayOfWeek int       `json:"dayOfWeek"`
	Position  int       `json:"position"`
}

// ReorderPlannedItemsRequest is the body of `POST /weeks/{weekStart}/items/reorder`.
type ReorderPlannedItemsRequest struct {
	Id    uuid.UUID                 `json:"-"`
	Items []ReorderPlannedItemEntry `json:"items"`
}

func (r ReorderPlannedItemsRequest) GetName() string { return "planned-items" }
func (r ReorderPlannedItemsRequest) GetID() string   { return r.Id.String() }
func (r *ReorderPlannedItemsRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
