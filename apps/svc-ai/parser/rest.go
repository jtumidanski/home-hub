package parser

import (
	"github.com/google/uuid"
)

// ParsedIngredientRestModel represents the JSON:API representation of a parsed ingredient
type ParsedIngredientRestModel struct {
	Id          uuid.UUID `json:"-"`
	Line        string    `json:"line"`
	Quantity    *float64  `json:"quantity"`
	QuantityRaw string    `json:"quantityRaw"`
	Unit        *string   `json:"unit"`
	UnitRaw     *string   `json:"unitRaw"`
	Ingredient  string    `json:"ingredient"`
	Preparation []string  `json:"preparation"`
	Notes       []string  `json:"notes"`
	Confidence  float64   `json:"confidence"`
	Provider    struct {
		Name      string `json:"name"`
		Model     string `json:"model"`
		LatencyMs int64  `json:"latencyMs"`
	} `json:"provider"`
	Warnings []string `json:"warnings,omitempty"`
}

// GetName returns the resource type name for JSON:API
func (r ParsedIngredientRestModel) GetName() string {
	return "parsed-ingredients"
}

// GetID returns the resource ID for JSON:API
func (r ParsedIngredientRestModel) GetID() string {
	if r.Id == uuid.Nil {
		return ""
	}
	return r.Id.String()
}

// SetID sets the resource ID from JSON:API
func (r *ParsedIngredientRestModel) SetID(idStr string) error {
	if idStr == "" {
		r.Id = uuid.Nil
		return nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	r.Id = id
	return nil
}

// TransformParseResult converts a ParseResult to a REST model
func TransformParseResult(result ParseResult) (ParsedIngredientRestModel, error) {
	// Generate a deterministic ID based on the line content (for API consistency)
	id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(result.Line))

	rest := ParsedIngredientRestModel{
		Id:          id,
		Line:        result.Line,
		Quantity:    result.Parsed.Quantity,
		QuantityRaw: result.Parsed.QuantityRaw,
		Unit:        result.Parsed.Unit,
		UnitRaw:     result.Parsed.UnitRaw,
		Ingredient:  result.Parsed.Ingredient,
		Preparation: result.Parsed.Preparation,
		Notes:       result.Parsed.Notes,
		Confidence:  result.Parsed.Confidence,
		Warnings:    result.Warnings,
	}

	rest.Provider.Name = result.Provider.Name
	rest.Provider.Model = result.Provider.Model
	rest.Provider.LatencyMs = result.Provider.LatencyMs

	return rest, nil
}

// TransformParseResultSlice converts a slice of ParseResults to REST models
func TransformParseResultSlice(results []ParseResult) ([]ParsedIngredientRestModel, error) {
	restModels := make([]ParsedIngredientRestModel, len(results))
	for i, result := range results {
		restModel, err := TransformParseResult(result)
		if err != nil {
			return nil, err
		}
		restModels[i] = restModel
	}
	return restModels, nil
}
