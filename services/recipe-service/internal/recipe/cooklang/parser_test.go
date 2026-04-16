package cooklang

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_BasicIngredient(t *testing.T) {
	result := Parse("Add @salt{1%tsp} to the pot.")
	require.Len(t, result.Steps, 1)
	require.Len(t, result.Ingredients, 1)

	assert.Equal(t, "salt", result.Ingredients[0].Name)
	assert.Equal(t, "1", result.Ingredients[0].Quantity)
	assert.Equal(t, "tsp", result.Ingredients[0].Unit)

	step := result.Steps[0]
	assert.Equal(t, 1, step.Number)
	require.Len(t, step.Segments, 3)
	assert.Equal(t, SegmentText, step.Segments[0].Type)
	assert.Equal(t, "Add ", step.Segments[0].Value)
	assert.Equal(t, SegmentIngredient, step.Segments[1].Type)
	assert.Equal(t, "salt", step.Segments[1].Name)
	assert.Equal(t, SegmentText, step.Segments[2].Type)
	assert.Equal(t, " to the pot.", step.Segments[2].Value)
}

func TestParse_MultiWordIngredient(t *testing.T) {
	result := Parse("Add @pecorino romano{100%g}.")
	require.Len(t, result.Ingredients, 1)
	assert.Equal(t, "pecorino romano", result.Ingredients[0].Name)
	assert.Equal(t, "100", result.Ingredients[0].Quantity)
	assert.Equal(t, "g", result.Ingredients[0].Unit)
}

func TestParse_IngredientWithoutBraces(t *testing.T) {
	result := Parse("Add @salt and stir.")
	require.Len(t, result.Ingredients, 1)
	assert.Equal(t, "salt", result.Ingredients[0].Name)
	assert.Equal(t, "", result.Ingredients[0].Quantity)
}

func TestParse_IngredientWithQuantityNoUnit(t *testing.T) {
	result := Parse("Add @eggs{3}.")
	require.Len(t, result.Ingredients, 1)
	assert.Equal(t, "eggs", result.Ingredients[0].Name)
	assert.Equal(t, "3", result.Ingredients[0].Quantity)
	assert.Equal(t, "", result.Ingredients[0].Unit)
}

func TestParse_Cookware(t *testing.T) {
	result := Parse("Place in a #large pot{} and boil.")
	require.Len(t, result.Steps, 1)

	segments := result.Steps[0].Segments
	found := false
	for _, s := range segments {
		if s.Type == SegmentCookware {
			assert.Equal(t, "large pot", s.Name)
			found = true
		}
	}
	assert.True(t, found)
}

func TestParse_Timer(t *testing.T) {
	result := Parse("Cook for ~{8%minutes}.")
	require.Len(t, result.Steps, 1)

	segments := result.Steps[0].Segments
	found := false
	for _, s := range segments {
		if s.Type == SegmentTimer {
			assert.Equal(t, "8", s.Quantity)
			assert.Equal(t, "minutes", s.Unit)
			found = true
		}
	}
	assert.True(t, found)
}

func TestParse_MultipleSteps(t *testing.T) {
	source := "Boil @water{2%l}.\n\nAdd @spaghetti{400%g}.\n\nDrain and serve."
	result := Parse(source)
	assert.Len(t, result.Steps, 3)
	assert.Equal(t, 1, result.Steps[0].Number)
	assert.Equal(t, 2, result.Steps[1].Number)
	assert.Equal(t, 3, result.Steps[2].Number)
}

func TestParse_IngredientAggregation(t *testing.T) {
	source := "Add @butter{50%g}.\n\nAdd more @butter{50%g}."
	result := Parse(source)
	require.Len(t, result.Ingredients, 1)
	assert.Equal(t, "butter", result.Ingredients[0].Name)
	assert.Equal(t, "100", result.Ingredients[0].Quantity)
	assert.Equal(t, "g", result.Ingredients[0].Unit)
}

func TestParse_LineComment(t *testing.T) {
	source := "Add @salt{1%tsp}. -- this is a comment\nStir well."
	result := Parse(source)
	require.Len(t, result.Steps, 1)
	require.Len(t, result.Ingredients, 1)
}

func TestParse_BlockComment(t *testing.T) {
	source := "Add @salt{1%tsp}. [- this is a block comment -] Stir well."
	result := Parse(source)
	require.Len(t, result.Steps, 1)
	require.Len(t, result.Ingredients, 1)
}

func TestParse_EmptySource(t *testing.T) {
	result := Parse("")
	assert.Empty(t, result.Steps)
	assert.Empty(t, result.Ingredients)
}

func TestParse_NoMarkers(t *testing.T) {
	result := Parse("Just mix everything together.")
	require.Len(t, result.Steps, 1)
	assert.Empty(t, result.Ingredients)
	require.Len(t, result.Steps[0].Segments, 1)
	assert.Equal(t, SegmentText, result.Steps[0].Segments[0].Type)
}

func TestParse_Carbonara(t *testing.T) {
	source := `Bring @water{2%l} to a boil in a #large pot{}.

Cook @spaghetti{400%g} until al dente, about ~{8%minutes}.

Meanwhile, dice @guanciale{200%g} and fry in a #skillet{} until crispy.

Whisk @eggs{3} with @pecorino romano{100%g} and @black pepper{1%tsp}.

Drain pasta, toss with guanciale, then mix in egg mixture off heat.`

	result := Parse(source)

	assert.Len(t, result.Steps, 5)
	require.Len(t, result.Ingredients, 6)

	assert.Equal(t, "water", result.Ingredients[0].Name)
	assert.Equal(t, "spaghetti", result.Ingredients[1].Name)
	assert.Equal(t, "guanciale", result.Ingredients[2].Name)
	assert.Equal(t, "eggs", result.Ingredients[3].Name)
	assert.Equal(t, "pecorino romano", result.Ingredients[4].Name)
	assert.Equal(t, "black pepper", result.Ingredients[5].Name)

	// Step 1 should have ingredient and cookware
	step1 := result.Steps[0]
	hasIngredient := false
	hasCookware := false
	for _, s := range step1.Segments {
		if s.Type == SegmentIngredient && s.Name == "water" {
			hasIngredient = true
		}
		if s.Type == SegmentCookware && s.Name == "large pot" {
			hasCookware = true
		}
	}
	assert.True(t, hasIngredient)
	assert.True(t, hasCookware)

	// Step 2 should have timer
	step2 := result.Steps[1]
	hasTimer := false
	for _, s := range step2.Segments {
		if s.Type == SegmentTimer {
			assert.Equal(t, "8", s.Quantity)
			assert.Equal(t, "minutes", s.Unit)
			hasTimer = true
		}
	}
	assert.True(t, hasTimer)
}

func TestParse_MetadataBlock(t *testing.T) {
	source := `---
source: https://example.com/pancakes
tags: fun, quick, breakfast
---

Crack the @eggs{3} into a blender.`

	result := Parse(source)

	// Metadata should be extracted
	assert.Equal(t, "https://example.com/pancakes", result.Metadata.Source)
	require.Len(t, result.Metadata.Tags, 3)
	assert.Equal(t, "fun", result.Metadata.Tags[0])
	assert.Equal(t, "quick", result.Metadata.Tags[1])
	assert.Equal(t, "breakfast", result.Metadata.Tags[2])

	// Metadata block should not appear as a step
	require.Len(t, result.Steps, 1)
	assert.Len(t, result.Ingredients, 1)
	assert.Equal(t, "eggs", result.Ingredients[0].Name)
}

func TestParse_MetadataBlockNotParsedAsStep(t *testing.T) {
	source := `---
source: https://www.jamieoliver.com/recipes/eggs-recipes/easy-pancakes/
tags: fun, quick
---

Crack the @eggs{3} into a blender, then add the @flour{125%g}.`

	result := Parse(source)

	// Should only have 1 step (the actual recipe), not the metadata
	require.Len(t, result.Steps, 1)
	assert.Equal(t, "fun", result.Metadata.Tags[0])
	assert.Equal(t, "quick", result.Metadata.Tags[1])
	assert.Equal(t, "https://www.jamieoliver.com/recipes/eggs-recipes/easy-pancakes/", result.Metadata.Source)
}

func TestParse_NoMetadata(t *testing.T) {
	result := Parse("Add @salt{1%tsp}.")
	assert.Empty(t, result.Metadata.Tags)
	assert.Empty(t, result.Metadata.Source)
}

func TestParse_SectionHeaders(t *testing.T) {
	source := `= Filling

Cook @rice{200%g}.

= Assembly

Stuff @bell peppers{4}.`

	result := Parse(source)
	require.Len(t, result.Steps, 2)
	assert.Equal(t, "Filling", result.Steps[0].Section)
	assert.Equal(t, "Assembly", result.Steps[1].Section)
}

func TestParse_Blockquotes(t *testing.T) {
	source := `> These freeze well. Double the batch.

Cook @rice{200%g}.`

	result := Parse(source)
	require.Len(t, result.Steps, 1)
	require.Len(t, result.Notes, 1)
	assert.Equal(t, "These freeze well. Double the batch.", result.Notes[0].Text)
	assert.Equal(t, 0, result.Notes[0].Position)
}

func TestParse_RecipeReference(t *testing.T) {
	source := "Serve with @./Sauces/Salsa Verde{}."
	result := Parse(source)
	require.Len(t, result.Steps, 1)

	// Should not be an ingredient
	assert.Empty(t, result.Ingredients)

	// Should have a reference segment
	found := false
	for _, s := range result.Steps[0].Segments {
		if s.Type == SegmentReference {
			assert.Equal(t, "Salsa Verde", s.Name)
			assert.Equal(t, "./Sauces/Salsa Verde", s.Path)
			found = true
		}
	}
	assert.True(t, found)
}

func TestParse_MetadataServingsAndTimes(t *testing.T) {
	source := `---
title: Stuffed Peppers
tags: [dinner, vegetarian]
servings: 4
prep time: 20 minutes
cook time: 35 minutes
---

Stuff @bell peppers{4}.`

	result := Parse(source)
	assert.Equal(t, "Stuffed Peppers", result.Metadata.Title)
	assert.Equal(t, "4", result.Metadata.Servings)
	assert.Equal(t, "20 minutes", result.Metadata.PrepTime)
	assert.Equal(t, "35 minutes", result.Metadata.CookTime)
	require.Len(t, result.Metadata.Tags, 2)
	assert.Equal(t, "dinner", result.Metadata.Tags[0])
	assert.Equal(t, "vegetarian", result.Metadata.Tags[1])
	require.Len(t, result.Steps, 1)
}

func TestParse_TagsBracketFormat(t *testing.T) {
	source := `---
tags: [dinner, vegetarian]
---

Cook @rice{200%g}.`

	result := Parse(source)
	require.Len(t, result.Metadata.Tags, 2)
	assert.Equal(t, "dinner", result.Metadata.Tags[0])
	assert.Equal(t, "vegetarian", result.Metadata.Tags[1])
}

func TestParse_ParenthesizedNotes(t *testing.T) {
	source := "Sauté @onion{1}(diced) and @garlic{3%cloves}(minced)."
	result := Parse(source)
	require.Len(t, result.Ingredients, 2)
	assert.Equal(t, "onion", result.Ingredients[0].Name)
	assert.Equal(t, "1", result.Ingredients[0].Quantity)
	assert.Equal(t, "garlic", result.Ingredients[1].Name)
	assert.Equal(t, "3", result.Ingredients[1].Quantity)
	assert.Equal(t, "cloves", result.Ingredients[1].Unit)
}

func TestParseMinutes(t *testing.T) {
	tests := []struct {
		input    string
		expected *int
	}{
		{"20 minutes", intPtr(20)},
		{"35 min", intPtr(35)},
		{"20", intPtr(20)},
		{"", nil},
		// Compound durations
		{"1h 20m", intPtr(80)},
		{"1h20m", intPtr(80)},
		{"1 hour 20 minutes", intPtr(80)},
		{"2h", intPtr(120)},
		{"45m", intPtr(45)},
		{"90 minutes", intPtr(90)},
		{"1.5 hours", intPtr(90)},
		{"1.5h", intPtr(90)},
		{"2 hours 30 minutes", intPtr(150)},
		{"1 hour", intPtr(60)},
		{"0.5 hours", intPtr(30)},
	}
	for _, tt := range tests {
		result := ParseMinutes(tt.input)
		if tt.expected == nil {
			assert.Nil(t, result, "input: %q", tt.input)
		} else {
			require.NotNil(t, result, "input: %q", tt.input)
			assert.Equal(t, *tt.expected, *result, "input: %q", tt.input)
		}
	}
}

func TestParse_MaxSourceSize(t *testing.T) {
	source := strings.Repeat("a", MaxSourceSize+1)
	result := Parse(source)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "maximum size")
}

func TestParse_StuffedPeppersFullRecipe(t *testing.T) {
	source := `---
title: Stuffed Peppers
tags: [dinner, vegetarian]
servings: 4
prep time: 20 minutes
cook time: 35 minutes
---

> These freeze well. Double the batch and freeze half for a quick weeknight dinner.

= Filling

Cook @rice{200%g} according to package directions.

Sauté @onion{1}(diced) and @garlic{3%cloves}(minced) in @olive oil{2%tbsp}
in a #large skillet{} until softened, about ~{5%minutes}.

Add @canned tomatoes{400%g}, @black beans{240%g}(drained),
@cumin{1%tsp}, and @smoked paprika{1%tsp}. Stir in the cooked rice.

= Assembly

Cut the tops off @bell peppers{4} and remove seeds.
Stuff with the filling and place in a #baking dish{}.

Top each pepper with @cheddar{100%g}(grated).

Bake in a preheated #oven{} at 190°C for ~{30%minutes} until peppers
are tender and cheese is bubbling.

Serve with @./Sauces/Salsa Verde{}.`

	result := Parse(source)

	// Metadata
	assert.Equal(t, "Stuffed Peppers", result.Metadata.Title)
	assert.Equal(t, "4", result.Metadata.Servings)
	assert.Equal(t, "20 minutes", result.Metadata.PrepTime)
	assert.Equal(t, "35 minutes", result.Metadata.CookTime)
	require.Len(t, result.Metadata.Tags, 2)
	require.Len(t, result.Notes, 1)

	// Steps — should NOT include metadata, blockquote, or section headers
	require.Len(t, result.Steps, 7)
	assert.Equal(t, "Filling", result.Steps[0].Section)
	assert.Equal(t, "Filling", result.Steps[1].Section)
	assert.Equal(t, "Filling", result.Steps[2].Section)
	assert.Equal(t, "Assembly", result.Steps[3].Section)
	assert.Equal(t, "Assembly", result.Steps[4].Section)
	assert.Equal(t, "Assembly", result.Steps[5].Section)
	assert.Equal(t, "Assembly", result.Steps[6].Section)

	// Ingredients — should not include recipe reference
	for _, ing := range result.Ingredients {
		assert.NotContains(t, ing.Name, "Salsa Verde", "Recipe reference should not be an ingredient")
		assert.NotContains(t, ing.Name, "./", "Path should not appear as ingredient name")
	}

	// Should have reference segment in last step
	lastStep := result.Steps[6]
	hasRef := false
	for _, s := range lastStep.Segments {
		if s.Type == SegmentReference {
			assert.Equal(t, "Salsa Verde", s.Name)
			hasRef = true
		}
	}
	assert.True(t, hasRef)
}

func intPtr(v int) *int { return &v }

func TestValidate_ValidSource(t *testing.T) {
	errs := Validate("Add @salt{1%tsp} and stir.")
	assert.Empty(t, errs)
}

func TestValidate_UnclosedBrace(t *testing.T) {
	errs := Validate("Add @salt{1%tsp and stir.")
	require.Len(t, errs, 1)
	assert.Equal(t, 1, errs[0].Line)
	assert.Contains(t, errs[0].Message, "Unclosed brace")
}

func TestValidate_EmptySource(t *testing.T) {
	errs := Validate("")
	assert.Empty(t, errs)
}

func TestParse_SectionVariants(t *testing.T) {
	cases := []string{
		"= Filling",
		"==Filling==",
		"== Filling",
		"== Filling ==",
		"=== Filling ===",
		"=  Filling  =",
	}
	for _, header := range cases {
		t.Run(header, func(t *testing.T) {
			source := header + "\n\nCook @rice{200%g}."
			result := Parse(source)
			require.Len(t, result.Steps, 1, "header: %q", header)
			assert.Equal(t, "Filling", result.Steps[0].Section, "header: %q", header)
			for _, seg := range result.Steps[0].Segments {
				assert.NotContains(t, seg.Value, "==")
			}
		})
	}
}

func TestParse_SectionInlineWithStep(t *testing.T) {
	source := "==Dough==\nMix @flour{500%g}"
	result := Parse(source)
	require.Len(t, result.Steps, 1)
	assert.Equal(t, "Dough", result.Steps[0].Section)
	for _, seg := range result.Steps[0].Segments {
		assert.NotContains(t, seg.Value, "==")
	}
}

func TestParse_BareSectionClearsActive(t *testing.T) {
	source := "= Filling\n\nCook @rice{200%g}.\n\n==\n\nServe."
	result := Parse(source)
	require.Len(t, result.Steps, 2)
	assert.Equal(t, "Filling", result.Steps[0].Section)
	assert.Equal(t, "", result.Steps[1].Section)
}

func TestParse_NotePositions(t *testing.T) {
	t.Run("before first step", func(t *testing.T) {
		source := "> First note\n\nStep one.\n\nStep two."
		result := Parse(source)
		require.Len(t, result.Steps, 2)
		require.Len(t, result.Notes, 1)
		assert.Equal(t, 0, result.Notes[0].Position)
		assert.Equal(t, "First note", result.Notes[0].Text)
	})

	t.Run("between steps", func(t *testing.T) {
		source := "Step one.\n\n> Middle note\n\nStep two."
		result := Parse(source)
		require.Len(t, result.Steps, 2)
		require.Len(t, result.Notes, 1)
		assert.Equal(t, 1, result.Notes[0].Position)
	})

	t.Run("after last step", func(t *testing.T) {
		source := "Step one.\n\nStep two.\n\n> Trailing note"
		result := Parse(source)
		require.Len(t, result.Steps, 2)
		require.Len(t, result.Notes, 1)
		assert.Equal(t, 2, result.Notes[0].Position)
	})

	t.Run("after section header before any step", func(t *testing.T) {
		source := "Step one.\n\n= Filling\n\n> Section note\n\nCook @rice{200%g}."
		result := Parse(source)
		require.Len(t, result.Steps, 2)
		require.Len(t, result.Notes, 1)
		assert.Equal(t, 1, result.Notes[0].Position)
		assert.Equal(t, "Filling", result.Steps[1].Section)
	})
}
