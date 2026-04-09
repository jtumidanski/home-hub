package summary

import "github.com/google/uuid"

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

type document struct {
	Data data `json:"data"`
}

type data struct {
	Type       string     `json:"type"`
	ID         string     `json:"id"`
	Attributes attributes `json:"attributes"`
}

type attributes struct {
	WeekStartDate       string        `json:"weekStartDate"`
	RestDayFlags        []int         `json:"restDayFlags"`
	TotalPlannedItems   int           `json:"totalPlannedItems"`
	TotalPerformedItems int           `json:"totalPerformedItems"`
	TotalSkippedItems   int           `json:"totalSkippedItems"`
	ByDay               []dayBlock    `json:"byDay"`
	ByTheme             []themeGroup  `json:"byTheme"`
	ByRegion            []regionGroup `json:"byRegion"`
}
