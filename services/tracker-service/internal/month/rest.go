package month

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// The month summary endpoint returns a composite document — a single resource
// with two heterogeneous relationship arrays (items + entries). api2go cannot
// represent this exact shape natively, so the rest layer assembles a typed
// JSON:API document. The handler stays free of envelope construction; it only
// invokes Marshal* and forwards bytes (or surfaces a marshal error).

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

