package today

import (
	"encoding/json"

	"github.com/google/uuid"
)

// The "today" view is a composite read that mixes two resource types
// (trackers + tracker-entries) under a single envelope keyed by date. api2go's
// MarshalIdentifier helpers cannot represent this exact shape (they would
// split related resources into the top-level `included` array, not into
// per-relationship `data` arrays), so the rest layer hand-rolls the JSON:API
// document. The handler stays free of envelope assembly — it only invokes
// MarshalDocument and forwards the bytes.

// ItemView is the per-item projection for today.
type ItemView struct {
	Id          uuid.UUID       `json:"id"`
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	ScaleType   string          `json:"scale_type"`
	ScaleConfig json.RawMessage `json:"scale_config"`
	Color       string          `json:"color"`
	SortOrder   int             `json:"sort_order"`
}

// EntryView is the per-entry projection for today. The "scheduled" flag is
// always true here because the today view, by definition, only includes
// entries on scheduled days.
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

// Document is the typed JSON:API document the today endpoint emits.
type Document struct {
	Data data `json:"data"`
}

type data struct {
	Type          string        `json:"type"`
	Attributes    attributes    `json:"attributes"`
	Relationships relationships `json:"relationships"`
}

type attributes struct {
	Date string `json:"date"`
}

type relationships struct {
	Items   relItems   `json:"items"`
	Entries relEntries `json:"entries"`
}

type relItems struct {
	Data []ItemView `json:"data"`
}

type relEntries struct {
	Data []EntryView `json:"data"`
}

// Transform projects a processor Result into the typed JSON:API document.
func Transform(r Result) Document {
	items := make([]ItemView, 0, len(r.Items))
	for _, m := range r.Items {
		items = append(items, ItemView{
			Id:          m.Id(),
			Type:        "trackers",
			Name:        m.Name(),
			ScaleType:   m.ScaleType(),
			ScaleConfig: m.ScaleConfig(),
			Color:       m.Color(),
			SortOrder:   m.SortOrder(),
		})
	}

	entries := make([]EntryView, 0, len(r.Entries))
	for _, e := range r.Entries {
		entries = append(entries, EntryView{
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

	return Document{
		Data: data{
			Type:       "tracker-today",
			Attributes: attributes{Date: r.Date.Format("2006-01-02")},
			Relationships: relationships{
				Items:   relItems{Data: items},
				Entries: relEntries{Data: entries},
			},
		},
	}
}

// MarshalDocument is a thin wrapper around json.Marshal that lets the resource
// layer surface marshal errors uniformly.
func MarshalDocument(doc Document) ([]byte, error) {
	return json.Marshal(doc)
}
