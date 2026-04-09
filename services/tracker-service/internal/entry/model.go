package entry

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id             uuid.UUID
	tenantID       uuid.UUID
	userID         uuid.UUID
	trackingItemID uuid.UUID
	date           time.Time
	value          json.RawMessage
	skipped        bool
	note           *string
	createdAt      time.Time
	updatedAt      time.Time
}

func (m Model) Id() uuid.UUID              { return m.id }
func (m Model) TenantID() uuid.UUID        { return m.tenantID }
func (m Model) UserID() uuid.UUID          { return m.userID }
func (m Model) TrackingItemID() uuid.UUID   { return m.trackingItemID }
func (m Model) Date() time.Time             { return m.date }
func (m Model) Value() json.RawMessage      { return m.value }
func (m Model) Skipped() bool               { return m.skipped }
func (m Model) Note() *string               { return m.note }
func (m Model) CreatedAt() time.Time        { return m.createdAt }
func (m Model) UpdatedAt() time.Time        { return m.updatedAt }
