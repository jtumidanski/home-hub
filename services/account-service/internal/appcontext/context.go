package appcontext

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/invitation"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/services/account-service/internal/tenant"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Resolved holds the fully resolved application context for a user.
type Resolved struct {
	Tenant                 tenant.Model
	ActiveHousehold        *household.Model
	Preference             preference.Model
	Memberships            []membership.Model
	ResolvedRole           string
	CanCreateHousehold     bool
	PendingInvitationCount int64
}

// Resolve builds the current application context for the given user and tenant.
// If tenantID is nil, the tenant is resolved from the user's first membership.
// Context resolution bypasses tenant filtering since it is the bootstrapping query.
// userEmail is used to query pending invitation count (may be empty if unavailable).
func Resolve(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID, userEmail string) (*Resolved, error) {
	// Bypass tenant filtering — this is the bootstrapping query that establishes
	// which tenant the user belongs to.
	ctx = database.WithoutTenantFilter(ctx)

	if tenantID == uuid.Nil {
		memProc := membership.NewProcessor(l, ctx, db)
		memberships, err := memProc.ByUserProvider(userID)()
		if err != nil {
			return nil, err
		}
		if len(memberships) == 0 {
			return nil, fmt.Errorf("no memberships found for user %s", userID)
		}
		tenantID = memberships[0].TenantID()
	}

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
				return nil, err
			}
		}
	}

	// Resolve role based on active household membership
	resolved.ResolvedRole = resolveRole(memberships, resolved.ActiveHousehold)
	resolved.CanCreateHousehold = resolved.ResolvedRole == "owner"

	// Count pending invitations for this user's email (bypasses tenant filter)
	if userEmail != "" {
		invProc := invitation.NewProcessor(l, ctx, db)
		count, err := invProc.CountByEmailPending(userEmail)
		if err != nil {
			l.WithError(err).Warn("Failed to count pending invitations")
		} else {
			resolved.PendingInvitationCount = count
		}
	}

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
