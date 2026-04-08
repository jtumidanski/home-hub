package entry

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id             uuid.UUID       `json:"-"`
	TrackingItemId uuid.UUID       `json:"tracking_item_id"`
	Date           string          `json:"date"`
	Value          json.RawMessage `json:"value"`
	Skipped        bool            `json:"skipped"`
	Note           *string         `json:"note,omitempty"`
	Scheduled      bool            `json:"scheduled"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (r RestModel) GetName() string       { return "tracker-entries" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type EntryRequest struct {
	Id    uuid.UUID `json:"-"`
	Value json.RawMessage `json:"value"`
	Note  *string         `json:"note,omitempty"`
}

func (r EntryRequest) GetName() string       { return "tracker-entries" }
func (r EntryRequest) GetID() string          { return r.Id.String() }
func (r *EntryRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model, scheduled bool) RestModel {
	return RestModel{
		Id:             m.Id(),
		TrackingItemId: m.TrackingItemID(),
		Date:           m.Date().Format("2006-01-02"),
		Value:          m.Value(),
		Skipped:        m.Skipped(),
		Note:           m.Note(),
		Scheduled:      scheduled,
		CreatedAt:      m.CreatedAt(),
		UpdatedAt:      m.UpdatedAt(),
	}
}

// TransformSlice projects a list of WithScheduled rows into REST models. List
// handlers must use this rather than inlining a transform loop so the slice
// shape stays consistent across endpoints.
func TransformSlice(rows []WithScheduled) []*RestModel {
	rest := make([]*RestModel, len(rows))
	for i, r := range rows {
		rm := Transform(r.Entry, r.Scheduled)
		rest[i] = &rm
	}
	return rest
}
