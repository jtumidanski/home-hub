package export

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/categoryclient"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/ingredient"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planner"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// tokenPrefix returns the first 12 characters of an access token (the JWT
// header — never the body or signature) so diagnostic logs can correlate a
// recipe-service outbound call with the cookie value the browser is sending,
// without writing any sensitive claim or signature material to logs.
func tokenPrefix(token string) string {
	if len(token) <= 12 {
		return ""
	}
	return token[:12]
}

// PlanData is a minimal representation of a plan, avoiding an import of the plan package.
type PlanData struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	StartsOn    time.Time
	AccessToken string
}

// QuantityUnit is an additional quantity+unit pair for cross-family grouping.
type QuantityUnit struct {
	Quantity float64
	Unit     string
}

// ConsolidatedIngredient represents a single line in the ingredient export.
type ConsolidatedIngredient struct {
	ID                uuid.UUID
	Name              string
	DisplayName       string
	Quantity          float64
	Unit              string
	UnitFamily        string
	Resolved          bool
	ExtraQuantities   []QuantityUnit
	CategoryName      string
	CategorySortOrder int
}

type Processor struct {
	l         logrus.FieldLogger
	ctx       context.Context
	db        *gorm.DB
	catClient *categoryclient.Client
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, catClient *categoryclient.Client) *Processor {
	return &Processor{l: l, ctx: ctx, db: db, catClient: catClient}
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

	// Batch-fetch all per-recipe data referenced by the plan items in a fixed
	// number of queries (one per source) instead of N per item.
	recipeIDs := make([]uuid.UUID, 0, len(items))
	seenRecipe := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		rid := item.RecipeID()
		if _, ok := seenRecipe[rid]; ok {
			continue
		}
		seenRecipe[rid] = struct{}{}
		recipeIDs = append(recipeIDs, rid)
	}

	normByRecipe, err := normProc.GetByRecipeIDs(recipeIDs)
	if err != nil {
		p.l.WithError(err).Error("Failed to batch-fetch normalized ingredients")
		normByRecipe = map[uuid.UUID][]normalization.Model{}
	}
	plannerByRecipe, err := plannerProc.GetByRecipeIDs(recipeIDs)
	if err != nil {
		p.l.WithError(err).Error("Failed to batch-fetch planner configs")
		plannerByRecipe = map[uuid.UUID]planner.Model{}
	}
	recipesByID, err := recipeProc.GetByIDs(recipeIDs)
	if err != nil {
		p.l.WithError(err).Error("Failed to batch-fetch recipes")
		recipesByID = map[uuid.UUID]recipe.Model{}
	}

	// Accumulator per canonical ingredient, keyed by base unit.
	type baseUnitAccum struct {
		baseUnit string
		qty      float64
	}
	type ingredientAccum struct {
		id                uuid.UUID
		name              string
		displayName       string
		unitFamily        string
		categoryName      string
		categorySortOrder int
		units             map[string]*baseUnitAccum // base unit → accumulated qty
	}
	resolved := make(map[uuid.UUID]*ingredientAccum)
	var unresolved []ConsolidatedIngredient

	// Diagnostic: log the request tenant context alongside the plan's
	// own tenant_id and a fingerprint of the access token. This is the
	// next step in tracking down why categoryclient is returning a
	// different set of categories than the user's browser sees from
	// /api/v1/categories — comparing the request tenant to the browser
	// tenant tells us whether the inbound auth is misrouted vs. the
	// outbound categoryclient call hitting a different tenant.
	diagFields := logrus.Fields{
		"plan_id":          pd.ID,
		"plan_tenant_id":   pd.TenantID,
		"token_len":        len(pd.AccessToken),
		"token_prefix":     tokenPrefix(pd.AccessToken),
	}
	if t, ok := tenantctx.FromContext(p.ctx); ok {
		diagFields["request_tenant_id"] = t.Id()
		diagFields["request_household_id"] = t.HouseholdId()
		diagFields["request_user_id"] = t.UserId()
	} else {
		diagFields["request_tenant_id"] = "<missing from context>"
	}
	p.l.WithFields(diagFields).Info("ConsolidateIngredients tenant context snapshot")

	// Build category lookup map for sort order. On any categoryclient
	// failure we log and fall back to an empty map so the endpoint stays
	// 200 OK with everything in "Uncategorized".
	categoryByID := loadCategoryLookup(p.l, p.catClient, pd.AccessToken, pd.ID)

	// effectiveMultiplierFromMaps computes the serving multiplier for an item
	// using only the pre-fetched maps, no DB calls.
	effectiveMultiplierFromMaps := func(item planitem.Model) float64 {
		if item.PlannedServings() != nil {
			yield := 0
			if pc, ok := plannerByRecipe[item.RecipeID()]; ok && pc.ServingsYield() != nil {
				yield = *pc.ServingsYield()
			} else if rm, ok := recipesByID[item.RecipeID()]; ok && rm.Servings() != nil {
				yield = *rm.Servings()
			}
			if yield > 0 {
				return float64(*item.PlannedServings()) / float64(yield)
			}
		}
		if item.ServingMultiplier() != nil {
			return *item.ServingMultiplier()
		}
		return 1.0
	}

	for _, item := range items {
		multiplier := effectiveMultiplierFromMaps(item)

		ingredients := normByRecipe[item.RecipeID()]
		if len(ingredients) == 0 {
			// Fall back to raw parsed Cooklang
			rm, ok := recipesByID[item.RecipeID()]
			if !ok {
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

			if ing.CanonicalIngredientID() != nil {
				canonID := *ing.CanonicalIngredientID()
				unit := ing.CanonicalUnit()
				if unit == "" {
					unit = ing.RawUnit()
				}
				baseUnit, factor := toBaseUnit(unit)
				normalizedQty := qty * factor

				acc, ok := resolved[canonID]
				if !ok {
					acc = &ingredientAccum{
						id:    canonID,
						name:  ing.RawName(),
						units: make(map[string]*baseUnitAccum),
					}
					resolved[canonID] = acc
				}

				if existing, ok := acc.units[baseUnit]; ok {
					existing.qty += normalizedQty
				} else {
					acc.units[baseUnit] = &baseUnitAccum{baseUnit: baseUnit, qty: normalizedQty}
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

	// Second pass: fetch all canonical ingredients (with aliases) at once and
	// fill in displayName / canonical name / unit family / category info.
	canonIDs := make([]uuid.UUID, 0, len(resolved))
	for id := range resolved {
		canonIDs = append(canonIDs, id)
	}
	canonicalsByID, err := ingredientProc.GetByIDs(canonIDs)
	if err != nil {
		p.l.WithError(err).Error("Failed to batch-fetch canonical ingredients")
		canonicalsByID = map[uuid.UUID]ingredient.Model{}
	}
	if len(canonicalsByID) < len(canonIDs) {
		p.l.WithFields(logrus.Fields{
			"plan_id":       pd.ID,
			"requested":     len(canonIDs),
			"received":      len(canonicalsByID),
			"missing_count": len(canonIDs) - len(canonicalsByID),
		}).Warn("Canonical ingredient batch fetch returned fewer rows than requested")
	}
	for canonID, acc := range resolved {
		canonical, ok := canonicalsByID[canonID]
		if !ok {
			continue
		}
		acc.displayName = canonical.DisplayName()
		acc.name = canonical.Name()
		acc.unitFamily = canonical.UnitFamily()
		if canonical.CategoryID() != nil {
			if ci, ok := categoryByID[*canonical.CategoryID()]; ok {
				acc.categoryName = ci.name
				acc.categorySortOrder = ci.sortOrder
			} else {
				p.l.WithFields(logrus.Fields{
					"plan_id":                 pd.ID,
					"canonical_ingredient_id": canonID,
					"category_id":             *canonical.CategoryID(),
				}).Warn("Canonical ingredient references unknown category")
			}
		}
	}

	// Materialize each accumulator into a ConsolidatedIngredient. Order is
	// finalized below by GroupByCategory so that the JSON:API response and
	// the markdown shopping list export share a single ordering rule.
	flat := make([]ConsolidatedIngredient, 0, len(resolved)+len(unresolved))
	for _, acc := range resolved {
		var pairs []QuantityUnit
		for _, bu := range acc.units {
			displayUnit, displayQty := fromBaseUnit(bu.baseUnit, bu.qty)
			pairs = append(pairs, QuantityUnit{Quantity: displayQty, Unit: displayUnit})
		}
		ci := ConsolidatedIngredient{
			ID:                acc.id,
			Name:              acc.name,
			DisplayName:       acc.displayName,
			Quantity:          pairs[0].Quantity,
			Unit:              pairs[0].Unit,
			UnitFamily:        acc.unitFamily,
			Resolved:          true,
			CategoryName:      acc.categoryName,
			CategorySortOrder: acc.categorySortOrder,
		}
		if len(pairs) > 1 {
			ci.ExtraQuantities = pairs[1:]
		}
		flat = append(flat, ci)
	}
	flat = append(flat, unresolved...)

	groups := GroupByCategory(flat)
	result := make([]ConsolidatedIngredient, 0, len(flat))
	for _, g := range groups {
		result = append(result, g.Ingredients...)
	}
	return result
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

// fractionRe matches a pure fraction like "1/2" or "3/4"
var fractionRe = regexp.MustCompile(`^(\d+)/(\d+)`)

// mixedNumberRe matches "2 1/4" style mixed numbers
var mixedNumberRe = regexp.MustCompile(`^(\d+)\s+(\d+)/(\d+)`)

// leadingIntRe matches a leading integer or decimal
var leadingIntRe = regexp.MustCompile(`^(\d+(?:\.\d+)?)`)

func parseQuantity(raw string) float64 {
	if raw == "" {
		return 0
	}
	raw = strings.TrimSpace(raw)
	// Handle additive expressions: "1 + 1", "to taste + to taste"
	if strings.Contains(raw, " + ") {
		parts := strings.Split(raw, " + ")
		var total float64
		for _, p := range parts {
			v := parseQuantity(strings.TrimSpace(p))
			total += v
		}
		return total
	}
	// Try direct float first (e.g., "1.5", "3")
	if v, err := strconv.ParseFloat(raw, 64); err == nil {
		return v
	}
	// Try mixed number: "2 1/4"
	if m := mixedNumberRe.FindStringSubmatch(raw); m != nil {
		whole, _ := strconv.ParseFloat(m[1], 64)
		num, _ := strconv.ParseFloat(m[2], 64)
		den, _ := strconv.ParseFloat(m[3], 64)
		if den != 0 {
			return whole + num/den
		}
	}
	// Try fraction: "1/2", "3/4"
	if m := fractionRe.FindStringSubmatch(raw); m != nil {
		num, _ := strconv.ParseFloat(m[1], 64)
		den, _ := strconv.ParseFloat(m[2], 64)
		if den != 0 {
			return num / den
		}
	}
	// Extract leading integer from complex strings: "3 small-medium, ..."
	if m := leadingIntRe.FindStringSubmatch(raw); m != nil {
		v, _ := strconv.ParseFloat(m[1], 64)
		return v
	}
	return 0
}

// toBaseUnit converts a canonical unit to its family's base unit and returns
// the conversion factor. E.g., "tablespoon" → ("teaspoon", 3.0).
// Units without a known conversion are returned as-is with factor 1.
func toBaseUnit(unit string) (string, float64) {
	switch unit {
	// Volume → teaspoon
	case "teaspoon":
		return "teaspoon", 1
	case "tablespoon":
		return "teaspoon", 3
	case "cup":
		return "teaspoon", 48
	case "fluid ounce":
		return "teaspoon", 6
	case "milliliter":
		return "teaspoon", 0.202884
	case "liter":
		return "teaspoon", 202.884
	// Weight → ounce
	case "gram":
		return "gram", 1
	case "kilogram":
		return "gram", 1000
	case "ounce":
		return "ounce", 1
	case "pound":
		return "ounce", 16
	default:
		return unit, 1
	}
}

// volumeSteps are volume units ordered from largest to smallest with their
// teaspoon equivalents.
var volumeSteps = []struct {
	unit      string
	teaspoons float64
}{
	{"cup", 48},
	{"tablespoon", 3},
	{"teaspoon", 1},
}

// weightOzSteps for ounce-family.
var weightOzSteps = []struct {
	unit   string
	ounces float64
}{
	{"pound", 16},
	{"ounce", 1},
}

// fromBaseUnit converts a quantity in a base unit to the most readable display unit.
func fromBaseUnit(baseUnit string, qty float64) (string, float64) {
	switch baseUnit {
	case "teaspoon":
		for _, step := range volumeSteps {
			converted := qty / step.teaspoons
			if converted >= 1 {
				return step.unit, converted
			}
		}
		return "teaspoon", qty
	case "gram":
		if qty >= 1000 {
			return "kilogram", qty / 1000
		}
		return "gram", qty
	case "ounce":
		for _, step := range weightOzSteps {
			converted := qty / step.ounces
			if converted >= 1 {
				return step.unit, converted
			}
		}
		return "ounce", qty
	default:
		return baseUnit, qty
	}
}
