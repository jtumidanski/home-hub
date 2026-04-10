package summary

import (
	"github.com/google/uuid"
)

// quantity is a paired value+unit used by strength volume and cardio totals.
type quantity struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// cardioBlock is the cardio sub-tally for a theme/region group.
type cardioBlock struct {
	TotalDurationSeconds int      `json:"totalDurationSeconds"`
	TotalDistance        quantity `json:"totalDistance"`
}

// groupRest is the per-group totals carried by both theme and region groups.
type groupRest struct {
	ItemCount      int          `json:"itemCount"`
	StrengthVolume *quantity    `json:"strengthVolume"`
	Cardio         *cardioBlock `json:"cardio"`
}

type themeGroup struct {
	ThemeID   uuid.UUID `json:"themeId"`
	ThemeName string    `json:"themeName"`
	groupRest
}

type regionGroup struct {
	RegionID   uuid.UUID `json:"regionId"`
	RegionName string    `json:"regionName"`
	groupRest
}

type dayItem struct {
	ItemID        uuid.UUID `json:"itemId"`
	ExerciseName  string    `json:"exerciseName"`
	Status        string    `json:"status"`
	Planned       any       `json:"planned"`
	ActualSummary any       `json:"actualSummary"`
}

type dayBlock struct {
	DayOfWeek int       `json:"dayOfWeek"`
	IsRestDay bool      `json:"isRestDay"`
	Items     []dayItem `json:"items"`
}

// RestModel is the JSON:API resource for the per-week summary projection.
// The id is the week's ISO start date so the URL identifier and the resource
// identifier match (the URL is `/weeks/{weekStart}/summary`).
type RestModel struct {
	Id                  string        `json:"-"`
	WeekStartDate       string        `json:"weekStartDate"`
	RestDayFlags        []int         `json:"restDayFlags"`
	TotalPlannedItems   int           `json:"totalPlannedItems"`
	TotalPerformedItems int           `json:"totalPerformedItems"`
	TotalSkippedItems   int           `json:"totalSkippedItems"`
	ByDay               []dayBlock    `json:"byDay"`
	ByTheme             []themeGroup  `json:"byTheme"`
	ByRegion            []regionGroup `json:"byRegion"`
}

func (r RestModel) GetName() string         { return "week-summaries" }
func (r RestModel) GetID() string           { return r.Id }
func (r *RestModel) SetID(id string) error { r.Id = id; return nil }
