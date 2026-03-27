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

func (p *Processor) Create(tenantID uuid.UUID, name, displayName, unitFamily string) (Model, error) {
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
		CreatedAt: now, UpdatedAt: now,
	}
	if err := p.db.WithContext(p.ctx).Create(&e).Error; err != nil {
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

func (p *Processor) Update(id uuid.UUID, name, displayName, unitFamily *string) (Model, error) {
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

	e.UpdatedAt = time.Now().UTC()
	if err := p.db.WithContext(p.ctx).Save(&e).Error; err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	count, err := getUsageCount(p.db.WithContext(p.ctx), id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrHasReferences
	}

	return p.db.WithContext(p.ctx).Where("id = ?", id).Delete(&Entity{}).Error
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
	if err := p.db.WithContext(p.ctx).Create(&alias).Error; err != nil {
		return AliasEntity{}, err
	}
	return alias, nil
}

func (p *Processor) RemoveAlias(aliasID uuid.UUID) error {
	return p.db.WithContext(p.ctx).Where("id = ?", aliasID).Delete(&AliasEntity{}).Error
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

func (p *Processor) SearchWithUsage(tenantID uuid.UUID, query string, page, pageSize int) ([]Model, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	entities, total, err := searchWithUsage(tenantID, query, page, pageSize)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, 0, err
	}

	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e.Entity)
		if err != nil {
			return nil, 0, err
		}
		models[i] = Model{
			id: m.id, tenantID: m.tenantID, name: m.name, displayName: m.displayName,
			unitFamily: m.unitFamily, aliases: m.aliases, aliasCount: m.aliasCount,
			usageCount: int(e.UsageCount), createdAt: m.createdAt, updatedAt: m.updatedAt,
		}
	}
	return models, total, nil
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
