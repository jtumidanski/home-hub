package normalization

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusMatched            Status = "matched"
	StatusAliasMatched       Status = "alias_matched"
	StatusUnresolved         Status = "unresolved"
	StatusManuallyConfirmed  Status = "manually_confirmed"
)

func ValidStatus(s string) bool {
	switch Status(s) {
	case StatusMatched, StatusAliasMatched, StatusUnresolved, StatusManuallyConfirmed:
		return true
	default:
		return false
	}
}

type Model struct {
	id                    uuid.UUID
	tenantID              uuid.UUID
	householdID           uuid.UUID
	recipeID              uuid.UUID
	rawName               string
	rawQuantity           string
	rawUnit               string
	position              int
	canonicalIngredientID *uuid.UUID
	canonicalUnit         string
	normalizationStatus   Status
	createdAt             time.Time
	updatedAt             time.Time
}

func (m Model) Id() uuid.UUID                    { return m.id }
func (m Model) TenantID() uuid.UUID              { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID           { return m.householdID }
func (m Model) RecipeID() uuid.UUID              { return m.recipeID }
func (m Model) RawName() string                  { return m.rawName }
func (m Model) RawQuantity() string              { return m.rawQuantity }
func (m Model) RawUnit() string                  { return m.rawUnit }
func (m Model) Position() int                    { return m.position }
func (m Model) CanonicalIngredientID() *uuid.UUID { return m.canonicalIngredientID }
func (m Model) CanonicalUnit() string            { return m.canonicalUnit }
func (m Model) NormalizationStatus() Status      { return m.normalizationStatus }
func (m Model) CreatedAt() time.Time             { return m.createdAt }
func (m Model) UpdatedAt() time.Time             { return m.updatedAt }
