package normalization

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/ingredient"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound           = errors.New("recipe ingredient not found")
	ErrAliasConflict      = errors.New("alias conflicts with existing canonical ingredient name")
	ErrAliasAlreadyExists = errors.New("alias already exists")
)

type ParsedIngredient struct {
	Name     string
	Quantity string
	Unit     string
}

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// tryMatch attempts exact canonical match then alias match. Returns canonical ID and status.
func (p *Processor) tryMatch(tenantID uuid.UUID, name string) (*uuid.UUID, Status) {
	if e, err := ingredient.GetByName(tenantID, name)(p.db.WithContext(p.ctx))(); err == nil {
		id := e.Id
		return &id, StatusMatched
	}
	if ce, _, err := ingredient.GetByAlias(tenantID, name)(p.db.WithContext(p.ctx)); err == nil {
		id := ce.Id
		return &id, StatusAliasMatched
	}
	return nil, StatusUnresolved
}

// matchWithNormalization tries exact+alias, then text-normalized exact+alias.
func (p *Processor) matchWithNormalization(tenantID uuid.UUID, rawName string) (*uuid.UUID, Status) {
	canonicalID, status := p.tryMatch(tenantID, rawName)
	if status != StatusUnresolved {
		return canonicalID, status
	}
	normalized := normalizeText(rawName)
	if normalized != rawName {
		return p.tryMatch(tenantID, normalized)
	}
	return nil, StatusUnresolved
}

func resolveUnit(rawUnit string) *string {
	if rawUnit == "" {
		return nil
	}
	if identity, ok := LookupUnit(strings.ToLower(strings.TrimSpace(rawUnit))); ok {
		return &identity.Canonical
	}
	return nil
}

func (p *Processor) NormalizeIngredients(tenantID, householdID, recipeID uuid.UUID, parsed []ParsedIngredient) ([]Model, error) {
	now := time.Now().UTC()
	entities := make([]Entity, len(parsed))

	for i, pi := range parsed {
		rawName := strings.ToLower(strings.TrimSpace(pi.Name))
		canonicalID, status := p.matchWithNormalization(tenantID, rawName)

		var rawQty *string
		if pi.Quantity != "" {
			rawQty = &pi.Quantity
		}
		var rawUnit *string
		if pi.Unit != "" {
			rawUnit = &pi.Unit
		}

		entities[i] = Entity{
			Id:                    uuid.New(),
			TenantId:              tenantID,
			HouseholdId:           householdID,
			RecipeId:              recipeID,
			RawName:               rawName,
			RawQuantity:           rawQty,
			RawUnit:               rawUnit,
			Position:              i,
			CanonicalIngredientId: canonicalID,
			CanonicalUnit:         resolveUnit(pi.Unit),
			NormalizationStatus:   string(status),
			CreatedAt:             now,
			UpdatedAt:             now,
		}
	}

	if err := bulkCreate(p.db.WithContext(p.ctx), entities); err != nil {
		return nil, err
	}

	models, err := entitiesToModels(entities)
	if err != nil {
		return nil, err
	}

	matched := 0
	for _, m := range models {
		if m.NormalizationStatus() != StatusUnresolved {
			matched++
		}
	}
	p.l.WithFields(logrus.Fields{
		"recipe_id":  recipeID,
		"total":      len(models),
		"matched":    matched,
		"unresolved": len(models) - matched,
	}).Info("Normalization pipeline completed")

	return models, nil
}

func (p *Processor) ReconcileIngredients(tenantID, householdID, recipeID uuid.UUID, newParsed []ParsedIngredient) ([]Model, error) {
	existing, err := getByRecipeID(recipeID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}

	existingByName := make(map[string]Entity)
	for _, e := range existing {
		existingByName[strings.ToLower(strings.TrimSpace(e.RawName))] = e
	}

	now := time.Now().UTC()
	var newEntities []Entity

	for i, pi := range newParsed {
		rawName := strings.ToLower(strings.TrimSpace(pi.Name))
		var rawQty *string
		if pi.Quantity != "" {
			rawQty = &pi.Quantity
		}
		var rawUnit *string
		if pi.Unit != "" {
			rawUnit = &pi.Unit
		}

		if prev, ok := existingByName[rawName]; ok {
			if Status(prev.NormalizationStatus) == StatusManuallyConfirmed {
				prev.Position = i
				prev.RawQuantity = rawQty
				prev.RawUnit = rawUnit
				prev.CanonicalUnit = resolveUnit(pi.Unit)
				prev.UpdatedAt = now
				newEntities = append(newEntities, prev)
			} else {
				canonicalID, status := p.matchWithNormalization(tenantID, rawName)
				prev.Position = i
				prev.RawQuantity = rawQty
				prev.RawUnit = rawUnit
				prev.CanonicalIngredientId = canonicalID
				prev.CanonicalUnit = resolveUnit(pi.Unit)
				prev.NormalizationStatus = string(status)
				prev.UpdatedAt = now
				newEntities = append(newEntities, prev)
			}
			delete(existingByName, rawName)
		} else {
			canonicalID, status := p.matchWithNormalization(tenantID, rawName)
			newEntities = append(newEntities, Entity{
				Id:                    uuid.New(),
				TenantId:              tenantID,
				HouseholdId:           householdID,
				RecipeId:              recipeID,
				RawName:               rawName,
				RawQuantity:           rawQty,
				RawUnit:               rawUnit,
				Position:              i,
				CanonicalIngredientId: canonicalID,
				CanonicalUnit:         resolveUnit(pi.Unit),
				NormalizationStatus:   string(status),
				CreatedAt:             now,
				UpdatedAt:             now,
			})
		}
	}

	if err := deleteByRecipeID(p.db.WithContext(p.ctx), recipeID); err != nil {
		return nil, err
	}
	if err := bulkCreate(p.db.WithContext(p.ctx), newEntities); err != nil {
		return nil, err
	}

	return entitiesToModels(newEntities)
}

type ResolveResult struct {
	Model       Model
	AliasCreated bool
}

func (p *Processor) ResolveIngredient(ingredientID, canonicalIngredientID uuid.UUID, saveAsAlias bool) (ResolveResult, error) {
	e, err := getByID(ingredientID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return ResolveResult{}, ErrNotFound
	}

	e.CanonicalIngredientId = &canonicalIngredientID
	e.NormalizationStatus = string(StatusManuallyConfirmed)

	if err := updateOne(p.db.WithContext(p.ctx), &e); err != nil {
		return ResolveResult{}, err
	}

	aliasCreated := false
	if saveAsAlias {
		if err := p.createAliasIfNotExists(e.TenantId, canonicalIngredientID, e.RawName); err != nil {
			p.l.WithError(err).Warn("Failed to create alias from resolution")
		} else {
			aliasCreated = true
		}
	}

	m, err := Make(e)
	if err != nil {
		return ResolveResult{}, err
	}

	// Emit audit events
	audit.Emit(p.l, p.db.WithContext(p.ctx), e.TenantId, "recipe_ingredient", ingredientID, "normalization.corrected", uuid.Nil, map[string]interface{}{
		"recipe_id":               e.RecipeId,
		"canonical_ingredient_id": canonicalIngredientID,
		"save_as_alias":           saveAsAlias,
	})
	if aliasCreated {
		audit.Emit(p.l, p.db.WithContext(p.ctx), e.TenantId, "canonical_ingredient", canonicalIngredientID, "ingredient.alias_created", uuid.Nil, map[string]interface{}{
			"alias_name": m.RawName(),
		})
	}

	return ResolveResult{Model: m, AliasCreated: aliasCreated}, nil
}

func (p *Processor) createAliasIfNotExists(tenantID, canonicalIngredientID uuid.UUID, aliasName string) error {
	normalized := strings.ToLower(strings.TrimSpace(aliasName))

	if _, err := ingredient.GetByName(tenantID, normalized)(p.db.WithContext(p.ctx))(); err == nil {
		return ErrAliasConflict
	}
	if _, _, err := ingredient.GetByAlias(tenantID, normalized)(p.db.WithContext(p.ctx)); err == nil {
		return ErrAliasAlreadyExists
	}

	alias := ingredient.AliasEntity{
		Id:                    uuid.New(),
		TenantId:              tenantID,
		CanonicalIngredientId: canonicalIngredientID,
		Name:                  normalized,
		CreatedAt:             time.Now().UTC(),
	}
	return p.db.WithContext(p.ctx).Create(&alias).Error
}

type RenormalizeSummary struct {
	Total           int
	Changed         int
	StillUnresolved int
}

func (p *Processor) Renormalize(tenantID, recipeID uuid.UUID) ([]Model, RenormalizeSummary, error) {
	existing, err := getByRecipeID(recipeID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, RenormalizeSummary{}, err
	}

	summary := RenormalizeSummary{Total: len(existing)}
	var toUpdate []Entity

	for i := range existing {
		e := &existing[i]
		if Status(e.NormalizationStatus) == StatusManuallyConfirmed {
			continue
		}

		oldStatus := e.NormalizationStatus
		rawName := strings.ToLower(strings.TrimSpace(e.RawName))
		canonicalID, status := p.matchWithNormalization(tenantID, rawName)

		e.CanonicalIngredientId = canonicalID
		e.NormalizationStatus = string(status)
		if string(status) != oldStatus {
			summary.Changed++
		}
		if status == StatusUnresolved {
			summary.StillUnresolved++
		}
		toUpdate = append(toUpdate, *e)
	}

	if err := bulkUpdate(p.db.WithContext(p.ctx), toUpdate); err != nil {
		return nil, RenormalizeSummary{}, err
	}

	updated, err := getByRecipeID(recipeID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, RenormalizeSummary{}, err
	}

	models, err := entitiesToModels(updated)
	if err != nil {
		return nil, RenormalizeSummary{}, err
	}

	audit.Emit(p.l, p.db.WithContext(p.ctx), tenantID, "recipe", recipeID, "recipe.renormalized", uuid.Nil, map[string]interface{}{
		"total":            summary.Total,
		"changed":          summary.Changed,
		"still_unresolved": summary.StillUnresolved,
	})

	return models, summary, nil
}

func (p *Processor) GetByRecipeID(recipeID uuid.UUID) ([]Model, error) {
	entities, err := getByRecipeID(recipeID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	return entitiesToModels(entities)
}

func (p *Processor) PreviewNormalization(tenantID uuid.UUID, parsed []ParsedIngredient) []PreviewResult {
	results := make([]PreviewResult, len(parsed))
	for i, pi := range parsed {
		rawName := strings.ToLower(strings.TrimSpace(pi.Name))
		result := PreviewResult{
			RawName:  pi.Name,
			Position: i,
			Status:   StatusUnresolved,
		}

		if pi.Unit != "" {
			if identity, ok := LookupUnit(strings.ToLower(strings.TrimSpace(pi.Unit))); ok {
				result.CanonicalUnit = identity.Canonical
				result.CanonicalUnitFamily = identity.Family
			}
		}

		canonicalID, status := p.matchWithNormalization(tenantID, rawName)
		result.Status = status
		result.CanonicalIngredientID = canonicalID
		if canonicalID != nil {
			if e, err := ingredient.GetByName(tenantID, rawName)(p.db.WithContext(p.ctx))(); err == nil {
				result.CanonicalName = e.Name
			} else {
				normalized := normalizeText(rawName)
				if e, err := ingredient.GetByName(tenantID, normalized)(p.db.WithContext(p.ctx))(); err == nil {
					result.CanonicalName = e.Name
				}
			}
			// If matched via alias, get the canonical name from the entity by ID
			if result.CanonicalName == "" && canonicalID != nil {
				if e, err := ingredient.GetByID(*canonicalID)(p.db.WithContext(p.ctx))(); err == nil {
					result.CanonicalName = e.Name
				}
			}
		}

		results[i] = result
	}
	return results
}

type PreviewResult struct {
	RawName               string     `json:"rawName"`
	Position              int        `json:"position"`
	Status                Status     `json:"status"`
	CanonicalIngredientID *uuid.UUID `json:"canonicalIngredientId,omitempty"`
	CanonicalName         string     `json:"canonicalName,omitempty"`
	CanonicalUnit         string     `json:"canonicalUnit,omitempty"`
	CanonicalUnitFamily   string     `json:"canonicalUnitFamily,omitempty"`
}

func normalizeText(s string) string {
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

func entitiesToModels(entities []Entity) ([]Model, error) {
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
