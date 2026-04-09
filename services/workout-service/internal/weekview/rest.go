// Package weekview owns the composite "week with embedded items" REST shape
// plus the HTTP handlers that mutate it. It's a separate package because the
// week and planneditem domain packages would otherwise import each other.
//
// The package is read-only with respect to domain logic — it delegates every
// state change to the underlying domain processors and only handles request
// parsing, projection assembly, and JSON envelope marshaling.
package weekview

import (
	"encoding/json"
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

// Document is the typed JSON:API document the week endpoint emits.
type Document struct {
	Data data `json:"data"`
}

type data struct {
	Type       string     `json:"type"`
	ID         string     `json:"id"`
	Attributes attributes `json:"attributes"`
}

type attributes struct {
	WeekStartDate string     `json:"weekStartDate"`
	RestDayFlags  []int      `json:"restDayFlags"`
	Items         []ItemRest `json:"items"`
}

func MarshalDocument(doc Document) ([]byte, error) {
	return json.Marshal(doc)
}
