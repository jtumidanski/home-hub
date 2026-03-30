package export

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/ingredient"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planner"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PlanData is a minimal representation of a plan, avoiding an import of the plan package.
type PlanData struct {
	ID       uuid.UUID
	Name     string
	StartsOn time.Time
}

// ConsolidatedIngredient represents a single line in the ingredient export.
type ConsolidatedIngredient struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
	Quantity    float64
	Unit        string
	UnitFamily  string
	Resolved    bool
}

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// ConsolidateIngredients builds a merged ingredient list from all items in a plan.
func (p *Processor) ConsolidateIngredients(pd PlanData) []ConsolidatedIngredient {
	itemProc := planitem.NewProcessor(p.l, p.ctx, p.db)
	items, err := itemProc.GetByPlanWeekID(pd.ID)
	if err != nil {
		p.l.WithError(err).Error("Failed to get plan items for consolidation")
		return nil
	}

	normProc := normalization.NewProcessor(p.l, p.ctx, p.db)
	recipeProc := recipe.NewProcessor(p.l, p.ctx, p.db)
	plannerProc := planner.NewProcessor(p.l, p.ctx, p.db)
	ingredientProc := ingredient.NewProcessor(p.l, p.ctx, p.db)

	// Key: canonical_ingredient_id + canonical_unit for resolved ingredients.
	type consolidationKey struct {
		CanonicalID uuid.UUID
		Unit        string
	}
	resolved := make(map[consolidationKey]*ConsolidatedIngredient)
	var unresolved []ConsolidatedIngredient

	for _, item := range items {
		multiplier := effectiveMultiplier(item, recipeProc, plannerProc)

		ingredients, err := normProc.GetByRecipeID(item.RecipeID())
		if err != nil || len(ingredients) == 0 {
			// Fall back to raw parsed Cooklang
			rm, _, recipeErr := recipeProc.Get(item.RecipeID())
			if recipeErr != nil {
				continue
			}
			parsed := recipeProc.ParseSource(rm.Source())
			for _, ing := range parsed.Ingredients {
				qty := parseQuantity(ing.Quantity) * multiplier
				if qty == 0 {
					continue
				}
				unresolved = append(unresolved, ConsolidatedIngredient{
					ID:       uuid.New(),
					Name:     ing.Name,
					Quantity: qty,
					Unit:     ing.Unit,
					Resolved: false,
				})
			}
			continue
		}

		for _, ing := range ingredients {
			qty := parseQuantity(ing.RawQuantity()) * multiplier
			if qty == 0 {
				continue
			}

			if ing.CanonicalIngredientID() != nil && ing.CanonicalUnit() != "" {
				key := consolidationKey{
					CanonicalID: *ing.CanonicalIngredientID(),
					Unit:        ing.CanonicalUnit(),
				}
				if existing, ok := resolved[key]; ok {
					existing.Quantity += qty
				} else {
					ci := ConsolidatedIngredient{
						ID:       *ing.CanonicalIngredientID(),
						Name:     ing.RawName(),
						Quantity: qty,
						Unit:     ing.CanonicalUnit(),
						Resolved: true,
					}
					// Look up display name
					if canonical, err := ingredientProc.Get(*ing.CanonicalIngredientID()); err == nil {
						ci.DisplayName = canonical.DisplayName()
						ci.Name = canonical.Name()
						ci.UnitFamily = canonical.UnitFamily()
					}
					resolved[key] = &ci
				}
			} else {
				unresolved = append(unresolved, ConsolidatedIngredient{
					ID:       uuid.New(),
					Name:     ing.RawName(),
					Quantity: qty,
					Unit:     ing.RawUnit(),
					Resolved: false,
				})
			}
		}
	}

	result := make([]ConsolidatedIngredient, 0, len(resolved)+len(unresolved))
	for _, ci := range resolved {
		result = append(result, *ci)
	}
	result = append(result, unresolved...)
	return result
}

// effectiveMultiplier computes the serving scaling for a plan item.
func effectiveMultiplier(item planitem.Model, recipeProc *recipe.Processor, plannerProc *planner.Processor) float64 {
	if item.PlannedServings() != nil {
		servingsYield := getServingsYield(item.RecipeID(), recipeProc, plannerProc)
		if servingsYield > 0 {
			return float64(*item.PlannedServings()) / float64(servingsYield)
		}
	}
	if item.ServingMultiplier() != nil {
		return *item.ServingMultiplier()
	}
	return 1.0
}

func getServingsYield(recipeID uuid.UUID, recipeProc *recipe.Processor, plannerProc *planner.Processor) int {
	if pc, err := plannerProc.GetByRecipeID(recipeID); err == nil && pc.ServingsYield() != nil {
		return *pc.ServingsYield()
	}
	if rm, _, err := recipeProc.Get(recipeID); err == nil && rm.Servings() != nil {
		return *rm.Servings()
	}
	return 0
}

func parseQuantity(raw string) float64 {
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	return v
}
