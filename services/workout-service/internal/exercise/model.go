package exercise

import (
	"time"

	"github.com/google/uuid"
)

const (
	KindStrength  = "strength"
	KindIsometric = "isometric"
	KindCardio    = "cardio"

	WeightTypeFree       = "free"
	WeightTypeBodyweight = "bodyweight"

	WeightUnitLb = "lb"
	WeightUnitKg = "kg"

	DistanceUnitMi = "mi"
	DistanceUnitKm = "km"
	DistanceUnitM  = "m"
)

type Model struct {
	id                     uuid.UUID
	tenantID               uuid.UUID
	userID                 uuid.UUID
	name                   string
	kind                   string
	weightType             string
	themeID                uuid.UUID
	regionID               uuid.UUID
	secondaryRegionIDs     []uuid.UUID
	defaultSets            *int
	defaultReps            *int
	defaultWeight          *float64
	defaultWeightUnit      *string
	defaultDurationSeconds *int
	defaultDistance        *float64
	defaultDistanceUnit    *string
	notes                  *string
	createdAt              time.Time
	updatedAt              time.Time
	deletedAt              *time.Time
}

func (m Model) Id() uuid.UUID                    { return m.id }
func (m Model) TenantID() uuid.UUID              { return m.tenantID }
func (m Model) UserID() uuid.UUID                { return m.userID }
func (m Model) Name() string                     { return m.name }
func (m Model) Kind() string                     { return m.kind }
func (m Model) WeightType() string               { return m.weightType }
func (m Model) ThemeID() uuid.UUID               { return m.themeID }
func (m Model) RegionID() uuid.UUID              { return m.regionID }
func (m Model) SecondaryRegionIDs() []uuid.UUID  { return m.secondaryRegionIDs }
func (m Model) DefaultSets() *int                { return m.defaultSets }
func (m Model) DefaultReps() *int                { return m.defaultReps }
func (m Model) DefaultWeight() *float64          { return m.defaultWeight }
func (m Model) DefaultWeightUnit() *string       { return m.defaultWeightUnit }
func (m Model) DefaultDurationSeconds() *int     { return m.defaultDurationSeconds }
func (m Model) DefaultDistance() *float64        { return m.defaultDistance }
func (m Model) DefaultDistanceUnit() *string     { return m.defaultDistanceUnit }
func (m Model) Notes() *string                   { return m.notes }
func (m Model) CreatedAt() time.Time             { return m.createdAt }
func (m Model) UpdatedAt() time.Time             { return m.updatedAt }
func (m Model) DeletedAt() *time.Time            { return m.deletedAt }
