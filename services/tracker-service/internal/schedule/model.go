package schedule

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id             uuid.UUID
	trackingItemID uuid.UUID
	schedule       []int
	effectiveDate  time.Time
	createdAt      time.Time
}

func (m Model) Id() uuid.UUID             { return m.id }
func (m Model) TrackingItemID() uuid.UUID  { return m.trackingItemID }
func (m Model) Schedule() []int            { return m.schedule }
func (m Model) EffectiveDate() time.Time   { return m.effectiveDate }
func (m Model) CreatedAt() time.Time       { return m.createdAt }
