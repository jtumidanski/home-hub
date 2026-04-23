package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/layout"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Processor orchestrates dashboard CRUD with scope/visibility rules.
type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// ErrInvalidScope is returned when a caller provides a scope outside {household, user}.
var ErrInvalidScope = errors.New("invalid scope")

// ErrNotFound is returned when a dashboard row is missing or invisible to the caller.
var ErrNotFound = errors.New("not found")

// ErrForbidden is returned when a caller attempts to edit a user-scoped dashboard
// owned by someone else.
var ErrForbidden = errors.New("forbidden")

// ErrMixedScope is returned when a Reorder batch mixes household- and user-scoped
// dashboards; the UI renders these in separate sections so reorder runs one
// scope at a time.
var ErrMixedScope = errors.New("reorder requires single scope")

// ReorderPair maps a dashboard id to its desired sort_order in the new order.
type ReorderPair struct {
	ID        uuid.UUID
	SortOrder int
}

// Reorder applies the new sort_order values to the specified dashboards in a
// single transaction. All ids must be visible to the caller and share a single
// scope (all household or all user-owned by caller).
func (p *Processor) Reorder(tenantID, householdID, callerUserID uuid.UUID, pairs []ReorderPair) ([]Model, error) {
	if len(pairs) == 0 {
		return []Model{}, nil
	}
	ids := make([]uuid.UUID, 0, len(pairs))
	for _, pr := range pairs {
		ids = append(ids, pr.ID)
	}
	var rows []Entity
	if err := p.db.WithContext(p.ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) != len(pairs) {
		return nil, ErrNotFound
	}
	var scope string
	for _, r := range rows {
		if r.TenantId != tenantID || r.HouseholdId != householdID {
			return nil, ErrNotFound
		}
		var rowScope string
		if r.UserId == nil {
			rowScope = "household"
		} else {
			if *r.UserId != callerUserID {
				return nil, ErrNotFound
			}
			rowScope = "user"
		}
		if scope == "" {
			scope = rowScope
		} else if scope != rowScope {
			return nil, ErrMixedScope
		}
	}
	upd := map[uuid.UUID]int{}
	for _, pr := range pairs {
		upd[pr.ID] = pr.SortOrder
	}
	if err := updateSortOrders(p.db.WithContext(p.ctx), upd); err != nil {
		return nil, err
	}
	list, err := visibleToCaller(tenantID, householdID, callerUserID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	out := make([]Model, 0, len(list))
	for _, e := range list {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// UpdateAttrs carries optional edits for Update.
type UpdateAttrs struct {
	Name      *string
	Layout    *json.RawMessage
	SortOrder *int
}

// Update applies the given attrs and returns the refreshed row. Household rows
// are editable by any household member; user rows only by the owner.
func (p *Processor) Update(id, tenantID, householdID, callerUserID uuid.UUID, attrs UpdateAttrs) (Model, error) {
	row, err := p.requireEditable(id, tenantID, householdID, callerUserID)
	if err != nil {
		return Model{}, err
	}
	fields := map[string]any{}
	if attrs.Name != nil {
		name := trimName(*attrs.Name)
		if err := validateNameLen(name); err != nil {
			return Model{}, err
		}
		fields["name"] = name
	}
	if attrs.Layout != nil {
		if _, err := layout.Validate(*attrs.Layout); err != nil {
			return Model{}, err
		}
		fields["layout"] = datatypes.JSON(*attrs.Layout)
	}
	if attrs.SortOrder != nil {
		fields["sort_order"] = *attrs.SortOrder
	}
	if len(fields) == 0 {
		return Make(row)
	}
	updated, err := updateFields(p.db.WithContext(p.ctx), id, fields)
	if err != nil {
		return Model{}, err
	}
	return Make(updated)
}

// Delete removes the dashboard if the caller has edit rights under the same
// scope rules as Update.
func (p *Processor) Delete(id, tenantID, householdID, callerUserID uuid.UUID) error {
	if _, err := p.requireEditable(id, tenantID, householdID, callerUserID); err != nil {
		return err
	}
	return deleteByID(p.db.WithContext(p.ctx), id)
}

// requireEditable enforces the PRD §4.10 edit rules: household rows are editable
// by any member of the household; user rows only by the owner. Cross-tenant /
// cross-household rows are hidden as ErrNotFound.
func (p *Processor) requireEditable(id, tenantID, householdID, callerUserID uuid.UUID) (Entity, error) {
	row, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Entity{}, ErrNotFound
	}
	if row.TenantId != tenantID || row.HouseholdId != householdID {
		return Entity{}, ErrNotFound
	}
	if row.UserId != nil && *row.UserId != callerUserID {
		return Entity{}, ErrForbidden
	}
	return row, nil
}

// List returns dashboards visible to the caller (household-scoped + caller's own
// user-scoped rows) ordered by sort_order then created_at.
func (p *Processor) List(tenantID, householdID, callerUserID uuid.UUID) ([]Model, error) {
	rows, err := visibleToCaller(tenantID, householdID, callerUserID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	out := make([]Model, 0, len(rows))
	for _, e := range rows {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// GetByID returns the dashboard if it's visible to the caller. Rows in a
// different tenant/household, or owned by another user, surface as ErrNotFound.
func (p *Processor) GetByID(id, tenantID, householdID, callerUserID uuid.UUID) (Model, error) {
	row, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if row.TenantId != tenantID || row.HouseholdId != householdID {
		return Model{}, ErrNotFound
	}
	if row.UserId != nil && *row.UserId != callerUserID {
		return Model{}, ErrNotFound
	}
	return Make(row)
}

// CreateAttrs carries Create inputs for a dashboard.
type CreateAttrs struct {
	Name      string
	Scope     string // "household" | "user"
	Layout    json.RawMessage
	SortOrder *int
}

// Create inserts a new dashboard under the requested scope after validating
// name and layout.
func (p *Processor) Create(tenantID, householdID, callerUserID uuid.UUID, attrs CreateAttrs) (Model, error) {
	var userID *uuid.UUID
	switch attrs.Scope {
	case "household":
		userID = nil
	case "user":
		u := callerUserID
		userID = &u
	default:
		return Model{}, ErrInvalidScope
	}

	layoutBytes := attrs.Layout
	if len(layoutBytes) == 0 {
		layoutBytes = json.RawMessage(`{"version":1,"widgets":[]}`)
	}
	if _, err := layout.Validate(layoutBytes); err != nil {
		return Model{}, err
	}

	sortOrder := 0
	if attrs.SortOrder != nil {
		sortOrder = *attrs.SortOrder
	} else {
		max, err := maxSortOrderInScope(p.db.WithContext(p.ctx), tenantID, householdID, userID)
		if err != nil {
			return Model{}, err
		}
		sortOrder = max + 1
	}

	name := trimName(attrs.Name)
	if err := validateNameLen(name); err != nil {
		return Model{}, err
	}

	e := Entity{
		TenantId:      tenantID,
		HouseholdId:   householdID,
		UserId:        userID,
		Name:          name,
		SortOrder:     sortOrder,
		Layout:        datatypes.JSON(layoutBytes),
		SchemaVersion: 1,
	}
	saved, err := insert(p.db.WithContext(p.ctx), e)
	if err != nil {
		return Model{}, err
	}
	return Make(saved)
}

func trimName(s string) string {
	return strings.TrimSpace(s)
}

func validateNameLen(s string) error {
	if s == "" {
		return ErrNameRequired
	}
	if len(s) > 80 {
		return ErrNameTooLong
	}
	return nil
}
