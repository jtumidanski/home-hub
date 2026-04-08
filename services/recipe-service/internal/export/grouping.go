package export

import (
	"sort"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/categoryclient"
	"github.com/sirupsen/logrus"
)

// catInfo is the per-category data the consolidation pipeline needs from
// category-service: a display name and a sort order.
type catInfo struct {
	name      string
	sortOrder int
}

// loadCategoryLookup fetches the tenant's categories via categoryclient and
// builds a uuid → catInfo map. On any error (including a nil client or
// missing access token, which can occur in tests) it returns an empty map
// and, when there is an actual fetch error, logs at error level with a
// stable, greppable message and the plan ID. Callers must treat the
// returned map as authoritative for their tenant — every category lookup
// miss after this call indicates either an unknown ID or a degraded fetch.
func loadCategoryLookup(l logrus.FieldLogger, client *categoryclient.Client, accessToken string, planID, tenantID, householdID uuid.UUID) map[uuid.UUID]catInfo {
	out := make(map[uuid.UUID]catInfo)
	if client == nil || accessToken == "" {
		// Diagnostic: this branch produces an empty categoryByID with no
		// HTTP call, which manifests downstream as a flood of "Canonical
		// ingredient references unknown category" warns and no error log.
		// Make the early-return loud so it's distinguishable from a real
		// fetch failure or a successful but mismatched fetch.
		l.WithFields(logrus.Fields{
			"plan_id":     planID,
			"client_nil":  client == nil,
			"token_empty": accessToken == "",
		}).Warn("Skipping category lookup; categoryclient unavailable or access token missing")
		return out
	}
	cats, err := client.ListCategories(accessToken, tenantID, householdID)
	if err != nil {
		l.WithError(err).WithField("plan_id", planID).
			Error("Failed to fetch categories for plan ingredient consolidation")
		return out
	}
	for _, c := range cats {
		out[c.ID] = catInfo{name: c.Name, sortOrder: c.SortOrder}
	}
	// Diagnostic: temporary info-level log so we can see, in deployed env,
	// exactly what categoryclient handed back. Compared against the user's
	// /api/v1/categories response, this tells us whether recipe-service is
	// (a) seeing zero categories, (b) seeing the user's real tenant, or
	// (c) seeing a different tenant entirely. Remove once the root cause
	// of the "Canonical ingredient references unknown category" floods
	// is identified.
	sampleIDs := make([]string, 0, 3)
	sampleNames := make([]string, 0, 3)
	for _, c := range cats {
		if len(sampleIDs) >= 3 {
			break
		}
		sampleIDs = append(sampleIDs, c.ID.String())
		sampleNames = append(sampleNames, c.Name)
	}
	l.WithFields(logrus.Fields{
		"plan_id":      planID,
		"count":        len(cats),
		"sample_ids":   sampleIDs,
		"sample_names": sampleNames,
	}).Info("categoryclient ListCategories returned")
	return out
}

// CategoryGroup is a single bucket of consolidated ingredients sharing a
// category. The empty Name represents the "Uncategorized" bucket, which is
// always rendered last regardless of SortOrder.
type CategoryGroup struct {
	Name        string
	SortOrder   int
	Ingredients []ConsolidatedIngredient
}

// GroupByCategory groups a flat slice of ConsolidatedIngredient by
// CategoryName. Groups are sorted by CategorySortOrder ascending; the
// uncategorized bucket (CategoryName == "") is always sorted last. Within
// each group, ingredients are sorted alphabetically by DisplayName, falling
// back to Name when DisplayName is empty.
//
// Both the JSON:API meal-plan ingredients endpoint and the markdown shopping
// list export consume this helper so that preview and export ordering can
// never drift apart.
func GroupByCategory(ingredients []ConsolidatedIngredient) []CategoryGroup {
	groupByName := make(map[string]*CategoryGroup)
	var uncategorized *CategoryGroup

	for _, ci := range ingredients {
		if ci.CategoryName == "" {
			if uncategorized == nil {
				uncategorized = &CategoryGroup{Name: ""}
			}
			uncategorized.Ingredients = append(uncategorized.Ingredients, ci)
			continue
		}
		g, ok := groupByName[ci.CategoryName]
		if !ok {
			g = &CategoryGroup{Name: ci.CategoryName, SortOrder: ci.CategorySortOrder}
			groupByName[ci.CategoryName] = g
		}
		g.Ingredients = append(g.Ingredients, ci)
	}

	groups := make([]*CategoryGroup, 0, len(groupByName))
	for _, g := range groupByName {
		groups = append(groups, g)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].SortOrder != groups[j].SortOrder {
			return groups[i].SortOrder < groups[j].SortOrder
		}
		return groups[i].Name < groups[j].Name
	})

	result := make([]CategoryGroup, 0, len(groups)+1)
	for _, g := range groups {
		sortIngredientsByDisplayName(g.Ingredients)
		result = append(result, *g)
	}
	if uncategorized != nil {
		sortIngredientsByDisplayName(uncategorized.Ingredients)
		result = append(result, *uncategorized)
	}
	return result
}

func sortIngredientsByDisplayName(items []ConsolidatedIngredient) {
	sort.SliceStable(items, func(i, j int) bool {
		ni := items[i].DisplayName
		if ni == "" {
			ni = items[i].Name
		}
		nj := items[j].DisplayName
		if nj == "" {
			nj = items[j].Name
		}
		return ni < nj
	})
}
