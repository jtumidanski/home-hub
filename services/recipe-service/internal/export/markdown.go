package export

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planner"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
)

var slotOrder = map[string]int{
	planitem.SlotBreakfast: 0,
	planitem.SlotLunch:     1,
	planitem.SlotDinner:    2,
	planitem.SlotSnack:     3,
	planitem.SlotSide:      4,
}

// GenerateMarkdown produces a markdown string for the given plan and its items.
func (p *Processor) GenerateMarkdown(pd PlanData) string {
	itemProc := planitem.NewProcessor(p.l, p.ctx, p.db)
	recipeProc := recipe.NewProcessor(p.l, p.ctx, p.db)
	plannerProc := planner.NewProcessor(p.l, p.ctx, p.db)

	items, err := itemProc.GetByPlanWeekID(pd.ID)
	if err != nil {
		p.l.WithError(err).Error("Failed to get plan items for markdown export")
		return ""
	}

	var sb strings.Builder

	// Heading
	sb.WriteString(fmt.Sprintf("# Meal Plan: %s\n", pd.Name))

	// Group items by day
	type dayGroup struct {
		date  time.Time
		items []planitem.Model
	}
	dayMap := make(map[string]*dayGroup)
	for _, item := range items {
		key := item.Day().Format("2006-01-02")
		if _, ok := dayMap[key]; !ok {
			dayMap[key] = &dayGroup{date: item.Day()}
		}
		dayMap[key].items = append(dayMap[key].items, item)
	}

	// Sort days in calendar order
	sortedDays := make([]*dayGroup, 0, len(dayMap))
	for _, dg := range dayMap {
		sortedDays = append(sortedDays, dg)
	}
	sort.Slice(sortedDays, func(i, j int) bool {
		return sortedDays[i].date.Before(sortedDays[j].date)
	})

	// Write each day
	for _, dg := range sortedDays {
		dayName := dg.date.Format("Monday")
		dateStr := dg.date.Format("2006-01-02")
		sb.WriteString(fmt.Sprintf("\n## %s (%s)\n", dayName, dateStr))

		// Sort items by slot order then position
		sort.Slice(dg.items, func(i, j int) bool {
			oi, oj := slotOrder[dg.items[i].Slot()], slotOrder[dg.items[j].Slot()]
			if oi != oj {
				return oi < oj
			}
			return dg.items[i].Position() < dg.items[j].Position()
		})

		for _, item := range dg.items {
			slotLabel := capitalizeFirst(item.Slot())
			title := "(deleted recipe)"

			rm, _, recipeErr := recipeProc.Get(item.RecipeID())
			if recipeErr == nil {
				title = rm.Title()
			}

			servingsNote := servingsAnnotation(item, recipeProc, plannerProc)
			if servingsNote != "" {
				sb.WriteString(fmt.Sprintf("- **%s:** %s (%s)\n", slotLabel, title, servingsNote))
			} else {
				sb.WriteString(fmt.Sprintf("- **%s:** %s\n", slotLabel, title))
			}
		}
	}

	// Consolidated ingredients grouped by category
	consolidated := p.ConsolidateIngredients(pd)
	if len(consolidated) > 0 {
		sb.WriteString("\n## Shopping List\n")

		// Group by category
		type categoryGroup struct {
			name        string
			sortOrder   int
			ingredients []ConsolidatedIngredient
		}
		groupMap := make(map[string]*categoryGroup)
		var uncategorized []ConsolidatedIngredient

		for _, ci := range consolidated {
			if ci.CategoryName == "" {
				uncategorized = append(uncategorized, ci)
			} else {
				g, ok := groupMap[ci.CategoryName]
				if !ok {
					g = &categoryGroup{name: ci.CategoryName, sortOrder: ci.CategorySortOrder}
					groupMap[ci.CategoryName] = g
				}
				g.ingredients = append(g.ingredients, ci)
			}
		}

		// Sort groups by sort order
		groups := make([]*categoryGroup, 0, len(groupMap))
		for _, g := range groupMap {
			groups = append(groups, g)
		}
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].sortOrder < groups[j].sortOrder
		})

		for _, g := range groups {
			sb.WriteString(fmt.Sprintf("\n### %s\n", g.name))
			for _, ci := range g.ingredients {
				name := ci.DisplayName
				if name == "" {
					name = ci.Name
				}
				sb.WriteString(fmt.Sprintf("- %s %s %s\n", formatQuantity(ci.Quantity), ci.Unit, name))
			}
		}

		if len(uncategorized) > 0 {
			sb.WriteString("\n### Uncategorized\n")
			for _, ci := range uncategorized {
				name := ci.DisplayName
				if name == "" {
					name = ci.Name
				}
				if !ci.Resolved {
					sb.WriteString(fmt.Sprintf("- %s %s %s _(unresolved)_\n", formatQuantity(ci.Quantity), ci.Unit, name))
				} else {
					sb.WriteString(fmt.Sprintf("- %s %s %s\n", formatQuantity(ci.Quantity), ci.Unit, name))
				}
			}
		}
	}

	// Notes section for unresolved
	hasUnresolved := false
	for _, ci := range consolidated {
		if !ci.Resolved {
			hasUnresolved = true
			break
		}
	}
	if hasUnresolved {
		sb.WriteString("\n## Notes\n")
		sb.WriteString("- Some ingredients could not be fully consolidated and were listed as entered.\n")
	}

	return sb.String()
}

// servingsAnnotation returns "(serves N)" if servings differ from recipe default.
func servingsAnnotation(item planitem.Model, recipeProc *recipe.Processor, plannerProc *planner.Processor) string {
	if item.PlannedServings() != nil {
		return fmt.Sprintf("serves %d", *item.PlannedServings())
	}
	if item.ServingMultiplier() != nil && *item.ServingMultiplier() != 1.0 {
		yield := getServingsYield(item.RecipeID(), recipeProc, plannerProc)
		if yield > 0 {
			effectiveServings := float64(yield) * *item.ServingMultiplier()
			return fmt.Sprintf("serves %s", formatQuantity(effectiveServings))
		}
		return fmt.Sprintf("×%s", formatQuantity(*item.ServingMultiplier()))
	}
	return ""
}

func formatQuantity(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%d", int(v))
	}
	return fmt.Sprintf("%.1f", v)
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
