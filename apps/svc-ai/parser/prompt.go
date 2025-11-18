package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// SystemPrompt returns the system prompt for ingredient parsing
func SystemPrompt() string {
	return `You are an expert ingredient parser. Your task is to parse recipe ingredient lines into structured JSON format.

CRITICAL RULES:
1. Output ONLY valid JSON - no markdown, no explanations, no extra text
2. Use the exact JSON schema provided below
3. Extract quantity, unit, ingredient name, and preparation/notes
4. Normalize units to full names (tbsp→tablespoon, c→cup, oz→ounce, lb→pound, g→gram, ml→milliliter)
5. Assign confidence score (0.0-1.0) based on parsing certainty
6. If quantity is ambiguous (e.g., "a handful", "to taste"), set quantity to null and store in quantityRaw
7. Ingredient name should be the core ingredient without preparation or notes
8. Preparation: adjectives/verbs describing ingredient state (diced, minced, fresh, raw, cooked)
9. Notes: optional qualifiers (optional, to taste, or more, if needed)

JSON SCHEMA:
{
  "quantity": <number or null>,
  "quantityRaw": "<original quantity text>",
  "unit": "<normalized unit or null>",
  "unitRaw": "<original unit text or null>",
  "ingredient": "<core ingredient name>",
  "preparation": ["<prep1>", "<prep2>"],
  "notes": ["<note1>", "<note2>"],
  "confidence": <0.0 to 1.0>
}

EXAMPLES:

Input: "2 tbsp olive oil"
Output: {"quantity":2.0,"quantityRaw":"2","unit":"tablespoon","unitRaw":"tbsp","ingredient":"olive oil","preparation":[],"notes":[],"confidence":0.98}

Input: "4 boneless, skinless chicken breasts"
Output: {"quantity":4.0,"quantityRaw":"4","unit":"piece","unitRaw":null,"ingredient":"chicken breast","preparation":["boneless","skinless"],"notes":[],"confidence":0.95}

Input: "1 cup heavy cream"
Output: {"quantity":1.0,"quantityRaw":"1","unit":"cup","unitRaw":"cup","ingredient":"heavy cream","preparation":[],"notes":[],"confidence":0.99}

Input: "Salt and pepper to taste"
Output: {"quantity":null,"quantityRaw":"to taste","unit":null,"unitRaw":null,"ingredient":"salt and pepper","preparation":[],"notes":["to taste"],"confidence":0.85}

Input: "a handful of fresh basil leaves"
Output: {"quantity":null,"quantityRaw":"a handful","unit":null,"unitRaw":null,"ingredient":"basil leaves","preparation":["fresh"],"notes":["handful"],"confidence":0.75}

Input: "1-2 tsp vanilla extract"
Output: {"quantity":1.5,"quantityRaw":"1-2","unit":"teaspoon","unitRaw":"tsp","ingredient":"vanilla extract","preparation":[],"notes":[],"confidence":0.90}

Now parse the following ingredient line.`
}

// UserPrompt generates the user prompt for a specific ingredient line
func UserPrompt(line string, opts ParseOptions) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Ingredient line: %s\n\n", line))

	if opts.Locale != "" {
		b.WriteString(fmt.Sprintf("Locale: %s\n", opts.Locale))
	}

	if opts.Hints.MeasurementSystem != "" {
		b.WriteString(fmt.Sprintf("Measurement system: %s\n", opts.Hints.MeasurementSystem))
	}

	if len(opts.Hints.CommonUnits) > 0 {
		b.WriteString(fmt.Sprintf("Common units: %s\n", strings.Join(opts.Hints.CommonUnits, ", ")))
	}

	b.WriteString("\nOutput JSON only:")
	return b.String()
}

// RecipeExtractionPrompt returns the system prompt for extracting and parsing ingredients from full recipe text
func RecipeExtractionPrompt() string {
	return `You are an expert recipe analyzer. Your task is to extract ingredient lines from full recipe text and parse each into structured JSON format.

CRITICAL RULES:
1. Output ONLY valid JSON array - no markdown, no explanations, no extra text
2. Extract ONLY ingredient lines - ignore instructions, notes, metadata, cooking steps
3. Parse each ingredient line using the schema below
4. Normalize units to full names (tbsp→tablespoon, c→cup, oz→ounce, lb→pound, g→gram, ml→milliliter)
5. Assign confidence score (0.0-1.0) based on parsing certainty
6. If quantity is ambiguous (e.g., "a handful", "to taste"), set quantity to null and store in quantityRaw
7. Ingredient name should be the core ingredient without preparation or notes
8. Preparation: adjectives/verbs describing ingredient state (diced, minced, fresh, raw, cooked)
9. Notes: optional qualifiers (optional, to taste, or more, if needed)

JSON SCHEMA (output as array of these objects):
{
  "quantity": <number or null>,
  "quantityRaw": "<original quantity text>",
  "unit": "<normalized unit or null>",
  "unitRaw": "<original unit text or null>",
  "ingredient": "<core ingredient name>",
  "preparation": ["<prep1>", "<prep2>"],
  "notes": ["<note1>", "<note2>"],
  "confidence": <0.0 to 1.0>
}

EXAMPLE:

Input recipe text:
"""
Chocolate Chip Cookies

Ingredients:
- 2 cups all-purpose flour
- 1 tsp baking soda
- 1 cup butter, softened
- 3/4 cup sugar
- 2 eggs
- 2 cups chocolate chips

Instructions:
1. Preheat oven to 375°F
2. Mix flour and baking soda
3. Cream butter and sugar...
"""

Output:
[
  {"quantity":2.0,"quantityRaw":"2","unit":"cup","unitRaw":"cups","ingredient":"all-purpose flour","preparation":[],"notes":[],"confidence":0.99},
  {"quantity":1.0,"quantityRaw":"1","unit":"teaspoon","unitRaw":"tsp","ingredient":"baking soda","preparation":[],"notes":[],"confidence":0.98},
  {"quantity":1.0,"quantityRaw":"1","unit":"cup","unitRaw":"cup","ingredient":"butter","preparation":["softened"],"notes":[],"confidence":0.97},
  {"quantity":0.75,"quantityRaw":"3/4","unit":"cup","unitRaw":"cup","ingredient":"sugar","preparation":[],"notes":[],"confidence":0.98},
  {"quantity":2.0,"quantityRaw":"2","unit":"piece","unitRaw":null,"ingredient":"eggs","preparation":[],"notes":[],"confidence":0.95},
  {"quantity":2.0,"quantityRaw":"2","unit":"cup","unitRaw":"cups","ingredient":"chocolate chips","preparation":[],"notes":[],"confidence":0.99}
]

Now extract and parse ingredients from the following recipe text.`
}

// RecipeExtractionUserPrompt generates the user prompt for recipe extraction
func RecipeExtractionUserPrompt(recipeText string, opts ParseOptions) string {
	var b strings.Builder
	b.WriteString("Recipe text:\n\"\"\"\n")
	b.WriteString(recipeText)
	b.WriteString("\n\"\"\"\n\n")

	if opts.Locale != "" {
		b.WriteString(fmt.Sprintf("Locale: %s\n", opts.Locale))
	}

	if opts.Hints.MeasurementSystem != "" {
		b.WriteString(fmt.Sprintf("Measurement system: %s\n", opts.Hints.MeasurementSystem))
	}

	b.WriteString("\nOutput JSON array only:")
	return b.String()
}

// ParseLLMResponse parses the LLM's JSON response into a ParsedIngredient
func ParseLLMResponse(responseText string) (*ParsedIngredient, error) {
	// Trim whitespace and potential markdown code blocks
	cleaned := strings.TrimSpace(responseText)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var parsed ParsedIngredient
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w", err)
	}

	// Validate required fields
	if parsed.Ingredient == "" {
		return nil, fmt.Errorf("ingredient field is required but was empty")
	}

	// Validate confidence is in range
	if parsed.Confidence < 0.0 || parsed.Confidence > 1.0 {
		return nil, fmt.Errorf("confidence must be between 0.0 and 1.0, got %.2f", parsed.Confidence)
	}

	// Initialize empty slices if nil (for consistent JSON serialization)
	if parsed.Preparation == nil {
		parsed.Preparation = []string{}
	}
	if parsed.Notes == nil {
		parsed.Notes = []string{}
	}

	return &parsed, nil
}

// ParseRecipeExtractionResponse parses the LLM's JSON array response into ParsedIngredients
func ParseRecipeExtractionResponse(responseText string, logger logrus.FieldLogger) ([]ParsedIngredient, error) {
	// Trim whitespace and potential markdown code blocks
	cleaned := strings.TrimSpace(responseText)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	// Extract JSON array by tracking bracket depth
	// This handles cases where LLM adds extra text after the JSON
	startIdx := strings.Index(cleaned, "[")
	if startIdx == -1 {
		return nil, fmt.Errorf("no JSON array found in response (no opening bracket)")
	}

	// Track bracket depth to find matching closing bracket
	depth := 0
	endIdx := -1
	inString := false
	escape := false

	for i := startIdx; i < len(cleaned); i++ {
		char := cleaned[i]

		// Handle string literals (to ignore brackets inside strings)
		if char == '"' && !escape {
			inString = !inString
		}

		// Handle escape sequences
		if char == '\\' && !escape {
			escape = true
			continue
		} else {
			escape = false
		}

		// Track bracket depth (only outside strings)
		if !inString {
			if char == '[' {
				depth++
			} else if char == ']' {
				depth--
				if depth == 0 {
					endIdx = i
					break
				}
			}
		}
	}

	if endIdx == -1 {
		return nil, fmt.Errorf("no matching closing bracket found for JSON array")
	}

	// Extract just the JSON array
	jsonArrayStr := cleaned[startIdx : endIdx+1]

	// Log extracted JSON for debugging
	jsonSample := jsonArrayStr
	if len(jsonSample) > 300 {
		jsonSample = jsonSample[:300] + "..."
	}
	logger.WithFields(logrus.Fields{
		"extracted_length": len(jsonArrayStr),
		"json_sample":      jsonSample,
	}).Debug("Extracted JSON array from LLM response")

	var parsed []ParsedIngredient
	if err := json.Unmarshal([]byte(jsonArrayStr), &parsed); err != nil {
		// Return error with sample of the response for debugging
		sample := jsonArrayStr
		if len(sample) > 200 {
			sample = sample[:200] + "..."
		}
		return nil, fmt.Errorf("failed to parse LLM response as JSON array: %w (response started with: %s)", err, sample)
	}

	// Validate and normalize each ingredient
	for i := range parsed {
		if parsed[i].Ingredient == "" {
			return nil, fmt.Errorf("ingredient %d: ingredient field is required but was empty", i)
		}

		if parsed[i].Confidence < 0.0 || parsed[i].Confidence > 1.0 {
			return nil, fmt.Errorf("ingredient %d: confidence must be between 0.0 and 1.0, got %.2f", i, parsed[i].Confidence)
		}

		// Initialize empty slices if nil
		if parsed[i].Preparation == nil {
			parsed[i].Preparation = []string{}
		}
		if parsed[i].Notes == nil {
			parsed[i].Notes = []string{}
		}
	}

	return parsed, nil
}
