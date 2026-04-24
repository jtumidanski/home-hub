package dashboard

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id            uuid.UUID       `json:"-"`
	Name          string          `json:"name"`
	Scope         string          `json:"scope"`
	SortOrder     int             `json:"sortOrder"`
	Layout        json.RawMessage `json:"layout"`
	SchemaVersion int             `json:"schemaVersion"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "dashboards" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	v, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	r.Id = v
	return nil
}

type CreateRequest struct {
	Name      string          `json:"name"`
	Scope     string          `json:"scope"`
	Layout    json.RawMessage `json:"layout"`
	SortOrder *int            `json:"sortOrder"`
}

func (CreateRequest) GetName() string      { return "dashboards" }
func (CreateRequest) GetID() string        { return "" }
func (*CreateRequest) SetID(_ string) error { return nil }

type UpdateRequest struct {
	Name      *string          `json:"name"`
	Layout    *json.RawMessage `json:"layout"`
	SortOrder *int             `json:"sortOrder"`
}

func (UpdateRequest) GetName() string      { return "dashboards" }
func (UpdateRequest) GetID() string        { return "" }
func (*UpdateRequest) SetID(_ string) error { return nil }

type ReorderRequest struct {
	// Because api2go's input decoder expects a single resource, bulk reorder
	// is posted as a plain JSON body (non-JSON:API). See resource.go for the
	// custom handler wiring.
	Entries []ReorderEntry `json:"data"`
}

type ReorderEntry struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sortOrder"`
}

type SeedRequest struct {
	Name   string          `json:"name"`
	Layout json.RawMessage `json:"layout"`
}

func (SeedRequest) GetName() string      { return "dashboards" }
func (SeedRequest) GetID() string        { return "" }
func (*SeedRequest) SetID(_ string) error { return nil }

func Transform(m Model) (RestModel, error) {
	scope := "household"
	if m.UserID() != nil {
		scope = "user"
	}
	return RestModel{
		Id:            m.Id(),
		Name:          m.Name(),
		Scope:         scope,
		SortOrder:     m.SortOrder(),
		Layout:        json.RawMessage(m.LayoutJSON()),
		SchemaVersion: m.SchemaVersion(),
		CreatedAt:     m.CreatedAt(),
		UpdatedAt:     m.UpdatedAt(),
	}, nil
}
