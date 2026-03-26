package invitation

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotAuthorized       = errors.New("user does not have owner or admin role")
	ErrAlreadyInvited      = errors.New("a pending invitation already exists for this email and household")
	ErrAlreadyMember       = errors.New("email already has a membership in this household")
	ErrNotPending          = errors.New("invitation is not in pending status")
	ErrExpired             = errors.New("invitation has expired")
	ErrEmailMismatch       = errors.New("user email does not match invitation email")
	ErrCrossTenant         = errors.New("user belongs to a different tenant than the invitation")
	ErrAlreadyHasMembership = errors.New("user already has a membership in this household")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByHouseholdPendingProvider(householdID uuid.UUID) model.Provider[[]Model] {
	return model.SliceMap(Make)(getByHouseholdPending(householdID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByEmailPendingProvider(email string) model.Provider[[]Model] {
	ctx := database.WithoutTenantFilter(p.ctx)
	return model.SliceMap(Make)(getByEmailPending(email)(p.db.WithContext(ctx)))
}

func (p *Processor) CountByEmailPending(email string) (int64, error) {
	ctx := database.WithoutTenantFilter(p.ctx)
	return countByEmailPending(p.db.WithContext(ctx), email)
}

// Create creates a new invitation after validating authorization and uniqueness.
func (p *Processor) Create(tenantID, householdID uuid.UUID, email, role string, inviterID uuid.UUID) (Model, error) {
	if err := p.requirePrivilegedRole(householdID, inviterID); err != nil {
		return Model{}, err
	}

	// Default role to viewer
	if role == "" {
		role = "viewer"
	}

	// Check no existing pending invitation for this email+household
	_, err := getByHouseholdAndEmailPending(householdID, email)(p.db.WithContext(p.ctx))()
	if err == nil {
		return Model{}, ErrAlreadyInvited
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Model{}, err
	}

	// Check no existing membership for this email in this household.
	// We look up memberships by household and check email matches.
	// Since memberships don't store email, we can't check this directly here.
	// The uniqueness will be enforced when the invitation is accepted.

	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, strings.ToLower(email), role, inviterID)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

// Revoke sets a pending invitation to revoked status.
func (p *Processor) Revoke(id uuid.UUID, revokerID uuid.UUID) error {
	inv, err := p.ByIDProvider(id)()
	if err != nil {
		return err
	}
	if inv.Status() != "pending" {
		return ErrNotPending
	}
	if err := p.requirePrivilegedRole(inv.HouseholdID(), revokerID); err != nil {
		return err
	}
	return updateStatus(p.db.WithContext(p.ctx), id, "revoked")
}

// Accept accepts a pending invitation: creates membership, handles tenant assignment, creates preference.
func (p *Processor) Accept(id, userID uuid.UUID, userEmail string, userTenantID uuid.UUID) (Model, error) {
	ctx := database.WithoutTenantFilter(p.ctx)
	db := p.db.WithContext(ctx)

	inv, err := model.Map(Make)(getByID(id)(db))()
	if err != nil {
		return Model{}, err
	}
	if inv.Status() != "pending" {
		return Model{}, ErrNotPending
	}
	if inv.ExpiresAt().Before(timeNow()) {
		return Model{}, ErrExpired
	}
	if !strings.EqualFold(userEmail, inv.Email()) {
		return Model{}, ErrEmailMismatch
	}

	// Cross-tenant check: if user already has a tenant, it must match the invitation's tenant
	if userTenantID != uuid.Nil && userTenantID != inv.TenantID() {
		return Model{}, ErrCrossTenant
	}

	// Check no existing membership for this user in this household
	memProc := membership.NewProcessor(p.l, ctx, p.db)
	_, memErr := memProc.ByHouseholdAndUserProvider(inv.HouseholdID(), userID)()
	if memErr == nil {
		return Model{}, ErrAlreadyHasMembership
	}
	if !errors.Is(memErr, gorm.ErrRecordNotFound) {
		return Model{}, memErr
	}

	// Create membership
	_, err = memProc.Create(inv.TenantID(), inv.HouseholdID(), userID, inv.Role())
	if err != nil {
		return Model{}, err
	}

	// Create or find preference, then set active household
	prefProc := preference.NewProcessor(p.l, ctx, p.db)
	pref, err := prefProc.FindOrCreate(inv.TenantID(), userID)
	if err != nil {
		return Model{}, err
	}
	if _, err := prefProc.SetActiveHousehold(pref.Id(), inv.HouseholdID()); err != nil {
		return Model{}, err
	}

	// Update invitation status
	if err := updateStatus(db, id, "accepted"); err != nil {
		return Model{}, err
	}
	return model.Map(Make)(getByID(id)(db))()
}

// Decline sets a pending invitation to declined status.
func (p *Processor) Decline(id uuid.UUID, userEmail string) (Model, error) {
	ctx := database.WithoutTenantFilter(p.ctx)
	db := p.db.WithContext(ctx)

	inv, err := model.Map(Make)(getByID(id)(db))()
	if err != nil {
		return Model{}, err
	}
	if inv.Status() != "pending" {
		return Model{}, ErrNotPending
	}
	if !strings.EqualFold(userEmail, inv.Email()) {
		return Model{}, ErrEmailMismatch
	}

	if err := updateStatus(db, id, "declined"); err != nil {
		return Model{}, err
	}
	return model.Map(Make)(getByID(id)(db))()
}

// ByEmailPendingWithHouseholds returns pending invitations for an email along with their associated households.
func (p *Processor) ByEmailPendingWithHouseholds(email string) ([]Model, []household.Model, error) {
	ctx := database.WithoutTenantFilter(p.ctx)
	db := p.db.WithContext(ctx)

	invitations, err := model.SliceMap(Make)(getByEmailPending(email)(db))()
	if err != nil {
		return nil, nil, err
	}

	hhProc := household.NewProcessor(p.l, ctx, p.db)
	households := make([]household.Model, 0, len(invitations))
	for _, inv := range invitations {
		hh, err := hhProc.ByIDProvider(inv.HouseholdID())()
		if err == nil {
			households = append(households, hh)
		}
	}
	return invitations, households, nil
}

// requirePrivilegedRole checks that the given user has owner or admin role in the household.
func (p *Processor) requirePrivilegedRole(householdID, userID uuid.UUID) error {
	memProc := membership.NewProcessor(p.l, p.ctx, p.db)
	mem, err := memProc.ByHouseholdAndUserProvider(householdID, userID)()
	if err != nil {
		return ErrNotAuthorized
	}
	if mem.Role() != "owner" && mem.Role() != "admin" {
		return ErrNotAuthorized
	}
	return nil
}

// timeNow is a variable for testing.
var timeNow = func() time.Time { return time.Now().UTC() }
