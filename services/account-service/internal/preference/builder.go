package preference

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserIDRequired = errors.New("preference user ID is required")
	ErrThemeRequired  = errors.New("preference theme is required")
)

type Builder struct {
	id                uuid.UUID
	tenantID          uuid.UUID
	userID            uuid.UUID
	theme             string
	activeHouseholdID *uuid.UUID
	createdAt         time.Time
	updatedAt         time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = id
	return b
}

func (b *Builder) SetTenantID(tenantID uuid.UUID) *Builder {
	b.tenantID = tenantID
	return b
}

func (b *Builder) SetUserID(userID uuid.UUID) *Builder {
	b.userID = userID
	return b
}

func (b *Builder) SetTheme(theme string) *Builder {
	b.theme = theme
	return b
}

func (b *Builder) SetActiveHouseholdID(id *uuid.UUID) *Builder {
	b.activeHouseholdID = id
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
	if b.userID == uuid.Nil {
		return Model{}, ErrUserIDRequired
	}
	if b.theme == "" {
		return Model{}, ErrThemeRequired
	}
	return Model{
		id:                b.id,
		tenantID:          b.tenantID,
		userID:            b.userID,
		theme:             b.theme,
		activeHouseholdID: b.activeHouseholdID,
		createdAt:         b.createdAt,
		updatedAt:         b.updatedAt,
	}, nil
}
