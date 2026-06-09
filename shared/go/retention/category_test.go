package retention

import (
	"errors"
	"testing"
)

func TestCategoryValidate(t *testing.T) {
	tests := []struct {
		name string
		cat  Category
		days int
		want error
	}{
		{"unknown", Category("nope.invalid"), 30, ErrUnknownCategory},
		{"too low", CatProductivityCompletedTasks, 0, ErrDaysTooLow},
		{"negative", CatProductivityCompletedTasks, -1, ErrDaysTooLow},
		{"max ok normal", CatProductivityCompletedTasks, 3650, nil},
		{"too high normal", CatProductivityCompletedTasks, 3651, ErrDaysTooHigh},
		{"max ok restore", CatProductivityDeletedTasksRestoreWindow, 365, nil},
		{"too high restore", CatProductivityDeletedTasksRestoreWindow, 366, ErrDaysTooHigh},
		{"min ok", CatProductivityCompletedTasks, 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cat.Validate(tt.days); !errors.Is(got, tt.want) {
				t.Errorf("Validate(%d) = %v, want %v", tt.days, got, tt.want)
			}
		})
	}
}

func TestCategoryScope(t *testing.T) {
	if !CatProductivityCompletedTasks.IsHouseholdScoped() {
		t.Error("productivity.completed_tasks should be household-scoped")
	}
	if !CatTrackerEntries.IsUserScoped() {
		t.Error("tracker.entries should be user-scoped")
	}
	if CatProductivityCompletedTasks.IsUserScoped() {
		t.Error("productivity should not be user-scoped")
	}
}

func TestDashboardCategory(t *testing.T) {
	if !CatDashboardDashboards.IsKnown() {
		t.Fatal("CatDashboardDashboards should be known")
	}
	if !CatDashboardDashboards.IsHouseholdScoped() {
		t.Fatal("dashboards should be household-scoped")
	}
	if Defaults[CatDashboardDashboards] != 0 {
		t.Fatalf("default for dashboards should be 0 (never auto-purge), got %d", Defaults[CatDashboardDashboards])
	}
}

func TestReminderCategories(t *testing.T) {
	// Known + household-scoped.
	for _, c := range []Category{CatProductivityReminders, CatProductivityDeletedRemindersRestoreWindow} {
		if !c.IsKnown() {
			t.Errorf("%s should be known", c)
		}
		if !c.IsHouseholdScoped() {
			t.Errorf("%s should be household-scoped", c)
		}
	}
	// Defaults.
	if Defaults[CatProductivityReminders] != 365 {
		t.Errorf("reminders default = %d, want 365", Defaults[CatProductivityReminders])
	}
	if Defaults[CatProductivityDeletedRemindersRestoreWindow] != 30 {
		t.Errorf("restore-window default = %d, want 30", Defaults[CatProductivityDeletedRemindersRestoreWindow])
	}
	// MaxDays: primary 3650, restore window capped at 365 by the suffix.
	if CatProductivityReminders.MaxDays() != 3650 {
		t.Errorf("reminders MaxDays = %d, want 3650", CatProductivityReminders.MaxDays())
	}
	if CatProductivityDeletedRemindersRestoreWindow.MaxDays() != 365 {
		t.Errorf("restore-window MaxDays = %d, want 365", CatProductivityDeletedRemindersRestoreWindow.MaxDays())
	}
	// Both appear in the household enumeration.
	found := map[Category]bool{}
	for _, c := range HouseholdCategories() {
		found[c] = true
	}
	if !found[CatProductivityReminders] || !found[CatProductivityDeletedRemindersRestoreWindow] {
		t.Errorf("HouseholdCategories missing reminder categories: %v", HouseholdCategories())
	}
}

func TestDefaultsCoverage(t *testing.T) {
	for _, c := range All() {
		if _, ok := Defaults[c]; !ok {
			t.Errorf("category %s missing from Defaults", c)
		}
		if _, ok := scopeKindOf[c]; !ok {
			t.Errorf("category %s missing from scopeKindOf", c)
		}
	}
}
