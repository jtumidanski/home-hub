package month

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
)

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

func TransformMonthSummary(summary MonthSummary, items []trackingitem.Model, entries []entry.Model, snapshotsByItem map[uuid.UUID][]schedule.Model) json.RawMessage {
	monthStart, _ := time.Parse("2006-01", summary.Month)

	var itemInfos []MonthItemInfo
	for _, item := range items {
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
		for _, snap := range snapshotsByItem[item.Id()] {
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

	type entryRest struct {
		Id             uuid.UUID       `json:"id"`
		Type           string          `json:"type"`
		TrackingItemId uuid.UUID       `json:"tracking_item_id"`
		Date           string          `json:"date"`
		Value          json.RawMessage `json:"value"`
		Skipped        bool            `json:"skipped"`
		Note           *string         `json:"note,omitempty"`
		Scheduled      bool            `json:"scheduled"`
	}

	var entryRests []entryRest
	for _, e := range entries {
		entryRests = append(entryRests, entryRest{
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

	result := struct {
		Data struct {
			Type       string          `json:"type"`
			Attributes MonthSummary    `json:"attributes"`
			Relationships struct {
				Items struct {
					Data []MonthItemInfo `json:"data"`
				} `json:"items"`
				Entries struct {
					Data interface{} `json:"data"`
				} `json:"entries"`
			} `json:"relationships"`
		} `json:"data"`
	}{}

	result.Data.Type = "tracker-months"
	result.Data.Attributes = summary
	result.Data.Relationships.Items.Data = itemInfos
	if len(entryRests) > 0 {
		result.Data.Relationships.Entries.Data = entryRests
	} else {
		result.Data.Relationships.Entries.Data = []entryRest{}
	}

	b, _ := json.Marshal(result)
	return b
}
