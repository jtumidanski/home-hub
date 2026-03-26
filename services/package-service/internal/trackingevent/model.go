package trackingevent

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id          uuid.UUID
	packageID   uuid.UUID
	timestamp   time.Time
	status      string
	description string
	location    *string
	rawStatus   *string
	createdAt   time.Time
}

func (m Model) Id() uuid.UUID        { return m.id }
func (m Model) PackageID() uuid.UUID  { return m.packageID }
func (m Model) Timestamp() time.Time  { return m.timestamp }
func (m Model) Status() string        { return m.status }
func (m Model) Description() string   { return m.description }
func (m Model) Location() *string     { return m.location }
func (m Model) RawStatus() *string    { return m.rawStatus }
func (m Model) CreatedAt() time.Time  { return m.createdAt }
