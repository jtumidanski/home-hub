package trackingitem

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
)

type ScheduleHistoryEntry struct {
	Schedule      []int  `json:"schedule"`
	EffectiveDate string `json:"effective_date"`
}

// ItemWithSchedule pairs a tracking item with its current effective schedule.
// The list handler uses this paired view so TransformSlice can produce a
// REST projection without per-row processor calls in the resource layer.
type ItemWithSchedule struct {
	Item     Model
	Schedule []int
}

type RestModel struct {
	Id              uuid.UUID              `json:"-"`
	Name            string                 `json:"name"`
	ScaleType       string                 `json:"scale_type"`
	ScaleConfig     json.RawMessage        `json:"scale_config"`
	Schedule        []int                  `json:"schedule"`
	Color           string                 `json:"color"`
	SortOrder       int                    `json:"sort_order"`
	ScheduleHistory []ScheduleHistoryEntry `json:"schedule_history,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func (r RestModel) GetName() string       { return "trackers" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type CreateRequest struct {
	Id          uuid.UUID       `json:"-"`
	Name        string          `json:"name"`
	ScaleType   string          `json:"scale_type"`
	ScaleConfig json.RawMessage `json:"scale_config"`
	Schedule    []int           `json:"schedule"`
	Color       string          `json:"color"`
	SortOrder   int             `json:"sort_order"`
}

func (r CreateRequest) GetName() string       { return "trackers" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type UpdateRequest struct {
	Id          uuid.UUID        `json:"-"`
	Name        string           `json:"name,omitempty"`
	ScaleType   string           `json:"scale_type,omitempty"`
	ScaleConfig *json.RawMessage `json:"scale_config,omitempty"`
	Schedule    *[]int           `json:"schedule,omitempty"`
	Color       string           `json:"color,omitempty"`
	SortOrder   *int             `json:"sort_order,omitempty"`
}

func (r UpdateRequest) GetName() string       { return "trackers" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model, currentSchedule []int, history []schedule.Model) RestModel {
	var historyEntries []ScheduleHistoryEntry
	for _, h := range history {
		historyEntries = append(historyEntries, ScheduleHistoryEntry{
			Schedule:      h.Schedule(),
			EffectiveDate: h.EffectiveDate().Format("2006-01-02"),
		})
	}

	return RestModel{
		Id:              m.Id(),
		Name:            m.Name(),
		ScaleType:       m.ScaleType(),
		ScaleConfig:     m.ScaleConfig(),
		Schedule:        currentSchedule,
		Color:           m.Color(),
		SortOrder:       m.SortOrder(),
		ScheduleHistory: historyEntries,
		CreatedAt:       m.CreatedAt(),
		UpdatedAt:       m.UpdatedAt(),
	}
}

// TransformSlice projects a slice of items + their current schedules into REST
// models. List handlers must use this rather than inlining a transform loop.
func TransformSlice(rows []ItemWithSchedule) []*RestModel {
	rest := make([]*RestModel, len(rows))
	for i, r := range rows {
		rm := TransformList(r.Item, r.Schedule)
		rest[i] = &rm
	}
	return rest
}

func TransformList(m Model, currentSchedule []int) RestModel {
	return RestModel{
		Id:          m.Id(),
		Name:        m.Name(),
		ScaleType:   m.ScaleType(),
		ScaleConfig: m.ScaleConfig(),
		Schedule:    currentSchedule,
		Color:       m.Color(),
		SortOrder:   m.SortOrder(),
		CreatedAt:   m.CreatedAt(),
		UpdatedAt:   m.UpdatedAt(),
	}
}
