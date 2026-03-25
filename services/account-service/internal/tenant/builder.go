package tenant

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired = errors.New("tenant name is required")
)

type Builder struct {
	id        uuid.UUID
	name      string
	createdAt time.Time
	updatedAt time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = id
	return b
}

func (b *Builder) SetName(name string) *Builder {
	b.name = name
	return b
}

func (b *Builder) SetCreatedAt(t time.Time) *Builder {
	b.createdAt = t
	return b
}

func (b *Builder) SetUpdatedAt(t time.Time) *Builder {
	b.updatedAt = t
	return b
}

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	return Model{
		id:        b.id,
		name:      b.name,
		createdAt: b.createdAt,
		updatedAt: b.updatedAt,
	}, nil
}
