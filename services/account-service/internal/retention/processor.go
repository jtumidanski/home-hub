package retention

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	sharedretention "github.com/jtumidanski/home-hub/shared/go/retention"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrUnknownCategory = errors.New("retention: unknown category")
	ErrInvalidDays     = errors.New("retention: invalid days")
	ErrNotAuthorized   = errors.New("retention: not authorized")
	ErrScopeMismatch   = errors.New("retention: category does not match scope")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// ResolvePolicy returns the fully-resolved policy for both household and user
// scopes, merging defaults with overrides.
type ResolvedPolicy struct {
	Household   *ResolvedScope
	UserScope   *ResolvedScope
}

type ResolvedScope struct {
	ScopeKind sharedretention.ScopeKind
	ScopeId   uuid.UUID
	Values    map[sharedretention.Category]Resolved
}

type Resolved struct {
	Days   int
	Source string
}

// LoadOverrides returns the override map for one scope. Used by the resolved
// policy and by the internal /internal/retention-policies/overrides endpoint.
func (p *Processor) LoadOverrides(tenantID uuid.UUID, scopeKind sharedretention.ScopeKind, scopeID uuid.UUID) (map[sharedretention.Category]int, error) {
	rows, err := listOverrides(p.db.WithContext(p.ctx), tenantID, scopeKind, scopeID)
	if err != nil {
		return nil, err
	}
	out := make(map[sharedretention.Category]int, len(rows))
	for _, r := range rows {
		out[sharedretention.Category(r.Category)] = r.RetentionDays
	}
	return out, nil
}

// ResolveAll builds the resolved policy structure for a (household, user) pair.
// householdID may be uuid.Nil to skip household resolution.
func (p *Processor) ResolveAll(tenantID, householdID, userID uuid.UUID) (ResolvedPolicy, error) {
	out := ResolvedPolicy{}

	if householdID != uuid.Nil {
		hh, err := p.LoadOverrides(tenantID, sharedretention.ScopeHousehold, householdID)
		if err != nil {
			return ResolvedPolicy{}, err
		}
		out.Household = &ResolvedScope{
			ScopeKind: sharedretention.ScopeHousehold,
			ScopeId:   householdID,
			Values:    mergeWithDefaults(hh, sharedretention.HouseholdCategories(), "household"),
		}
	}
	if userID != uuid.Nil {
		us, err := p.LoadOverrides(tenantID, sharedretention.ScopeUser, userID)
		if err != nil {
			return ResolvedPolicy{}, err
		}
		out.UserScope = &ResolvedScope{
			ScopeKind: sharedretention.ScopeUser,
			ScopeId:   userID,
			Values:    mergeWithDefaults(us, sharedretention.UserCategories(), "user"),
		}
	}
	return out, nil
}

func mergeWithDefaults(overrides map[sharedretention.Category]int, cats []sharedretention.Category, source string) map[sharedretention.Category]Resolved {
	out := make(map[sharedretention.Category]Resolved, len(cats))
	for _, c := range cats {
		if d, ok := overrides[c]; ok {
			out[c] = Resolved{Days: d, Source: source}
			continue
		}
		out[c] = Resolved{Days: sharedretention.Defaults[c], Source: "default"}
	}
	return out
}

// ApplyHouseholdPatch validates and applies a sparse map of category → *int.
// A nil pointer means "delete the override". Authorization (caller is owner/admin
// of the household) must be checked before calling.
func (p *Processor) ApplyHouseholdPatch(tenantID, householdID uuid.UUID, patch map[sharedretention.Category]*int) error {
	for cat, days := range patch {
		if !cat.IsHouseholdScoped() {
			return ErrScopeMismatch
		}
		if days == nil {
			if err := deleteOverride(p.db.WithContext(p.ctx), tenantID, sharedretention.ScopeHousehold, householdID, cat); err != nil {
				return err
			}
			continue
		}
		if err := cat.Validate(*days); err != nil {
			return ErrInvalidDays
		}
		if _, err := upsertOverride(p.db.WithContext(p.ctx), tenantID, sharedretention.ScopeHousehold, householdID, cat, *days); err != nil {
			return err
		}
	}
	return nil
}

// ApplyUserPatch is the user-scoped equivalent of ApplyHouseholdPatch.
func (p *Processor) ApplyUserPatch(tenantID, userID uuid.UUID, patch map[sharedretention.Category]*int) error {
	for cat, days := range patch {
		if !cat.IsUserScoped() {
			return ErrScopeMismatch
		}
		if days == nil {
			if err := deleteOverride(p.db.WithContext(p.ctx), tenantID, sharedretention.ScopeUser, userID, cat); err != nil {
				return err
			}
			continue
		}
		if err := cat.Validate(*days); err != nil {
			return ErrInvalidDays
		}
		if _, err := upsertOverride(p.db.WithContext(p.ctx), tenantID, sharedretention.ScopeUser, userID, cat, *days); err != nil {
			return err
		}
	}
	return nil
}

// IsHouseholdAdmin reports whether the user has owner or admin role in the
// given household. Used to authorize household-scoped writes.
func (p *Processor) IsHouseholdAdmin(tenantID, householdID, userID uuid.UUID) (bool, error) {
	mp := membership.NewProcessor(p.l, p.ctx, p.db)
	m, err := mp.ByHouseholdAndUserProvider(householdID, userID)()
	if err != nil {
		return false, err
	}
	r := m.Role()
	return r == "owner" || r == "admin", nil
}
