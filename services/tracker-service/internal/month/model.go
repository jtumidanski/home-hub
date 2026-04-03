package month

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CompletionStats struct {
	Expected  int `json:"expected"`
	Filled    int `json:"filled"`
	Skipped   int `json:"skipped"`
	Remaining int `json:"remaining"`
}

type MonthSummary struct {
	Month      string          `json:"month"`
	Complete   bool            `json:"complete"`
	Completion CompletionStats `json:"completion"`
}

type ReportSummary struct {
	TotalItems     int     `json:"total_items"`
	CompletionRate float64 `json:"completion_rate"`
	SkipRate       float64 `json:"skip_rate"`
	TotalExpected  int     `json:"total_expected"`
	TotalFilled    int     `json:"total_filled"`
	TotalSkipped   int     `json:"total_skipped"`
}

type SentimentStats struct {
	ExpectedDays  int            `json:"expected_days"`
	FilledDays    int            `json:"filled_days"`
	SkippedDays   int            `json:"skipped_days"`
	Positive      int            `json:"positive"`
	Neutral       int            `json:"neutral"`
	Negative      int            `json:"negative"`
	PositiveRatio float64        `json:"positive_ratio"`
	DailyValues   []DailyRating  `json:"daily_values"`
}

type DailyRating struct {
	Date   string `json:"date"`
	Rating string `json:"rating"`
}

type NumericStats struct {
	ExpectedDays              int          `json:"expected_days"`
	FilledDays                int          `json:"filled_days"`
	SkippedDays               int          `json:"skipped_days"`
	Total                     int          `json:"total"`
	DailyAverage              float64      `json:"daily_average"`
	DaysWithEntriesAboveZero  int          `json:"days_with_entries_above_zero"`
	DaysWithEntriesAboveZeroPct float64    `json:"days_with_entries_above_zero_pct"`
	Min                       *DailyCount  `json:"min"`
	Max                       *DailyCount  `json:"max"`
	DailyValues               []DailyCount `json:"daily_values"`
}

type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type RangeStats struct {
	ExpectedDays int          `json:"expected_days"`
	FilledDays   int          `json:"filled_days"`
	SkippedDays  int          `json:"skipped_days"`
	Average      float64      `json:"average"`
	Min          *DailyValue  `json:"min"`
	Max          *DailyValue  `json:"max"`
	StdDev       float64      `json:"std_dev"`
	DailyValues  []DailyValue `json:"daily_values"`
}

type DailyValue struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

type ItemReport struct {
	TrackingItemId uuid.UUID       `json:"tracking_item_id"`
	Name           string          `json:"name"`
	ScaleType      string          `json:"scale_type"`
	Stats          json.RawMessage `json:"stats"`
}

type Report struct {
	Month   string       `json:"month"`
	Summary ReportSummary `json:"summary"`
	Items   []ItemReport `json:"items"`
}

type MonthItemInfo struct {
	Id              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	ScaleType       string                 `json:"scale_type"`
	ScaleConfig     json.RawMessage        `json:"scale_config"`
	Color           string                 `json:"color"`
	SortOrder       int                    `json:"sort_order"`
	ActiveFrom      string                 `json:"active_from"`
	ActiveUntil     *string                `json:"active_until"`
	ScheduleSnapshots []ScheduleSnapshotInfo `json:"schedule_snapshots"`
}

type ScheduleSnapshotInfo struct {
	Schedule      []int  `json:"schedule"`
	EffectiveDate string `json:"effective_date"`
}
