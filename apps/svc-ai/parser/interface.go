package parser

import "context"

// IngredientParser defines the interface for parsing ingredient lines
type IngredientParser interface {
	// Parse parses a single ingredient line into structured data
	Parse(ctx context.Context, line string, opts ParseOptions) (ParseResult, error)

	// ParseBatch parses multiple ingredient lines in a single request
	ParseBatch(ctx context.Context, lines []string, opts ParseOptions) ([]ParseResult, error)

	// ParseRecipe extracts ingredients from full recipe text and parses them
	ParseRecipe(ctx context.Context, recipeText string, opts ParseOptions) ([]ParseResult, error)

	// Name returns the name of this parser implementation
	Name() string

	// HealthCheck verifies the parser is available and responsive
	HealthCheck(ctx context.Context) error
}
