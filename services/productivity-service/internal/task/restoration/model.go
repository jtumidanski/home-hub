package restoration

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	taskID          uuid.UUID
	createdByUserID uuid.UUID
	createdAt       time.Time
}

func (m Model) Id() uuid.UUID               { return m.id }
func (m Model) TenantID() uuid.UUID          { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID       { return m.householdID }
func (m Model) TaskID() uuid.UUID            { return m.taskID }
func (m Model) CreatedByUserID() uuid.UUID   { return m.createdByUserID }
func (m Model) CreatedAt() time.Time         { return m.createdAt }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:              m.id,
		TenantId:        m.tenantID,
		HouseholdId:     m.householdID,
		TaskId:          m.taskID,
		CreatedByUserId: m.createdByUserID,
		CreatedAt:       m.createdAt,
	}
}
