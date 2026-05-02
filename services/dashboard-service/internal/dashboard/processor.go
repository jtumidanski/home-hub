package dashboard

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
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

// ErrAlreadyHousehold is returned by Promote when the target row is already
// household-scoped.
var ErrAlreadyHousehold = errors.New("dashboard already household-scoped")

// ErrNotCopyable is returned by CopyToMine when asked to copy a user-scoped
// dashboard. Only household dashboards are copyable.
var ErrNotCopyable = errors.New("only household dashboards can be copied to mine")

// CopyToMine deep-copies a household dashboard into the caller's user scope,
// regenerating widget ids so the two rows are independent.
func (p *Processor) CopyToMine(id, tenantID, householdID, callerUserID uuid.UUID) (Model, error) {
	row, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil || row.TenantId != tenantID || row.HouseholdId != householdID {
		return Model{}, ErrNotFound
	}
	if row.UserId != nil {
		return Model{}, ErrNotCopyable
	}

	regenerated, err := regenerateWidgetIDs([]byte(row.Layout))
	if err != nil {
		return Model{}, err
	}

	ownerID := callerUserID
	max, err := maxSortOrderInScope(p.db.WithContext(p.ctx), tenantID, householdID, &ownerID)
	if err != nil {
		return Model{}, err
	}

	copyName := row.Name + " (mine)"
	if len(copyName) > 80 {
		copyName = copyName[:80]
	}

	newRow := Entity{
		TenantId:      tenantID,
		HouseholdId:   householdID,
		UserId:        &ownerID,
		Name:          copyName,
		SortOrder:     max + 1,
		Layout:        datatypes.JSON(regenerated),
		SchemaVersion: row.SchemaVersion,
	}
	saved, err := insert(p.db.WithContext(p.ctx), newRow)
	if err != nil {
		return Model{}, err
	}
	return Make(saved)
}

// SeedResult reports whether Seed created a new row. When Created is false,
// Existing holds the dashboards visible to the caller.
type SeedResult struct {
	Created   bool
	Dashboard Model
	Existing  []Model
}

// Seed ensures a household-scoped dashboard exists for the given
// (tenant, household[, seedKey]). It is idempotent and race-safe.
//
// When seedKey is non-nil, the lookup is keyed on (tenant, household, seedKey)
// — distinct keys (e.g. "home" and "kiosk") coexist as separate rows. When
// seedKey is nil, the legacy behavior is preserved: if any household-scoped
// row already exists, no new row is created and the visible set is returned.
//
// Concurrent Seed calls across replicas are serialized via a Postgres advisory
// xact lock keyed on tenant+household[+seedKey]. On non-Postgres dialects
// (sqlite in tests), the partial unique index `idx_dashboards_seed_key` is the
// race backstop for keyed inserts.
func (p *Processor) Seed(tenantID, householdID, callerUserID uuid.UUID, name string, seedKey *string, layoutJSON json.RawMessage) (SeedResult, error) {
	name = trimName(name)
	if err := validateNameLen(name); err != nil {
		return SeedResult{}, err
	}
	if _, err := layout.Validate(layoutJSON); err != nil {
		return SeedResult{}, err
	}

	var out SeedResult
	err := p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		if err := p.acquireSeedLock(tx, tenantID, householdID, seedKey); err != nil {
			return err
		}

		if seedKey != nil {
			// Key path: look up the (tenant, household, seedKey) row; insert if
			// missing, otherwise return the existing row.
			var existing Entity
			err := tx.Where("tenant_id = ? AND household_id = ? AND seed_key = ?",
				tenantID, householdID, *seedKey).First(&existing).Error
			if err == nil {
				m, mkErr := Make(existing)
				if mkErr != nil {
					return mkErr
				}
				out.Existing = append(out.Existing, m)
				out.Created = false
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			key := *seedKey
			e := Entity{
				TenantId:      tenantID,
				HouseholdId:   householdID,
				UserId:        nil,
				Name:          name,
				SortOrder:     0,
				Layout:        datatypes.JSON(layoutJSON),
				SchemaVersion: 1,
				SeedKey:       &key,
			}
			saved, err := insert(tx, e)
			if err != nil {
				return err
			}
			m, err := Make(saved)
			if err != nil {
				return err
			}
			out.Dashboard = m
			out.Created = true
			return nil
		}

		// Legacy path: any household-scoped row blocks creation.
		count, err := countHouseholdScoped(tx, tenantID, householdID)
		if err != nil {
			return err
		}
		if count > 0 {
			list, err := visibleToCaller(tenantID, householdID, callerUserID)(tx)()
			if err != nil {
				return err
			}
			for _, e := range list {
				m, err := Make(e)
				if err != nil {
					return err
				}
				out.Existing = append(out.Existing, m)
			}
			out.Created = false
			return nil
		}
		e := Entity{
			TenantId:      tenantID,
			HouseholdId:   householdID,
			UserId:        nil,
			Name:          name,
			SortOrder:     0,
			Layout:        datatypes.JSON(layoutJSON),
			SchemaVersion: 1,
		}
		saved, err := insert(tx, e)
		if err != nil {
			return err
		}
		m, err := Make(saved)
		if err != nil {
			return err
		}
		out.Dashboard = m
		out.Created = true
		return nil
	})
	return out, err
}

// acquireSeedLock takes a Postgres transaction-scoped advisory lock so
// concurrent seeders serialize on the same (tenant, household[, seedKey]) key.
// On other dialects (sqlite in tests) this is a no-op; production always runs
// Postgres.
func (p *Processor) acquireSeedLock(tx *gorm.DB, tenantID, householdID uuid.UUID, seedKey *string) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	var key int64
	if seedKey != nil {
		key = seedLockKeyForKey(tenantID, householdID, *seedKey)
	} else {
		key = seedLockKey(tenantID, householdID)
	}
	return tx.Exec("SELECT pg_advisory_xact_lock(?)", key).Error
}

// seedLockKey derives a deterministic int64 lock key from tenant+household so
// every replica hashes to the same advisory-lock slot.
func seedLockKey(tenantID, householdID uuid.UUID) int64 {
	var combined [32]byte
	copy(combined[:16], tenantID[:])
	copy(combined[16:], householdID[:])
	sum := sha256.Sum256(combined[:])
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

// seedLockKeyForKey derives a deterministic int64 lock key from
// tenant+household+seedKey so distinct keys do not block each other while
// concurrent inserts of the same key still serialize.
func seedLockKeyForKey(tenantID, householdID uuid.UUID, seedKey string) int64 {
	combined := make([]byte, 0, 32+len(seedKey))
	combined = append(combined, tenantID[:]...)
	combined = append(combined, householdID[:]...)
	combined = append(combined, []byte(seedKey)...)
	sum := sha256.Sum256(combined)
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

// regenerateWidgetIDs assigns a fresh UUID to every widget in the layout
// document, preserving all other fields.
func regenerateWidgetIDs(raw []byte) ([]byte, error) {
	var doc struct {
		Version int              `json:"version"`
		Widgets []map[string]any `json:"widgets"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	for i := range doc.Widgets {
		doc.Widgets[i]["id"] = uuid.New().String()
	}
	return json.Marshal(doc)
}

// Promote turns a user-scoped dashboard into a household-scoped one by clearing
// its user_id. Only the owner may promote.
func (p *Processor) Promote(id, tenantID, householdID, callerUserID uuid.UUID) (Model, error) {
	row, err := p.requireEditable(id, tenantID, householdID, callerUserID)
	if err != nil {
		return Model{}, err
	}
	if row.UserId == nil {
		return Model{}, ErrAlreadyHousehold
	}
	if err := clearUserID(p.db.WithContext(p.ctx), id); err != nil {
		return Model{}, err
	}
	updated, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, err
	}
	return Make(updated)
}

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
