package export

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"math"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/categoryclient"
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
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	StartsOn   time.Time
	AuthHeader string
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

	// Build category lookup map for sort order
	type catInfo struct {
		name      string
		sortOrder int
	}
	categoryByID := make(map[uuid.UUID]catInfo)
	if p.catClient != nil && pd.AuthHeader != "" {
		if cats, err := p.catClient.ListCategories(pd.AuthHeader); err == nil {
			for _, c := range cats {
				categoryByID[c.ID] = catInfo{name: c.Name, sortOrder: c.SortOrder}
			}
		}
	}

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
			}
		}
	}

	result := make([]ConsolidatedIngredient, 0, len(resolved)+len(unresolved))
	// Sort resolved ingredients by category sort order, then alphabetically by display name.
	sortedAccums := make([]*ingredientAccum, 0, len(resolved))
	for _, acc := range resolved {
		sortedAccums = append(sortedAccums, acc)
	}
	sort.Slice(sortedAccums, func(i, j int) bool {
		// Uncategorized (sortOrder 0, no category) sorts last
		oi, oj := sortedAccums[i].categorySortOrder, sortedAccums[j].categorySortOrder
		if sortedAccums[i].categoryName == "" {
			oi = math.MaxInt32
		}
		if sortedAccums[j].categoryName == "" {
			oj = math.MaxInt32
		}
		if oi != oj {
			return oi < oj
		}
		ni, nj := sortedAccums[i].displayName, sortedAccums[j].displayName
		if ni == "" {
			ni = sortedAccums[i].name
		}
		if nj == "" {
			nj = sortedAccums[j].name
		}
		return ni < nj
	})
	for _, acc := range sortedAccums {
		// Collect all base-unit entries, convert to display units
		var pairs []QuantityUnit
		for _, bu := range acc.units {
			displayUnit, displayQty := fromBaseUnit(bu.baseUnit, bu.qty)
			pairs = append(pairs, QuantityUnit{Quantity: displayQty, Unit: displayUnit})
		}
		// First pair is primary, rest are extras
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
		result = append(result, ci)
	}
	result = append(result, unresolved...)
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
