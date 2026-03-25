package tenant

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id        uuid.UUID
	name      string
	createdAt time.Time
	updatedAt time.Time
}

func (m Model) Id() uuid.UUID       { return m.id }
func (m Model) Name() string        { return m.name }
func (m Model) CreatedAt() time.Time { return m.createdAt }
func (m Model) UpdatedAt() time.Time { return m.updatedAt }
