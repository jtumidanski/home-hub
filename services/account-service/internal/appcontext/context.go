package appcontext

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/services/account-service/internal/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"context"
)

// Resolved holds the fully resolved application context for a user.
type Resolved struct {
	Tenant             tenant.Model
	ActiveHousehold    *household.Model
	Preference         preference.Model
	Memberships        []membership.Model
	ResolvedRole       string
	CanCreateHousehold bool
}

// Resolve builds the current application context for the given user and tenant.
func Resolve(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) (*Resolved, error) {
	// Get tenant
	tenantProc := tenant.NewProcessor(l, ctx, db)
	t, err := tenantProc.ByIDProvider(tenantID)()
	if err != nil {
		return nil, err
	}

	// Get or create preference
	prefProc := preference.NewProcessor(l, ctx, db)
	pref, err := prefProc.FindOrCreate(tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Get memberships
	memProc := membership.NewProcessor(l, ctx, db)
	memberships, err := memProc.ByUserProvider(userID)()
	if err != nil {
		return nil, err
	}

	resolved := &Resolved{
		Tenant:      t,
		Preference:  pref,
		Memberships: memberships,
	}

	// Resolve active household
	if pref.ActiveHouseholdID() != nil {
		hhProc := household.NewProcessor(l, ctx, db)
		hh, err := hhProc.ByIDProvider(*pref.ActiveHouseholdID())()
		if err == nil {
			resolved.ActiveHousehold = &hh
		}
	}

	// If no valid active household, try to find the first one
	if resolved.ActiveHousehold == nil && len(memberships) > 0 {
		hhProc := household.NewProcessor(l, ctx, db)
		hh, err := hhProc.ByIDProvider(memberships[0].HouseholdID())()
		if err == nil {
			resolved.ActiveHousehold = &hh
			// Persist the resolved household
			if _, err := prefProc.SetActiveHousehold(pref.Id(), hh.Id()); err != nil {
				l.WithError(err).Error("Failed to persist resolved active household")
			}
		}
	}

	// Resolve role based on active household membership
	resolved.ResolvedRole = resolveRole(memberships, resolved.ActiveHousehold)
	resolved.CanCreateHousehold = resolved.ResolvedRole == "owner"

	return resolved, nil
}

func resolveRole(memberships []membership.Model, activeHousehold *household.Model) string {
	if activeHousehold == nil {
		return ""
	}
	for _, m := range memberships {
		if m.HouseholdID() == activeHousehold.Id() {
			return m.Role()
		}
	}
	return ""
}
