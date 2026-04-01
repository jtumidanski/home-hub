package list

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id           uuid.UUID
	tenantID     uuid.UUID
	householdID  uuid.UUID
	name         string
	status       string
	archivedAt   *time.Time
	createdBy    uuid.UUID
	itemCount    int
	checkedCount int
	createdAt    time.Time
	updatedAt    time.Time
}

func (m Model) Id() uuid.UUID        { return m.id }
func (m Model) TenantID() uuid.UUID  { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) Name() string          { return m.name }
func (m Model) Status() string        { return m.status }
func (m Model) ArchivedAt() *time.Time { return m.archivedAt }
func (m Model) CreatedBy() uuid.UUID  { return m.createdBy }
func (m Model) ItemCount() int        { return m.itemCount }
func (m Model) CheckedCount() int     { return m.checkedCount }
func (m Model) CreatedAt() time.Time  { return m.createdAt }
func (m Model) UpdatedAt() time.Time  { return m.updatedAt }

func (m Model) IsArchived() bool { return m.status == "archived" }

func (m Model) WithCounts(itemCount, checkedCount int) Model {
	m.itemCount = itemCount
	m.checkedCount = checkedCount
	return m
}
