package householdpreference

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id                   uuid.UUID  `json:"-"`
	DefaultDashboardId   *uuid.UUID `json:"default_dashboard_id"`
	KioskDashboardSeeded bool       `json:"kiosk_dashboard_seeded"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func (r RestModel) GetName() string { return "householdPreferences" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:                   m.Id(),
		DefaultDashboardId:   m.DefaultDashboardID(),
		KioskDashboardSeeded: m.KioskDashboardSeeded(),
		CreatedAt:            m.CreatedAt(),
		UpdatedAt:            m.UpdatedAt(),
	}, nil
}

// UpdateRequest carries patchable fields for a household preference.
//
// PATCH semantics caveat: Go's stdlib JSON (and api2go's unmarshaling) decode
// both JSON null and a missing attribute to a nil *uuid.UUID. Because this
// resource has only one mutable attribute today, we treat nil as "clear the
// field" — which also means PATCHing with an empty attributes object clears
// it. If a future resource adds more patchable attributes, switch to a custom
// JSON unmarshaler (for example *pointer or a struct with an IsSet flag) so
// absent-vs-explicit-null can be distinguished.
type UpdateRequest struct {
	Id                 uuid.UUID  `json:"-"`
	DefaultDashboardId *uuid.UUID `json:"default_dashboard_id,omitempty"`
}

func (r UpdateRequest) GetName() string { return "householdPreferences" }
func (r UpdateRequest) GetID() string   { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
