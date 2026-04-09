package month

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Month is a derived sub-domain: it has no persisted Model of its own, only
// computed value types and the REST projections that pair with them. The
// month summary endpoint returns a composite document — a single resource
// with two heterogeneous relationship arrays (items + entries). api2go cannot
// represent this exact shape natively, so the rest layer assembles a typed
// JSON:API document. The handler stays free of envelope construction; it only
// invokes Marshal* and forwards bytes (or surfaces a marshal error).

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
	ExpectedDays  int           `json:"expected_days"`
	FilledDays    int           `json:"filled_days"`
	SkippedDays   int           `json:"skipped_days"`
	Positive      int           `json:"positive"`
	Neutral       int           `json:"neutral"`
	Negative      int           `json:"negative"`
	PositiveRatio float64       `json:"positive_ratio"`
	DailyValues   []DailyRating `json:"daily_values"`
}

type DailyRating struct {
	Date   string `json:"date"`
	Rating string `json:"rating"`
}

type NumericStats struct {
	ExpectedDays                int          `json:"expected_days"`
	FilledDays                  int          `json:"filled_days"`
	SkippedDays                 int          `json:"skipped_days"`
	Total                       int          `json:"total"`
	DailyAverage                float64      `json:"daily_average"`
	DaysWithEntriesAboveZero    int          `json:"days_with_entries_above_zero"`
	DaysWithEntriesAboveZeroPct float64      `json:"days_with_entries_above_zero_pct"`
	Min                         *DailyCount  `json:"min"`
	Max                         *DailyCount  `json:"max"`
	DailyValues                 []DailyCount `json:"daily_values"`
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
	Month   string        `json:"month"`
	Summary ReportSummary `json:"summary"`
	Items   []ItemReport  `json:"items"`
}

type MonthItemInfo struct {
	Id                uuid.UUID              `json:"id"`
	Name              string                 `json:"name"`
	ScaleType         string                 `json:"scale_type"`
	ScaleConfig       json.RawMessage        `json:"scale_config"`
	Color             string                 `json:"color"`
	SortOrder         int                    `json:"sort_order"`
	ActiveFrom        string                 `json:"active_from"`
	ActiveUntil       *string                `json:"active_until"`
	ScheduleSnapshots []ScheduleSnapshotInfo `json:"schedule_snapshots"`
}

type ScheduleSnapshotInfo struct {
	Schedule      []int  `json:"schedule"`
	EffectiveDate string `json:"effective_date"`
}

type MonthSummaryRest struct {
	Id         string          `json:"-"`
	Month      string          `json:"month"`
	Complete   bool            `json:"complete"`
	Completion CompletionStats `json:"completion"`
}

func (r MonthSummaryRest) GetName() string       { return "tracker-months" }
func (r MonthSummaryRest) GetID() string          { return r.Id }
func (r *MonthSummaryRest) SetID(id string) error { r.Id = id; return nil }

type ReportRest struct {
	Id      string        `json:"-"`
	Month   string        `json:"month"`
	Summary ReportSummary `json:"summary"`
	Items   []ItemReport  `json:"items"`
}

func (r ReportRest) GetName() string       { return "tracker-reports" }
func (r ReportRest) GetID() string          { return r.Id }
func (r *ReportRest) SetID(id string) error { r.Id = id; return nil }

// EntryView mirrors the wire shape used by the calendar grid for each entry
// embedded in a month summary.
type EntryView struct {
	Id             uuid.UUID       `json:"id"`
	Type           string          `json:"type"`
	TrackingItemId uuid.UUID       `json:"tracking_item_id"`
	Date           string          `json:"date"`
	Value          json.RawMessage `json:"value"`
	Skipped        bool            `json:"skipped"`
	Note           *string         `json:"note,omitempty"`
	Scheduled      bool            `json:"scheduled"`
}

type monthSummaryDoc struct {
	Data monthSummaryData `json:"data"`
}

type monthSummaryData struct {
	Type          string                    `json:"type"`
	Attributes    MonthSummary              `json:"attributes"`
	Relationships monthSummaryRelationships `json:"relationships"`
}

type monthSummaryRelationships struct {
	Items   monthItemsRel   `json:"items"`
	Entries monthEntriesRel `json:"entries"`
}

type monthItemsRel struct {
	Data []MonthItemInfo `json:"data"`
}

type monthEntriesRel struct {
	Data []EntryView `json:"data"`
}

// MarshalMonthSummary converts a SummaryDetail into the wire-format JSON:API
// document used by the calendar grid.
func MarshalMonthSummary(detail SummaryDetail) ([]byte, error) {
	monthStart, err := time.Parse("2006-01", detail.Summary.Month)
	if err != nil {
		return nil, err
	}

	itemInfos := make([]MonthItemInfo, 0, len(detail.Items))
	for _, item := range detail.Items {
		activeFrom := item.CreatedAt().Truncate(24 * time.Hour)
		if activeFrom.Before(monthStart) {
			activeFrom = monthStart
		}
		var activeUntil *string
		if item.DeletedAt() != nil {
			s := item.DeletedAt().Format("2006-01-02")
			activeUntil = &s
		}

		var snapInfos []ScheduleSnapshotInfo
		for _, snap := range detail.SnapshotsByItem[item.Id()] {
			snapInfos = append(snapInfos, ScheduleSnapshotInfo{
				Schedule:      snap.Schedule(),
				EffectiveDate: snap.EffectiveDate().Format("2006-01-02"),
			})
		}

		itemInfos = append(itemInfos, MonthItemInfo{
			Id:                item.Id(),
			Name:              item.Name(),
			ScaleType:         item.ScaleType(),
			ScaleConfig:       item.ScaleConfig(),
			Color:             item.Color(),
			SortOrder:         item.SortOrder(),
			ActiveFrom:        activeFrom.Format("2006-01-02"),
			ActiveUntil:       activeUntil,
			ScheduleSnapshots: snapInfos,
		})
	}

	entryViews := make([]EntryView, 0, len(detail.Entries))
	for _, e := range detail.Entries {
		entryViews = append(entryViews, EntryView{
			Id:             e.Id(),
			Type:           "tracker-entries",
			TrackingItemId: e.TrackingItemID(),
			Date:           e.Date().Format("2006-01-02"),
			Value:          e.Value(),
			Skipped:        e.Skipped(),
			Note:           e.Note(),
			Scheduled:      true,
		})
	}

	doc := monthSummaryDoc{
		Data: monthSummaryData{
			Type:       "tracker-months",
			Attributes: detail.Summary,
			Relationships: monthSummaryRelationships{
				Items:   monthItemsRel{Data: itemInfos},
				Entries: monthEntriesRel{Data: entryViews},
			},
		},
	}
	return json.Marshal(doc)
}

type reportDoc struct {
	Data reportData `json:"data"`
}

type reportData struct {
	Type       string `json:"type"`
	Attributes Report `json:"attributes"`
}

// MarshalReport wraps a Report in a single-resource JSON:API document.
func MarshalReport(report Report) ([]byte, error) {
	doc := reportDoc{
		Data: reportData{
			Type:       "tracker-reports",
			Attributes: report,
		},
	}
	return json.Marshal(doc)
}

