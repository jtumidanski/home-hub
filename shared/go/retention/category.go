// Package retention provides shared building blocks for the data retention
// framework: category enum + defaults, jittered loop, advisory locking,
// policy client with safety guard, audit writer, and Prometheus metrics.
package retention

import (
	"errors"
	"strings"
)

// Category identifies a single retention rule, e.g. "productivity.completed_tasks".
type Category string

const (
	CatProductivityCompletedTasks            Category = "productivity.completed_tasks"
	CatProductivityDeletedTasksRestoreWindow Category = "productivity.deleted_tasks_restore_window"
	CatRecipeDeletedRecipesRestoreWindow     Category = "recipe.deleted_recipes_restore_window"
	CatRecipeRestorationAudit                Category = "recipe.restoration_audit"
	CatTrackerEntries                        Category = "tracker.entries"
	CatTrackerDeletedItemsRestoreWindow      Category = "tracker.deleted_items_restore_window"
	CatWorkoutPerformances                   Category = "workout.performances"
	CatWorkoutDeletedCatalogRestoreWindow    Category = "workout.deleted_catalog_restore_window"
	CatCalendarPastEvents                    Category = "calendar.past_events"
	CatPackageArchiveWindow                  Category = "package.archive_window"
	CatPackageArchivedDeleteWindow           Category = "package.archived_delete_window"
	CatSystemRetentionAudit                  Category = "system.retention_audit"
)

// ScopeKind is the kind of entity a retention category applies to.
type ScopeKind string

const (
	ScopeHousehold ScopeKind = "household"
	ScopeUser      ScopeKind = "user"
)

// Defaults are the system-shipped retention windows in days.
// These are compiled into binaries so they version with the deployment.
var Defaults = map[Category]int{
	CatProductivityCompletedTasks:            365,
	CatProductivityDeletedTasksRestoreWindow: 30,
	CatRecipeDeletedRecipesRestoreWindow:     30,
	CatRecipeRestorationAudit:                90,
	CatTrackerEntries:                        730,
	CatTrackerDeletedItemsRestoreWindow:      30,
	CatWorkoutPerformances:                   1825,
	CatWorkoutDeletedCatalogRestoreWindow:    30,
	CatCalendarPastEvents:                    365,
	CatPackageArchiveWindow:                  7,
	CatPackageArchivedDeleteWindow:           30,
	CatSystemRetentionAudit:                  180,
}

// scopeKindOf maps each category to the entity scope it applies to.
var scopeKindOf = map[Category]ScopeKind{
	CatProductivityCompletedTasks:            ScopeHousehold,
	CatProductivityDeletedTasksRestoreWindow: ScopeHousehold,
	CatRecipeDeletedRecipesRestoreWindow:     ScopeHousehold,
	CatRecipeRestorationAudit:                ScopeHousehold,
	CatTrackerEntries:                        ScopeUser,
	CatTrackerDeletedItemsRestoreWindow:      ScopeUser,
	CatWorkoutPerformances:                   ScopeUser,
	CatWorkoutDeletedCatalogRestoreWindow:    ScopeUser,
	CatCalendarPastEvents:                    ScopeHousehold,
	CatPackageArchiveWindow:                  ScopeHousehold,
	CatPackageArchivedDeleteWindow:           ScopeHousehold,
	CatSystemRetentionAudit:                  ScopeHousehold,
}

// Errors returned by Category methods.
var (
	ErrUnknownCategory = errors.New("retention: unknown category")
	ErrDaysTooLow      = errors.New("retention: days must be >= 1")
	ErrDaysTooHigh     = errors.New("retention: days exceeds maximum")
)

// IsKnown reports whether the category is in the registry.
func (c Category) IsKnown() bool {
	_, ok := Defaults[c]
	return ok
}

// IsHouseholdScoped reports whether the category applies to a household scope.
func (c Category) IsHouseholdScoped() bool {
	return scopeKindOf[c] == ScopeHousehold
}

// IsUserScoped reports whether the category applies to a user scope.
func (c Category) IsUserScoped() bool {
	return scopeKindOf[c] == ScopeUser
}

// Scope returns the ScopeKind this category belongs to.
func (c Category) Scope() ScopeKind { return scopeKindOf[c] }

// MaxDays returns the upper bound for days values for this category.
// Soft-delete restore windows cap at 365 days, all others at 3650.
func (c Category) MaxDays() int {
	if strings.HasSuffix(string(c), "_restore_window") {
		return 365
	}
	return 3650
}

// Validate verifies that days is within bounds for this category and that
// the category itself is known.
func (c Category) Validate(days int) error {
	if !c.IsKnown() {
		return ErrUnknownCategory
	}
	if days < 1 {
		return ErrDaysTooLow
	}
	if days > c.MaxDays() {
		return ErrDaysTooHigh
	}
	return nil
}

// All returns every known category in deterministic order (sorted alphabetically).
func All() []Category {
	out := make([]Category, 0, len(Defaults))
	for c := range Defaults {
		out = append(out, c)
	}
	// stable sort by string
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

// HouseholdCategories returns all household-scoped categories.
func HouseholdCategories() []Category {
	out := make([]Category, 0)
	for _, c := range All() {
		if c.IsHouseholdScoped() {
			out = append(out, c)
		}
	}
	return out
}

// UserCategories returns all user-scoped categories.
func UserCategories() []Category {
	out := make([]Category, 0)
	for _, c := range All() {
		if c.IsUserScoped() {
			out = append(out, c)
		}
	}
	return out
}
