package membership

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotAuthorized    = errors.New("user does not have owner or admin role")
	ErrCannotModifySelf = errors.New("user cannot modify their own role")
	ErrCannotModifyOwner = errors.New("admin cannot modify an owner's role")
	ErrCannotRemoveOwner = errors.New("admin cannot remove an owner")
	ErrLastOwner        = errors.New("cannot leave: you are the last owner of this household")
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

func (p *Processor) ByUserProvider(userID uuid.UUID) model.Provider[[]Model] {
	return model.SliceMap(Make)(getByUser(userID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByHouseholdAndUserProvider(householdID, userID uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByHouseholdAndUser(householdID, userID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) ByHouseholdProvider(householdID uuid.UUID) model.Provider[[]Model] {
	return model.SliceMap(Make)(getByHousehold(householdID)(p.db.WithContext(p.ctx)))
}

func (p *Processor) CountOwnersByHousehold(householdID uuid.UUID) (int64, error) {
	return countOwnersByHousehold(p.db.WithContext(p.ctx), householdID)
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, role string) (Model, error) {
	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, userID, role)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) UpdateRole(id uuid.UUID, role string) (Model, error) {
	e, err := updateRole(p.db.WithContext(p.ctx), id, role)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

// UpdateRoleAuthorized updates a membership's role with authorization checks:
// - Requester must be owner or admin in the household
// - Admin cannot modify an owner's role
// - User cannot modify their own role
func (p *Processor) UpdateRoleAuthorized(id uuid.UUID, role string, requesterID uuid.UUID) (Model, error) {
	target, err := p.ByIDProvider(id)()
	if err != nil {
		return Model{}, err
	}

	requester, err := p.ByHouseholdAndUserProvider(target.HouseholdID(), requesterID)()
	if err != nil {
		return Model{}, ErrNotAuthorized
	}
	if requester.Role() != "owner" && requester.Role() != "admin" {
		return Model{}, ErrNotAuthorized
	}
	if target.UserID() == requesterID {
		return Model{}, ErrCannotModifySelf
	}
	if requester.Role() == "admin" && target.Role() == "owner" {
		return Model{}, ErrCannotModifyOwner
	}

	return p.UpdateRole(id, role)
}

// DeleteAuthorized deletes a membership with authorization checks:
// - Self-deletion (leave): allowed for any member, blocked if last owner
// - Admin/owner deletion: allowed, admin cannot remove owner
func (p *Processor) DeleteAuthorized(id uuid.UUID, requesterID uuid.UUID) error {
	target, err := p.ByIDProvider(id)()
	if err != nil {
		return err
	}

	isSelf := target.UserID() == requesterID

	if isSelf {
		// Last-owner guard
		if target.Role() == "owner" {
			count, err := p.CountOwnersByHousehold(target.HouseholdID())
			if err != nil {
				return err
			}
			if count <= 1 {
				return ErrLastOwner
			}
		}
	} else {
		// Not self — requires owner/admin
		requester, err := p.ByHouseholdAndUserProvider(target.HouseholdID(), requesterID)()
		if err != nil {
			return ErrNotAuthorized
		}
		if requester.Role() != "owner" && requester.Role() != "admin" {
			return ErrNotAuthorized
		}
		if requester.Role() == "admin" && target.Role() == "owner" {
			return ErrCannotRemoveOwner
		}
	}

	return deleteByID(p.db.WithContext(p.ctx), id)
}

// DeleteAuthorizedWithCleanup performs authorized deletion and clears the user's active
// household preference if they are leaving their currently active household.
func (p *Processor) DeleteAuthorizedWithCleanup(id uuid.UUID, requesterID uuid.UUID) error {
	target, lookupErr := p.ByIDProvider(id)()

	if err := p.DeleteAuthorized(id, requesterID); err != nil {
		return err
	}

	if lookupErr == nil && target.UserID() == requesterID {
		prefProc := preference.NewProcessor(p.l, p.ctx, p.db)
		pref, err := prefProc.ByUserProvider(requesterID)()
		if err == nil && pref.ActiveHouseholdID() != nil && *pref.ActiveHouseholdID() == target.HouseholdID() {
			if _, err := prefProc.ClearActiveHousehold(pref.Id()); err != nil {
				p.l.WithError(err).Warn("Failed to clear active household after leave")
			}
		}
	}

	return nil
}

func (p *Processor) Delete(id uuid.UUID) error {
	return deleteByID(p.db.WithContext(p.ctx), id)
}
