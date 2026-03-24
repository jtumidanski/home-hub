package snooze

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	reminderID      uuid.UUID
	durationMinutes int
	snoozedUntil    time.Time
	createdByUserID uuid.UUID
	createdAt       time.Time
}

func (m Model) Id() uuid.UUID            { return m.id }
func (m Model) TenantID() uuid.UUID      { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID   { return m.householdID }
func (m Model) ReminderID() uuid.UUID    { return m.reminderID }
func (m Model) DurationMinutes() int     { return m.durationMinutes }
func (m Model) SnoozedUntil() time.Time  { return m.snoozedUntil }
func (m Model) CreatedByUserID() uuid.UUID { return m.createdByUserID }
func (m Model) CreatedAt() time.Time     { return m.createdAt }

func (m Model) ToEntity() Entity {
	return Entity{
		Id:              m.id,
		TenantId:        m.tenantID,
		HouseholdId:     m.householdID,
		ReminderId:      m.reminderID,
		DurationMinutes: m.durationMinutes,
		SnoozedUntil:    m.snoozedUntil,
		CreatedByUserId: m.createdByUserID,
		CreatedAt:       m.createdAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetReminderID(e.ReminderId).
		SetDurationMinutes(e.DurationMinutes).
		SetSnoozedUntil(e.SnoozedUntil).
		SetCreatedByUserID(e.CreatedByUserId).
		SetCreatedAt(e.CreatedAt).
		Build()
}
