package ingredient

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("canonical ingredient not found")
	ErrAliasConflict  = errors.New("alias conflicts with existing canonical ingredient name or alias")
	ErrHasReferences  = errors.New("canonical ingredient is still referenced by recipe ingredients")
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
	return model.Map(Make)(GetByID(id)(p.db.WithContext(p.ctx)))
}

func (p *Processor) Create(tenantID uuid.UUID, name, displayName, unitFamily string, categoryID *uuid.UUID) (Model, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if _, err := NewBuilder().SetName(normalized).SetUnitFamily(unitFamily).Build(); err != nil {
		return Model{}, err
	}

	now := time.Now().UTC()
	var dn *string
	if displayName != "" {
		dn = &displayName
	}
	var uf *string
	if unitFamily != "" {
		uf = &unitFamily
	}

	e := Entity{
		Id: uuid.New(), TenantId: tenantID,
		Name: normalized, DisplayName: dn, UnitFamily: uf,
		CategoryId: categoryID,
		CreatedAt:  now, UpdatedAt: now,
	}
	if err := createEntity(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	e.Aliases = []AliasEntity{}
	return Make(e)
}

func (p *Processor) Get(id uuid.UUID) (Model, error) {
	m, err := p.ByIDProvider(id)()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return m, nil
}

// GetByIDs fetches canonical ingredients (with aliases) for many ids and
// returns them keyed by ingredient id. Empty input returns an empty map
// without hitting the database.
func (p *Processor) GetByIDs(ids []uuid.UUID) (map[uuid.UUID]Model, error) {
	result := make(map[uuid.UUID]Model)
	if len(ids) == 0 {
		return result, nil
	}
	entities, err := GetByIDs(ids)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		result[e.Id] = m
	}
	return result, nil
}

func (p *Processor) Search(tenantID uuid.UUID, query string, page, pageSize int) ([]Model, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	entities, total, err := search(tenantID, query, page, pageSize)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, 0, err
	}

	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, 0, err
		}
		models[i] = m
	}
	return models, total, nil
}

type UpdateCategoryOpt struct {
	Set   bool       // true means the caller wants to change the category
	Value *uuid.UUID // nil = clear category, non-nil = assign category
}

func (p *Processor) Update(id uuid.UUID, name, displayName, unitFamily *string, categoryOpt *UpdateCategoryOpt) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}

	if name != nil {
		e.Name = strings.ToLower(strings.TrimSpace(*name))
	}
	if displayName != nil {
		e.DisplayName = displayName
	}
	if unitFamily != nil {
		if *unitFamily == "" {
			e.UnitFamily = nil
		} else {
			if !ValidUnitFamily(*unitFamily) {
				return Model{}, ErrInvalidUnitFamily
			}
			e.UnitFamily = unitFamily
		}
	}
	if categoryOpt != nil && categoryOpt.Set {
		e.CategoryId = categoryOpt.Value
	}

	if err := saveEntity(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	// Nullify references in recipe_ingredients, setting them back to unresolved
	if err := nullifyReferences(p.db.WithContext(p.ctx), id); err != nil {
		p.l.WithError(err).Error("Failed to nullify ingredient references")
	}
	return deleteEntity(p.db.WithContext(p.ctx), id)
}

func (p *Processor) AddAlias(tenantID, ingredientID uuid.UUID, aliasName string) (AliasEntity, error) {
	normalized := strings.ToLower(strings.TrimSpace(aliasName))

	// Check if alias conflicts with an existing canonical name
	if _, err := GetByName(tenantID, normalized)(p.db.WithContext(p.ctx))(); err == nil {
		return AliasEntity{}, ErrAliasConflict
	}
	// Check if alias already exists
	if _, _, err := GetByAlias(tenantID, normalized)(p.db.WithContext(p.ctx)); err == nil {
		return AliasEntity{}, ErrAliasConflict
	}

	alias := AliasEntity{
		Id:                    uuid.New(),
		TenantId:              tenantID,
		CanonicalIngredientId: ingredientID,
		Name:                  normalized,
		CreatedAt:             time.Now().UTC(),
	}
	if err := createAlias(p.db.WithContext(p.ctx), &alias); err != nil {
		return AliasEntity{}, err
	}
	return alias, nil
}

func (p *Processor) RemoveAlias(aliasID uuid.UUID) error {
	return deleteAlias(p.db.WithContext(p.ctx), aliasID)
}

func (p *Processor) GetUsageCount(id uuid.UUID) (int64, error) {
	return getUsageCount(p.db.WithContext(p.ctx), id)
}

func (p *Processor) GetIngredientRecipes(id uuid.UUID, page, pageSize int) ([]RecipeRef, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return getIngredientRecipes(p.db.WithContext(p.ctx), id, page, pageSize)
}

func (p *Processor) Reassign(fromID, toID uuid.UUID) (int64, error) {
	count, err := reassignCanonical(p.db.WithContext(p.ctx), fromID, toID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (p *Processor) SearchWithUsage(tenantID uuid.UUID, query string, categoryFilter string, page, pageSize int) ([]Model, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	entities, total, err := searchWithUsage(tenantID, query, categoryFilter, page, pageSize)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, 0, err
	}

	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e.Entity)
		if err != nil {
			return nil, 0, err
		}
		models[i] = m.WithUsageCount(int(e.UsageCount))
	}
	return models, total, nil
}

func (p *Processor) BulkCategorize(tenantID uuid.UUID, ingredientIDs []uuid.UUID, categoryID uuid.UUID) error {
	return bulkUpdateCategory(p.db.WithContext(p.ctx), ingredientIDs, tenantID, categoryID)
}

// LookupByName resolves a free-form ingredient name to a canonical ingredient
// for the given tenant. It tries (in order): exact name match, alias match,
// then a normalized variant (strips leading articles and a trailing plural
// 's') against both names and aliases. Returns (model, true, nil) on a match,
// (Model{}, false, nil) on a clean miss, and (Model{}, false, err) for any
// database error other than "not found".
func (p *Processor) LookupByName(tenantID uuid.UUID, name string) (Model, bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return Model{}, false, nil
	}

	candidates := []string{normalized}
	if alt := lookupNormalizeText(normalized); alt != normalized && alt != "" {
		candidates = append(candidates, alt)
	}

	for _, candidate := range candidates {
		if e, err := GetByName(tenantID, candidate)(p.db.WithContext(p.ctx))(); err == nil {
			m, err := Make(e)
			if err != nil {
				return Model{}, false, err
			}
			return m, true, nil
		}
		if e, _, err := GetByAlias(tenantID, candidate)(p.db.WithContext(p.ctx)); err == nil {
			m, err := Make(*e)
			if err != nil {
				return Model{}, false, err
			}
			return m, true, nil
		}
	}
	return Model{}, false, nil
}

// lookupNormalizeText strips leading articles and a trailing plural 's' so
// that "the eggs" matches a canonical entry stored as "egg". Mirrors the
// normalization package's helper but kept private here to avoid an import
// cycle on a four-line function.
func lookupNormalizeText(s string) string {
	for _, article := range []string{"the ", "a ", "an "} {
		if strings.HasPrefix(s, article) {
			s = s[len(article):]
			break
		}
	}
	fields := strings.Fields(s)
	s = strings.Join(fields, " ")
	if len(s) > 1 && strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		s = s[:len(s)-1]
	}
	return strings.TrimSpace(s)
}

func (p *Processor) CountByCategory(tenantID uuid.UUID) (map[uuid.UUID]int, error) {
	return countByCategory(tenantID)(p.db.WithContext(p.ctx))
}

func (p *Processor) Suggest(tenantID uuid.UUID, prefix string, limit int) ([]Model, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	entities, err := suggestByPrefix(tenantID, prefix, limit)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, err
	}
	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		models[i] = m
	}
	return models, nil
}
