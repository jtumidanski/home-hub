package export

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/categoryclient"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// flatten extracts the names from a slice of CategoryGroup in render order so
// the assertions read like a snapshot of what the user would see.
func flatten(groups []CategoryGroup) []string {
	var out []string
	for _, g := range groups {
		heading := g.Name
		if heading == "" {
			heading = "Uncategorized"
		}
		out = append(out, "## "+heading)
		for _, ci := range g.Ingredients {
			name := ci.DisplayName
			if name == "" {
				name = ci.Name
			}
			out = append(out, "- "+name)
		}
	}
	return out
}

func TestGroupByCategory_AllCategorized(t *testing.T) {
	// Scenario (a): every ingredient is categorized — groups appear in
	// CategorySortOrder ascending, ingredients alphabetical within group.
	in := []ConsolidatedIngredient{
		{Name: "carrot", DisplayName: "Carrot", CategoryName: "Produce", CategorySortOrder: 1},
		{Name: "apple", DisplayName: "Apple", CategoryName: "Produce", CategorySortOrder: 1},
		{Name: "milk", DisplayName: "Milk", CategoryName: "Dairy", CategorySortOrder: 2},
		{Name: "butter", DisplayName: "Butter", CategoryName: "Dairy", CategorySortOrder: 2},
	}

	got := GroupByCategory(in)

	assert.Equal(t, []string{
		"## Produce",
		"- Apple",
		"- Carrot",
		"## Dairy",
		"- Butter",
		"- Milk",
	}, flatten(got))
}

func TestGroupByCategory_AllUncategorized(t *testing.T) {
	// Scenario (b): nothing is categorized — single trailing Uncategorized
	// group, alphabetical, no panic on empty CategoryName.
	in := []ConsolidatedIngredient{
		{Name: "salt", DisplayName: "Salt"},
		{Name: "pepper", DisplayName: "Pepper"},
		{Name: "olive oil"},
	}

	got := GroupByCategory(in)

	require.Len(t, got, 1)
	assert.Equal(t, "", got[0].Name)
	// Sort is case-sensitive (matches the pre-refactor processor.go behavior),
	// so capital-letter names sort ahead of lowercase ones.
	assert.Equal(t, []string{
		"## Uncategorized",
		"- Pepper",
		"- Salt",
		"- olive oil",
	}, flatten(got))
}

func TestGroupByCategory_Mixed(t *testing.T) {
	// Scenario (c): mixed categorized + uncategorized — uncategorized last
	// regardless of how it appears in the input.
	in := []ConsolidatedIngredient{
		{Name: "salt", DisplayName: "Salt"},
		{Name: "carrot", DisplayName: "Carrot", CategoryName: "Produce", CategorySortOrder: 1},
		{Name: "milk", DisplayName: "Milk", CategoryName: "Dairy", CategorySortOrder: 2},
		{Name: "pepper", DisplayName: "Pepper"},
	}

	got := GroupByCategory(in)

	require.Len(t, got, 3)
	assert.Equal(t, "Produce", got[0].Name)
	assert.Equal(t, "Dairy", got[1].Name)
	assert.Equal(t, "", got[2].Name)
	assert.Equal(t, []string{"Pepper", "Salt"}, []string{
		got[2].Ingredients[0].DisplayName,
		got[2].Ingredients[1].DisplayName,
	})
}

func TestGroupByCategory_UnresolvedPresent(t *testing.T) {
	// Scenario (d): unresolved ingredients are not dropped — they land in
	// the Uncategorized bucket alongside other category-less items.
	in := []ConsolidatedIngredient{
		{Name: "carrot", DisplayName: "Carrot", CategoryName: "Produce", CategorySortOrder: 1, Resolved: true},
		{Name: "garlic", Resolved: false},
		{Name: "onion", Resolved: false},
	}

	got := GroupByCategory(in)

	require.Len(t, got, 2)
	assert.Equal(t, "Produce", got[0].Name)
	assert.Equal(t, "", got[1].Name)
	assert.Equal(t, []string{"garlic", "onion"}, []string{
		got[1].Ingredients[0].Name,
		got[1].Ingredients[1].Name,
	})
}

func TestGroupByCategory_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		in   []ConsolidatedIngredient
		want []string
	}{
		{
			name: "empty input returns empty slice",
			in:   nil,
			want: nil,
		},
		{
			name: "single group",
			in: []ConsolidatedIngredient{
				{Name: "milk", DisplayName: "Milk", CategoryName: "Dairy", CategorySortOrder: 2},
			},
			want: []string{"## Dairy", "- Milk"},
		},
		{
			name: "ties on sort order break by name alphabetically",
			in: []ConsolidatedIngredient{
				{Name: "milk", DisplayName: "Milk", CategoryName: "Dairy", CategorySortOrder: 1},
				{Name: "apple", DisplayName: "Apple", CategoryName: "Produce", CategorySortOrder: 1},
			},
			want: []string{
				"## Dairy",
				"- Milk",
				"## Produce",
				"- Apple",
			},
		},
		{
			name: "DisplayName falls back to Name",
			in: []ConsolidatedIngredient{
				{Name: "zucchini", CategoryName: "Produce", CategorySortOrder: 1},
				{Name: "apple", DisplayName: "Apple", CategoryName: "Produce", CategorySortOrder: 1},
			},
			want: []string{
				"## Produce",
				"- Apple",
				"- zucchini",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flatten(GroupByCategory(tt.in))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadCategoryLookup_NilClientReturnsEmpty(t *testing.T) {
	l, _ := test.NewNullLogger()
	got := loadCategoryLookup(l, nil, "anything", uuid.New())
	assert.Empty(t, got)
}

func TestLoadCategoryLookup_EmptyTokenReturnsEmpty(t *testing.T) {
	l, _ := test.NewNullLogger()
	got := loadCategoryLookup(l, categoryclient.New("http://example.invalid"), "", uuid.New())
	assert.Empty(t, got)
}

func TestLoadCategoryLookup_ServerErrorLogsAndReturnsEmpty(t *testing.T) {
	// Scenario (e): categoryclient returns an error — every ingredient
	// must end up in the uncategorized bucket and a stable error log line
	// must be emitted with the plan_id field.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	l, hook := test.NewNullLogger()
	planID := uuid.New()

	got := loadCategoryLookup(l, categoryclient.New(srv.URL), "fake-token", planID)

	assert.Empty(t, got, "category map must be empty so caller falls back to Uncategorized")

	require.Len(t, hook.Entries, 1)
	entry := hook.Entries[0]
	assert.Equal(t, logrus.ErrorLevel, entry.Level)
	assert.Equal(t, "Failed to fetch categories for plan ingredient consolidation", entry.Message)
	assert.Equal(t, planID, entry.Data["plan_id"])
	assert.NotNil(t, entry.Data["error"])
}

func TestLoadCategoryLookup_PopulatesMapOnSuccess(t *testing.T) {
	categoryID := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":[{"id":"` + categoryID.String() + `","type":"categories","attributes":{"name":"Produce","sort_order":1}}]}`))
	}))
	defer srv.Close()

	l, _ := test.NewNullLogger()
	got := loadCategoryLookup(l, categoryclient.New(srv.URL), "fake-token", uuid.New())

	require.Len(t, got, 1)
	assert.Equal(t, "Produce", got[categoryID].name)
	assert.Equal(t, 1, got[categoryID].sortOrder)
}
