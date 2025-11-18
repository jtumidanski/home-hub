package parser

// ParseResult represents the structured result of parsing an ingredient line
type ParseResult struct {
	// The original ingredient line
	Line string `json:"line"`

	// Parsed components
	Parsed ParsedIngredient `json:"parsed"`

	// Provider information
	Provider ProviderInfo `json:"provider"`

	// Warnings (e.g., "low_confidence_quantity")
	Warnings []string `json:"warnings,omitempty"`
}

// ParsedIngredient contains the structured ingredient data
type ParsedIngredient struct {
	// Quantity as a float (nullable)
	Quantity *float64 `json:"quantity"`

	// Raw quantity string from the ingredient line (e.g., "1-2", "a handful")
	QuantityRaw string `json:"quantityRaw"`

	// Normalized unit (e.g., "tablespoon", "cup")
	Unit *string `json:"unit"`

	// Raw unit from the ingredient line (e.g., "tbsp", "c")
	UnitRaw *string `json:"unitRaw"`

	// Core ingredient name (e.g., "olive oil", "chicken breast")
	Ingredient string `json:"ingredient"`

	// Preparation steps (e.g., ["diced", "fresh"])
	Preparation []string `json:"preparation"`

	// Additional notes (e.g., ["optional", "to taste"])
	Notes []string `json:"notes"`

	// Confidence score (0.0 to 1.0)
	Confidence float64 `json:"confidence"`
}

// ProviderInfo contains metadata about the AI provider used
type ProviderInfo struct {
	// Provider name (e.g., "local_ollama", "cloud_free_tier")
	Name string `json:"name"`

	// Model used (e.g., "llama3", "gpt-4o-mini")
	Model string `json:"model"`

	// Latency in milliseconds
	LatencyMs int64 `json:"latencyMs"`
}

// ParseOptions contains optional parameters for parsing
type ParseOptions struct {
	// Locale for parsing (e.g., "en-US", "en-GB")
	Locale string

	// Hints to help the parser
	Hints ParseHints
}

// ParseHints provides contextual hints to improve parsing accuracy
type ParseHints struct {
	// Common units in the user's region (e.g., ["tsp", "tbsp", "cup", "g", "ml"])
	CommonUnits []string

	// Measurement system ("imperial" or "metric")
	MeasurementSystem string
}

// BatchParseResult contains results for batch parsing
type BatchParseResult struct {
	// Individual parse results for each line
	Results []ParseResult `json:"results"`
}
