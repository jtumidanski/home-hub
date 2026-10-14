package recipeimage

import (
	"time"

	"github.com/google/uuid"
)

type Size string

const (
	SizeFull  Size = "full"
	SizeThumb Size = "thumb"
)

// ParseSize maps the ?size= query value. Empty means full.
func ParseSize(s string) (Size, bool) {
	switch s {
	case "", string(SizeFull):
		return SizeFull, true
	case string(SizeThumb):
		return SizeThumb, true
	default:
		return "", false
	}
}

type Model struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	recipeID    uuid.UUID
	fullKey     string
	thumbKey    string
	contentType string
	width       int
	height      int
	fullBytes   int
	thumbBytes  int
	createdAt   time.Time
	updatedAt   time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) RecipeID() uuid.UUID    { return m.recipeID }
func (m Model) FullKey() string        { return m.fullKey }
func (m Model) ThumbKey() string       { return m.thumbKey }
func (m Model) ContentType() string    { return m.contentType }
func (m Model) Width() int             { return m.width }
func (m Model) Height() int            { return m.height }
func (m Model) FullBytes() int         { return m.fullBytes }
func (m Model) ThumbBytes() int        { return m.thumbBytes }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }

func (m Model) Key(size Size) string {
	if size == SizeThumb {
		return m.thumbKey
	}
	return m.fullKey
}
