package dashboard

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Model struct {
	id            uuid.UUID
	tenantID      uuid.UUID
	householdID   uuid.UUID
	userID        *uuid.UUID
	name          string
	sortOrder     int
	layout        datatypes.JSON
	schemaVersion int
	createdAt     time.Time
	updatedAt     time.Time
}

func (m Model) Id() uuid.UUID            { return m.id }
func (m Model) TenantID() uuid.UUID      { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID   { return m.householdID }
func (m Model) UserID() *uuid.UUID       { return m.userID }
func (m Model) Name() string             { return m.name }
func (m Model) SortOrder() int           { return m.sortOrder }
func (m Model) Layout() datatypes.JSON   { return m.layout }
func (m Model) SchemaVersion() int       { return m.schemaVersion }
func (m Model) CreatedAt() time.Time     { return m.createdAt }
func (m Model) UpdatedAt() time.Time     { return m.updatedAt }

// IsHouseholdScoped reports whether this dashboard is shared across the household
// (i.e., not owned by a specific user).
func (m Model) IsHouseholdScoped() bool {
	return m.userID == nil
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:            m.id,
		TenantId:      m.tenantID,
		HouseholdId:   m.householdID,
		UserId:        m.userID,
		Name:          m.name,
		SortOrder:     m.sortOrder,
		Layout:        m.layout,
		SchemaVersion: m.schemaVersion,
		CreatedAt:     m.createdAt,
		UpdatedAt:     m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetUserID(e.UserId).
		SetName(e.Name).
		SetSortOrder(e.SortOrder).
		SetLayout(e.Layout).
		SetSchemaVersion(e.SchemaVersion).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
