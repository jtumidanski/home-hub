package exercise

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired         = errors.New("exercise name is required")
	ErrNameTooLong          = errors.New("exercise name must not exceed 100 characters")
	ErrInvalidKind          = errors.New("kind must be one of: strength, isometric, cardio")
	ErrInvalidWeightType    = errors.New("weightType must be one of: free, bodyweight")
	ErrInvalidWeightUnit    = errors.New("weightUnit must be one of: lb, kg")
	ErrInvalidDistanceUnit  = errors.New("distanceUnit must be one of: mi, km, m")
	ErrInvalidNumeric       = errors.New("numeric defaults must be non-negative")
	ErrPrimaryInSecondary   = errors.New("primary regionId must not appear in secondaryRegionIds")
	ErrThemeRequired        = errors.New("themeId is required")
	ErrRegionRequired       = errors.New("regionId is required")
	ErrKindImmutable        = errors.New("exercise kind cannot be changed after creation")
	ErrWeightTypeImmutable  = errors.New("exercise weightType cannot be changed after creation")
	ErrInvalidDefaultsShape = errors.New("defaults shape does not match exercise kind")
)

var validKinds = map[string]bool{
	KindStrength: true, KindIsometric: true, KindCardio: true,
}

var validWeightTypes = map[string]bool{
	WeightTypeFree: true, WeightTypeBodyweight: true,
}

var validWeightUnits = map[string]bool{WeightUnitLb: true, WeightUnitKg: true}
var validDistanceUnits = map[string]bool{DistanceUnitMi: true, DistanceUnitKm: true, DistanceUnitM: true}

func ValidWeightUnit(u string) bool   { return validWeightUnits[u] }
func ValidDistanceUnit(u string) bool { return validDistanceUnits[u] }

type Builder struct {
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

func NewBuilder() *Builder { return &Builder{weightType: WeightTypeFree} }

func (b *Builder) SetId(id uuid.UUID) *Builder                { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder          { b.tenantID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder            { b.userID = id; return b }
func (b *Builder) SetName(n string) *Builder                   { b.name = n; return b }
func (b *Builder) SetKind(k string) *Builder                   { b.kind = k; return b }
func (b *Builder) SetWeightType(w string) *Builder             { if w != "" { b.weightType = w }; return b }
func (b *Builder) SetThemeID(id uuid.UUID) *Builder           { b.themeID = id; return b }
func (b *Builder) SetRegionID(id uuid.UUID) *Builder          { b.regionID = id; return b }
func (b *Builder) SetSecondaryRegionIDs(ids []uuid.UUID) *Builder { b.secondaryRegionIDs = ids; return b }
func (b *Builder) SetDefaultSets(v *int) *Builder              { b.defaultSets = v; return b }
func (b *Builder) SetDefaultReps(v *int) *Builder              { b.defaultReps = v; return b }
func (b *Builder) SetDefaultWeight(v *float64) *Builder        { b.defaultWeight = v; return b }
func (b *Builder) SetDefaultWeightUnit(v *string) *Builder     { b.defaultWeightUnit = v; return b }
func (b *Builder) SetDefaultDurationSeconds(v *int) *Builder   { b.defaultDurationSeconds = v; return b }
func (b *Builder) SetDefaultDistance(v *float64) *Builder      { b.defaultDistance = v; return b }
func (b *Builder) SetDefaultDistanceUnit(v *string) *Builder   { b.defaultDistanceUnit = v; return b }
func (b *Builder) SetNotes(v *string) *Builder                 { b.notes = v; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder           { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder           { b.updatedAt = t; return b }
func (b *Builder) SetDeletedAt(t *time.Time) *Builder          { b.deletedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if len(b.name) > 100 {
		return Model{}, ErrNameTooLong
	}
	if !validKinds[b.kind] {
		return Model{}, ErrInvalidKind
	}
	if !validWeightTypes[b.weightType] {
		return Model{}, ErrInvalidWeightType
	}
	if b.themeID == uuid.Nil {
		return Model{}, ErrThemeRequired
	}
	if b.regionID == uuid.Nil {
		return Model{}, ErrRegionRequired
	}
	for _, sid := range b.secondaryRegionIDs {
		if sid == b.regionID {
			return Model{}, ErrPrimaryInSecondary
		}
	}
	if err := validateNumericDefaults(b); err != nil {
		return Model{}, err
	}
	if b.defaultWeightUnit != nil && *b.defaultWeightUnit != "" && !validWeightUnits[*b.defaultWeightUnit] {
		return Model{}, ErrInvalidWeightUnit
	}
	if b.defaultDistanceUnit != nil && *b.defaultDistanceUnit != "" && !validDistanceUnits[*b.defaultDistanceUnit] {
		return Model{}, ErrInvalidDistanceUnit
	}
	// Reject default fields that are nonsensical for the chosen kind. The
	// data-model lists which `default_*` columns are meaningful per kind;
	// rejecting at the model boundary prevents unreachable defaults from
	// silently surviving in storage.
	if err := validateDefaultsShape(b); err != nil {
		return Model{}, err
	}
	return Model{
		id:                     b.id,
		tenantID:               b.tenantID,
		userID:                 b.userID,
		name:                   b.name,
		kind:                   b.kind,
		weightType:             b.weightType,
		themeID:                b.themeID,
		regionID:               b.regionID,
		secondaryRegionIDs:     b.secondaryRegionIDs,
		defaultSets:            b.defaultSets,
		defaultReps:            b.defaultReps,
		defaultWeight:          b.defaultWeight,
		defaultWeightUnit:      b.defaultWeightUnit,
		defaultDurationSeconds: b.defaultDurationSeconds,
		defaultDistance:        b.defaultDistance,
		defaultDistanceUnit:    b.defaultDistanceUnit,
		notes:                  b.notes,
		createdAt:              b.createdAt,
		updatedAt:              b.updatedAt,
		deletedAt:              b.deletedAt,
	}, nil
}

func validateNumericDefaults(b *Builder) error {
	if b.defaultSets != nil && *b.defaultSets < 0 {
		return ErrInvalidNumeric
	}
	if b.defaultReps != nil && *b.defaultReps < 0 {
		return ErrInvalidNumeric
	}
	if b.defaultWeight != nil && *b.defaultWeight < 0 {
		return ErrInvalidNumeric
	}
	if b.defaultDurationSeconds != nil && *b.defaultDurationSeconds < 0 {
		return ErrInvalidNumeric
	}
	if b.defaultDistance != nil && *b.defaultDistance < 0 {
		return ErrInvalidNumeric
	}
	return nil
}

func validateDefaultsShape(b *Builder) error {
	switch b.kind {
	case KindStrength:
		if b.defaultDurationSeconds != nil || b.defaultDistance != nil || b.defaultDistanceUnit != nil {
			return ErrInvalidDefaultsShape
		}
	case KindIsometric:
		if b.defaultReps != nil || b.defaultDistance != nil || b.defaultDistanceUnit != nil {
			return ErrInvalidDefaultsShape
		}
	case KindCardio:
		if b.defaultSets != nil || b.defaultReps != nil || b.defaultWeight != nil || b.defaultWeightUnit != nil {
			return ErrInvalidDefaultsShape
		}
	}
	return nil
}
