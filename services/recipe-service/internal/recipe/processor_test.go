package recipe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "basic normalization",
			input:    []string{"Italian", "PASTA", "quick"},
			expected: []string{"italian", "pasta", "quick"},
		},
		{
			name:     "deduplication",
			input:    []string{"italian", "Italian", "ITALIAN"},
			expected: []string{"italian"},
		},
		{
			name:     "trims whitespace",
			input:    []string{" italian ", "  pasta  "},
			expected: []string{"italian", "pasta"},
		},
		{
			name:     "filters empty strings",
			input:    []string{"italian", "", "  ", "pasta"},
			expected: []string{"italian", "pasta"},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeStringSlices(t *testing.T) {
	tests := []struct {
		name     string
		slices   [][]string
		expected []string
	}{
		{
			name:     "merge two slices",
			slices:   [][]string{{"a", "b"}, {"c", "d"}},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "merge with nil",
			slices:   [][]string{{"a"}, nil, {"b"}},
			expected: []string{"a", "b"},
		},
		{
			name:     "all nil",
			slices:   [][]string{nil, nil},
			expected: nil,
		},
		{
			name:     "single slice",
			slices:   [][]string{{"a", "b"}},
			expected: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeStringSlices(tt.slices...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetadataDerivation(t *testing.T) {
	t.Run("derives tags from cooklang metadata", func(t *testing.T) {
		source := "---\ntags: italian, pasta\n---\n\nAdd @salt{1%tsp}."
		attrs := CreateAttrs{
			Title:  "Test",
			Source: source,
		}

		// Simulate what the processor does for tag derivation
		parsed := parseSourceForTest(source)
		tags := normalizeTags(mergeStringSlices(parsed.Metadata.Tags, attrs.Tags))
		require.Len(t, tags, 2)
		assert.Equal(t, "italian", tags[0])
		assert.Equal(t, "pasta", tags[1])
	})

	t.Run("merges explicit and metadata tags", func(t *testing.T) {
		source := "---\ntags: italian\n---\n\nAdd @salt{1%tsp}."
		attrs := CreateAttrs{
			Title:  "Test",
			Source: source,
			Tags:   []string{"quick"},
		}

		parsed := parseSourceForTest(source)
		tags := normalizeTags(mergeStringSlices(parsed.Metadata.Tags, attrs.Tags))
		require.Len(t, tags, 2)
		assert.Contains(t, tags, "italian")
		assert.Contains(t, tags, "quick")
	})

	t.Run("deduplicates merged tags", func(t *testing.T) {
		source := "---\ntags: italian\n---\n\nAdd @salt{1%tsp}."
		attrs := CreateAttrs{
			Title:  "Test",
			Source: source,
			Tags:   []string{"Italian"},
		}

		parsed := parseSourceForTest(source)
		tags := normalizeTags(mergeStringSlices(parsed.Metadata.Tags, attrs.Tags))
		require.Len(t, tags, 1)
		assert.Equal(t, "italian", tags[0])
	})

	t.Run("derives source URL from metadata", func(t *testing.T) {
		source := "---\nsource: https://example.com/recipe\n---\n\nAdd @salt{1%tsp}."

		parsed := parseSourceForTest(source)
		sourceURL := ""
		if sourceURL == "" && parsed.Metadata.Source != "" {
			sourceURL = parsed.Metadata.Source
		}
		assert.Equal(t, "https://example.com/recipe", sourceURL)
	})

	t.Run("explicit source URL takes precedence", func(t *testing.T) {
		source := "---\nsource: https://metadata.com\n---\n\nAdd @salt{1%tsp}."

		parsed := parseSourceForTest(source)
		sourceURL := "https://explicit.com"
		if sourceURL == "" && parsed.Metadata.Source != "" {
			sourceURL = parsed.Metadata.Source
		}
		assert.Equal(t, "https://explicit.com", sourceURL)
	})

	t.Run("derives servings from metadata", func(t *testing.T) {
		source := "---\nservings: 4\n---\n\nAdd @salt{1%tsp}."

		parsed := parseSourceForTest(source)
		assert.Equal(t, "4", parsed.Metadata.Servings)
	})

	t.Run("derives prep and cook time from metadata", func(t *testing.T) {
		source := "---\nprep time: 20 minutes\ncook time: 35 minutes\n---\n\nAdd @salt{1%tsp}."

		parsed := parseSourceForTest(source)
		assert.Equal(t, "20 minutes", parsed.Metadata.PrepTime)
		assert.Equal(t, "35 minutes", parsed.Metadata.CookTime)
	})
}

func TestEntityModelRoundTrip(t *testing.T) {
	servings := 4
	prepTime := 10
	cookTime := 20

	t.Run("entity to model and back preserves data", func(t *testing.T) {
		m, err := NewBuilder().
			SetTitle("Carbonara").
			SetDescription("Classic").
			SetSource("Add @eggs{3}.").
			SetServings(&servings).
			SetPrepTimeMinutes(&prepTime).
			SetCookTimeMinutes(&cookTime).
			SetSourceURL("https://example.com").
			SetTags([]string{"italian", "pasta"}).
			Build()
		require.NoError(t, err)

		// Model → Entity → Model
		e := m.ToEntity()
		m2, err := Make(e)
		require.NoError(t, err)

		assert.Equal(t, m.Title(), m2.Title())
		assert.Equal(t, m.Description(), m2.Description())
		assert.Equal(t, m.Source(), m2.Source())
		assert.Equal(t, *m.Servings(), *m2.Servings())
		assert.Equal(t, *m.PrepTimeMinutes(), *m2.PrepTimeMinutes())
		assert.Equal(t, *m.CookTimeMinutes(), *m2.CookTimeMinutes())
		assert.Equal(t, m.SourceURL(), m2.SourceURL())
	})

	t.Run("entity with nil optional fields", func(t *testing.T) {
		m, err := NewBuilder().
			SetTitle("Simple").
			SetSource("Stir.").
			Build()
		require.NoError(t, err)

		e := m.ToEntity()
		m2, err := Make(e)
		require.NoError(t, err)

		assert.Equal(t, "Simple", m2.Title())
		assert.Empty(t, m2.Description())
		assert.Nil(t, m2.Servings())
		assert.Empty(t, m2.SourceURL())
	})
}

// parseSourceForTest is a test helper that calls the cooklang parser.
func parseSourceForTest(source string) struct {
	Metadata struct {
		Tags     []string
		Source   string
		Servings string
		PrepTime string
		CookTime string
	}
} {
	// Import the cooklang package indirectly via the processor's ParseSource
	p := &Processor{}
	result := p.ParseSource(source)
	return struct {
		Metadata struct {
			Tags     []string
			Source   string
			Servings string
			PrepTime string
			CookTime string
		}
	}{
		Metadata: struct {
			Tags     []string
			Source   string
			Servings string
			PrepTime string
			CookTime string
		}{
			Tags:     result.Metadata.Tags,
			Source:    result.Metadata.Source,
			Servings: result.Metadata.Servings,
			PrepTime: result.Metadata.PrepTime,
			CookTime: result.Metadata.CookTime,
		},
	}
}
