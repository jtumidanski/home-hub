package recipe

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	title           string
	description     string
	source          string
	servings        *int
	prepTimeMinutes *int
	cookTimeMinutes *int
	sourceURL       string
	tags            []string
	deletedAt       *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

func (m Model) Id() uuid.UUID          { return m.id }
func (m Model) TenantID() uuid.UUID    { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID { return m.householdID }
func (m Model) Title() string          { return m.title }
func (m Model) Description() string    { return m.description }
func (m Model) Source() string         { return m.source }
func (m Model) Servings() *int         { return m.servings }
func (m Model) PrepTimeMinutes() *int  { return m.prepTimeMinutes }
func (m Model) CookTimeMinutes() *int  { return m.cookTimeMinutes }
func (m Model) SourceURL() string      { return m.sourceURL }
func (m Model) Tags() []string         { return m.tags }
func (m Model) DeletedAt() *time.Time  { return m.deletedAt }
func (m Model) CreatedAt() time.Time   { return m.createdAt }
func (m Model) UpdatedAt() time.Time   { return m.updatedAt }
func (m Model) IsDeleted() bool        { return m.deletedAt != nil }
