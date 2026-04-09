package region

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id        uuid.UUID
	tenantID  uuid.UUID
	userID    uuid.UUID
	name      string
	sortOrder int
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func (m Model) Id() uuid.UUID        { return m.id }
func (m Model) TenantID() uuid.UUID  { return m.tenantID }
func (m Model) UserID() uuid.UUID    { return m.userID }
func (m Model) Name() string         { return m.name }
func (m Model) SortOrder() int       { return m.sortOrder }
func (m Model) CreatedAt() time.Time { return m.createdAt }
func (m Model) UpdatedAt() time.Time { return m.updatedAt }
func (m Model) DeletedAt() *time.Time { return m.deletedAt }
