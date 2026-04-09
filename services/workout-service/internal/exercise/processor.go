package exercise

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/region"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/theme"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound        = errors.New("exercise not found")
	ErrDuplicateName   = errors.New("exercise name already exists for this user")
	ErrThemeNotFound   = errors.New("theme not found")
	ErrRegionNotFound  = errors.New("region not found")
	ErrSecondaryNotFound = errors.New("secondary region not found")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// Defaults groups the kind-specific default fields. Pointers preserve "absent
// vs explicit zero" semantics so we can validate which combinations are
// meaningful for the chosen kind.
type Defaults struct {
	Sets            *int     `json:"sets,omitempty"`
	Reps            *int     `json:"reps,omitempty"`
	Weight          *float64 `json:"weight,omitempty"`
	WeightUnit      *string  `json:"weightUnit,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	Distance        *float64 `json:"distance,omitempty"`
	DistanceUnit    *string  `json:"distanceUnit,omitempty"`
}

type CreateInput struct {
	Name               string
	Kind               string
	WeightType         string
	ThemeID            uuid.UUID
	RegionID           uuid.UUID
	SecondaryRegionIDs []uuid.UUID
	Defaults           Defaults
	Notes              *string
}

type UpdateInput struct {
	Name               *string
	ThemeID            *uuid.UUID
	RegionID           *uuid.UUID
	SecondaryRegionIDs *[]uuid.UUID
	Defaults           *Defaults
	Notes              *string
}

func (p *Processor) List(userID uuid.UUID, themeID, regionID *uuid.UUID) ([]Model, error) {
	rows, err := ListByUser(p.db.WithContext(p.ctx), userID, themeID, regionID)
	if err != nil {
		return nil, err
	}
	models := make([]Model, 0, len(rows))
	for _, e := range rows {
		m, err := Make(e)
		if err != nil {
			p.l.WithError(err).WithField("exercise_id", e.Id).Warn("Skipping unreadable exercise")
			continue
		}
		models = append(models, m)
	}
	return models, nil
}

func (p *Processor) Get(id uuid.UUID) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return Make(e)
}

func (p *Processor) Create(tenantID, userID uuid.UUID, in CreateInput) (Model, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.WeightType == "" {
		in.WeightType = WeightTypeFree
	}

	// Build the model first so the builder enforces shape, kind, weight unit,
	// distance unit, primary-not-in-secondary, and required fields. The DB
	// reference checks below assume those invariants already hold.
	mb := NewBuilder().
		SetTenantID(tenantID).
		SetUserID(userID).
		SetName(in.Name).
		SetKind(in.Kind).
		SetWeightType(in.WeightType).
		SetThemeID(in.ThemeID).
		SetRegionID(in.RegionID).
		SetSecondaryRegionIDs(in.SecondaryRegionIDs).
		SetDefaultSets(in.Defaults.Sets).
		SetDefaultReps(in.Defaults.Reps).
		SetDefaultWeight(in.Defaults.Weight).
		SetDefaultWeightUnit(in.Defaults.WeightUnit).
		SetDefaultDurationSeconds(in.Defaults.DurationSeconds).
		SetDefaultDistance(in.Defaults.Distance).
		SetDefaultDistanceUnit(in.Defaults.DistanceUnit).
		SetNotes(in.Notes)

	if _, err := mb.Build(); err != nil {
		return Model{}, err
	}

	if err := p.checkReferences(userID, in.ThemeID, in.RegionID, in.SecondaryRegionIDs); err != nil {
		return Model{}, err
	}

	if _, err := GetByName(userID, in.Name)(p.db.WithContext(p.ctx))(); err == nil {
		return Model{}, ErrDuplicateName
	}

	secondaryJSON, _ := json.Marshal(uuidsToStrings(in.SecondaryRegionIDs))
	e := Entity{
		TenantId:               tenantID,
		UserId:                 userID,
		Name:                   in.Name,
		Kind:                   in.Kind,
		WeightType:             in.WeightType,
		ThemeId:                in.ThemeID,
		RegionId:               in.RegionID,
		SecondaryRegionIds:     secondaryJSON,
		DefaultSets:            in.Defaults.Sets,
		DefaultReps:            in.Defaults.Reps,
		DefaultWeight:          in.Defaults.Weight,
		DefaultWeightUnit:      in.Defaults.WeightUnit,
		DefaultDurationSeconds: in.Defaults.DurationSeconds,
		DefaultDistance:        in.Defaults.Distance,
		DefaultDistanceUnit:    in.Defaults.DistanceUnit,
		Notes:                  in.Notes,
	}
	if err := createExercise(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(id uuid.UUID, in UpdateInput) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}

	if in.Name != nil {
		trimmed := strings.TrimSpace(*in.Name)
		if trimmed == "" {
			return Model{}, ErrNameRequired
		}
		if len(trimmed) > 100 {
			return Model{}, ErrNameTooLong
		}
		if existing, err := GetByName(e.UserId, trimmed)(p.db.WithContext(p.ctx))(); err == nil && existing.Id != id {
			return Model{}, ErrDuplicateName
		}
		e.Name = trimmed
	}
	if in.ThemeID != nil {
		if _, err := theme.GetByID(*in.ThemeID)(p.db.WithContext(p.ctx))(); err != nil {
			return Model{}, ErrThemeNotFound
		}
		e.ThemeId = *in.ThemeID
	}
	if in.RegionID != nil {
		if _, err := region.GetByID(*in.RegionID)(p.db.WithContext(p.ctx))(); err != nil {
			return Model{}, ErrRegionNotFound
		}
		e.RegionId = *in.RegionID
	}
	if in.SecondaryRegionIDs != nil {
		// Validate all secondary regions exist and the primary is not among them.
		for _, sid := range *in.SecondaryRegionIDs {
			if sid == e.RegionId {
				return Model{}, ErrPrimaryInSecondary
			}
			if _, err := region.GetByID(sid)(p.db.WithContext(p.ctx))(); err != nil {
				return Model{}, ErrSecondaryNotFound
			}
		}
		j, _ := json.Marshal(uuidsToStrings(*in.SecondaryRegionIDs))
		e.SecondaryRegionIds = j
	}
	if in.Defaults != nil {
		// Round-trip the merged entity through the builder so all defaults-shape
		// rules apply consistently — defaults belonging to the wrong kind are
		// rejected here, not in storage.
		e.DefaultSets = in.Defaults.Sets
		e.DefaultReps = in.Defaults.Reps
		e.DefaultWeight = in.Defaults.Weight
		e.DefaultWeightUnit = in.Defaults.WeightUnit
		e.DefaultDurationSeconds = in.Defaults.DurationSeconds
		e.DefaultDistance = in.Defaults.Distance
		e.DefaultDistanceUnit = in.Defaults.DistanceUnit
	}
	if in.Notes != nil {
		e.Notes = in.Notes
	}

	// Final validation through the builder using the merged state.
	if _, err := Make(e); err != nil {
		return Model{}, err
	}

	if err := updateExercise(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return ErrNotFound
	}
	return softDeleteExercise(p.db.WithContext(p.ctx), &e)
}

// checkReferences verifies that the theme, primary region, and every secondary
// region referenced by an exercise actually exist and are owned by the user.
// The tenant callback handles tenant scoping; we re-check user_id explicitly
// because soft-deleted parents are still allowed for historical reads but not
// for new exercise creation.
func (p *Processor) checkReferences(userID uuid.UUID, themeID, regionID uuid.UUID, secondary []uuid.UUID) error {
	tdb := p.db.WithContext(p.ctx)
	if t, err := theme.GetByID(themeID)(tdb)(); err != nil || t.UserId != userID {
		return ErrThemeNotFound
	}
	if r, err := region.GetByID(regionID)(tdb)(); err != nil || r.UserId != userID {
		return ErrRegionNotFound
	}
	for _, sid := range secondary {
		r, err := region.GetByID(sid)(tdb)()
		if err != nil || r.UserId != userID {
			return ErrSecondaryNotFound
		}
	}
	return nil
}
