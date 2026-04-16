package cooklang

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Parse parses Cooklang source text into structured ingredients and steps.
func Parse(source string) ParseResult {
	if len(source) > MaxSourceSize {
		return ParseResult{
			Errors: []ParseError{{Line: 1, Column: 1, Message: fmt.Sprintf("Source exceeds maximum size of %d bytes", MaxSourceSize)}},
		}
	}

	metadata, body := extractMetadata(source)
	body = stripComments(body)
	body, notes := stripNotesPositional(body)
	blocks := splitSteps(body)

	var steps []Step
	ingredientMap := make(map[string]*Ingredient)
	var ingredientOrder []string
	currentSection := ""

	for _, block := range blocks {
		text := strings.TrimSpace(block)
		if text == "" {
			continue
		}

		if name, ok := matchSection(text); ok {
			currentSection = name
			continue
		}

		segments, blockIngredients := parseBlock(text)
		steps = append(steps, Step{
			Number:   len(steps) + 1,
			Section:  currentSection,
			Segments: segments,
		})
		for _, ing := range blockIngredients {
			key := strings.ToLower(ing.Name)
			if existing, ok := ingredientMap[key]; ok {
				existing.Quantity = combineQuantities(existing.Quantity, existing.Unit, ing.Quantity, ing.Unit)
				if existing.Unit == "" && ing.Unit != "" {
					existing.Unit = ing.Unit
				}
			} else {
				copy := ing
				ingredientMap[key] = &copy
				ingredientOrder = append(ingredientOrder, key)
			}
		}
	}

	ingredients := make([]Ingredient, 0, len(ingredientOrder))
	for _, key := range ingredientOrder {
		ingredients = append(ingredients, *ingredientMap[key])
	}

	return ParseResult{
		Ingredients: ingredients,
		Steps:       steps,
		Metadata:    metadata,
		Notes:       notes,
	}
}

var sectionRe = regexp.MustCompile(`^=+\s*(.*?)\s*=*\s*$`)

// matchSection returns the section name (possibly empty to clear the active
// section) if the line is a section header. Recognizes any number of leading
// `=` characters with optional trailing `=` characters: `= Foo`, `==Foo==`,
// `=== Foo ===`, `==`, etc.
func matchSection(trimmed string) (string, bool) {
	if !strings.HasPrefix(trimmed, "=") {
		return "", false
	}
	m := sectionRe.FindStringSubmatch(trimmed)
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}

// extractMetadata strips metadata blocks (--- ... ---) from the source
// and parses key: value pairs within them. Returns metadata and remaining body.
func extractMetadata(source string) (Metadata, string) {
	meta := Metadata{Extra: make(map[string]string)}
	remaining := source

	for {
		start := strings.Index(remaining, "---")
		if start < 0 {
			break
		}
		// Only match if --- is at the start of a line (possibly with leading whitespace)
		prefix := remaining[:start]
		if prefix != "" {
			lastNewline := strings.LastIndex(prefix, "\n")
			linePrefix := prefix[lastNewline+1:]
			if strings.TrimSpace(linePrefix) != "" {
				break
			}
		}

		afterStart := remaining[start+3:]
		end := strings.Index(afterStart, "---")
		if end < 0 {
			break
		}

		block := afterStart[:end]
		lines := strings.Split(block, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			colonIdx := strings.Index(line, ":")
			if colonIdx < 0 {
				continue
			}
			key := strings.TrimSpace(strings.ToLower(line[:colonIdx]))
			value := strings.TrimSpace(line[colonIdx+1:])

			switch key {
			case "tags":
				// Handle both "tag1, tag2" and "[tag1, tag2]" formats
				value = strings.TrimPrefix(value, "[")
				value = strings.TrimSuffix(value, "]")
				parts := strings.Split(value, ",")
				for _, p := range parts {
					t := strings.TrimSpace(strings.ToLower(p))
					if t != "" {
						meta.Tags = append(meta.Tags, t)
					}
				}
			case "source":
				meta.Source = value
			case "title":
				meta.Title = value
			case "description":
				meta.Description = value
			case "servings":
				meta.Servings = value
			case "prep time":
				meta.PrepTime = value
			case "cook time":
				meta.CookTime = value
			default:
				meta.Extra[key] = value
			}
		}

		// Remove the metadata block from the source
		remaining = remaining[:start] + afterStart[end+3:]
	}

	return meta, remaining
}

// stripNotesPositional removes blockquote lines (> ...) from the source and
// returns them as PositionalNote values whose Position is the index of the
// step block that follows the note (mirrors splitSteps block counting:
// section-header lines and blank lines do not advance the index).
func stripNotesPositional(source string) (string, []PositionalNote) {
	lines := strings.Split(source, "\n")
	var resultLines []string
	var notes []PositionalNote
	blockIndex := 0
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ">") {
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			if text != "" {
				notes = append(notes, PositionalNote{Position: blockIndex, Text: text})
			}
			continue
		}
		if trimmed == "" {
			if inBlock {
				blockIndex++
				inBlock = false
			}
			resultLines = append(resultLines, line)
			continue
		}
		if _, isSection := matchSection(trimmed); isSection {
			if inBlock {
				blockIndex++
				inBlock = false
			}
			resultLines = append(resultLines, line)
			continue
		}
		inBlock = true
		resultLines = append(resultLines, line)
	}
	return strings.Join(resultLines, "\n"), notes
}

// Validate checks Cooklang source for syntax errors without fully parsing.
func Validate(source string) []ParseError {
	if len(source) > MaxSourceSize {
		return []ParseError{{Line: 1, Column: 1, Message: fmt.Sprintf("Source exceeds maximum size of %d bytes", MaxSourceSize)}}
	}

	var errs []ParseError
	// Strip metadata blocks before validating
	_, body := extractMetadata(source)
	lines := strings.Split(body, "\n")
	inBlockComment := false

	for lineNum, line := range lines {
		if inBlockComment {
			if idx := strings.Index(line, "-]"); idx >= 0 {
				inBlockComment = false
				line = line[idx+2:]
			} else {
				continue
			}
		}

		if idx := strings.Index(line, "[-"); idx >= 0 {
			if end := strings.Index(line[idx:], "-]"); end < 0 {
				inBlockComment = true
			}
		}

		// Skip notes and section headers
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ">") {
			continue
		}
		if _, isSection := matchSection(trimmed); isSection {
			continue
		}

		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}

		for i := 0; i < len(line); i++ {
			ch := line[i]
			if ch == '@' || ch == '#' || ch == '~' {
				braceStart := -1
				nameStart := i + 1

				for j := nameStart; j < len(line); j++ {
					if line[j] == '{' {
						braceStart = j
						break
					}
					if line[j] == ' ' || line[j] == '.' || line[j] == ',' {
						break
					}
				}

				if braceStart >= 0 {
					braceEnd := strings.IndexByte(line[braceStart:], '}')
					if braceEnd < 0 {
						errs = append(errs, ParseError{
							Line:    lineNum + 1,
							Column:  braceStart + 1,
							Message: fmt.Sprintf("Unclosed brace in %s block", markerName(ch)),
						})
					}
				}
			}
		}
	}

	return errs
}

func markerName(ch byte) string {
	switch ch {
	case '@':
		return "ingredient"
	case '#':
		return "cookware"
	case '~':
		return "timer"
	}
	return "unknown"
}

func stripComments(source string) string {
	// Remove block comments [- ... -]
	for {
		start := strings.Index(source, "[-")
		if start < 0 {
			break
		}
		end := strings.Index(source[start:], "-]")
		if end < 0 {
			source = source[:start]
			break
		}
		source = source[:start] + source[start+end+2:]
	}

	// Remove line comments --
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "--"); idx >= 0 {
			lines[i] = line[:idx]
		}
	}
	return strings.Join(lines, "\n")
}

func splitSteps(source string) []string {
	lines := strings.Split(source, "\n")
	var blocks []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			blocks = append(blocks, current.String())
			current.Reset()
		}
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flush()
			continue
		}
		if _, isSection := matchSection(trimmed); isSection {
			flush()
			blocks = append(blocks, trimmed)
			continue
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(trimmed)
	}
	flush()
	return blocks
}

func parseBlock(text string) ([]Segment, []Ingredient) {
	var segments []Segment
	var ingredients []Ingredient
	i := 0

	for i < len(text) {
		ch := text[i]
		if ch == '@' || ch == '#' || ch == '~' {
			name, qty, unit, path, advance := parseMarker(text[i:], ch)
			if advance > 0 {
				switch ch {
				case '@':
					if path != "" {
						// Recipe reference: @./Sauces/Salsa Verde{}
						segments = append(segments, Segment{
							Type: SegmentReference, Name: name, Path: path,
						})
					} else {
						segments = append(segments, Segment{
							Type: SegmentIngredient, Name: name, Quantity: qty, Unit: unit,
						})
						ingredients = append(ingredients, Ingredient{Name: name, Quantity: qty, Unit: unit})
					}
				case '#':
					segments = append(segments, Segment{
						Type: SegmentCookware, Name: name,
					})
				case '~':
					segments = append(segments, Segment{
						Type: SegmentTimer, Quantity: qty, Unit: unit,
					})
				}
				i += advance
				continue
			}
		}

		// Accumulate text
		start := i
		for i < len(text) && text[i] != '@' && text[i] != '#' && text[i] != '~' {
			i++
		}
		if i > start {
			segments = append(segments, Segment{Type: SegmentText, Value: text[start:i]})
		}
	}

	return segments, ingredients
}

// parseMarker parses @name{qty%unit}, #name{}, ~{qty%unit}, etc.
// Returns name, quantity, unit, path (for recipe references), and bytes consumed.
func parseMarker(text string, marker byte) (string, string, string, string, int) {
	if len(text) < 2 {
		return "", "", "", "", 0
	}

	pos := 1 // skip the marker character

	// Check for recipe reference: @./path/to/recipe{}
	if marker == '@' && pos < len(text) && text[pos] == '.' {
		braceIdx := strings.IndexByte(text[pos:], '{')
		if braceIdx >= 0 {
			braceIdx += pos
			braceEnd := strings.IndexByte(text[braceIdx:], '}')
			if braceEnd < 0 {
				return "", "", "", "", 0
			}
			path := text[pos : braceIdx]
			// Derive a display name from the path (last segment)
			name := path
			if lastSlash := strings.LastIndex(path, "/"); lastSlash >= 0 {
				name = path[lastSlash+1:]
			}
			name = strings.TrimPrefix(name, ".")
			name = strings.TrimPrefix(name, "/")
			return name, "", "", path, braceIdx + braceEnd + 1
		}
	}

	// Look ahead for a brace to determine if this is a multi-word name
	braceIdx := -1
	scanPos := pos
	for scanPos < len(text) {
		if text[scanPos] == '{' {
			braceIdx = scanPos
			break
		}
		if text[scanPos] == '@' || text[scanPos] == '#' || text[scanPos] == '~' {
			break
		}
		if text[scanPos] == '\n' {
			break
		}
		scanPos++
	}

	if braceIdx >= 0 {
		// Multi-word name with brace: @pecorino romano{100%g}
		name := text[pos:braceIdx]
		braceEnd := strings.IndexByte(text[braceIdx:], '}')
		if braceEnd < 0 {
			return "", "", "", "", 0
		}
		braceContent := text[braceIdx+1 : braceIdx+braceEnd]

		// Strip parenthesized notes after closing brace: @onion{1}(diced)
		consumed := braceIdx + braceEnd + 1
		if consumed < len(text) && text[consumed] == '(' {
			closeP := strings.IndexByte(text[consumed:], ')')
			if closeP >= 0 {
				consumed += closeP + 1
			}
		}

		qty, unit := parseQuantityUnit(braceContent)
		return strings.TrimSpace(name), qty, unit, "", consumed
	}

	// No brace — single word name (stop at space, punctuation, or non-word char)
	nameStart := pos
	for pos < len(text) {
		if !isSingleWordNameChar(rune(text[pos])) {
			break
		}
		pos++
	}

	name := text[nameStart:pos]
	if name == "" {
		return "", "", "", "", 0
	}
	return strings.TrimSpace(name), "", "", "", pos
}

func isSingleWordNameChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '\''
}

// knownUnits is a set of recognized unit abbreviations used to distinguish
// a unit word from a descriptor word (e.g., "tsp" is a unit, "small" is not).
var knownUnits = map[string]bool{
	"tsp": true, "teaspoon": true, "teaspoons": true,
	"tbsp": true, "tablespoon": true, "tablespoons": true,
	"cup": true, "cups": true,
	"oz": true, "ounce": true, "ounces": true,
	"lb": true, "pound": true, "pounds": true,
	"g": true, "gram": true, "grams": true,
	"kg": true, "kilogram": true, "kilograms": true,
	"ml": true, "milliliter": true, "milliliters": true,
	"l": true, "liter": true, "liters": true,
	"pinch": true, "pinches": true, "dash": true, "dashes": true,
	"clove": true, "cloves": true, "sprig": true, "sprigs": true,
	"head": true, "heads": true, "bunch": true, "bunches": true,
	"stalk": true, "stalks": true, "slice": true, "slices": true,
	"each": true, "piece": true, "pcs": true, "whole": true,
}

// qtyRe matches a leading numeric portion: integer, decimal, or fraction.
// Examples: "2", "1.5", "1/2", "2 1/4" (mixed number)
var qtyRe = regexp.MustCompile(`^(\d+(?:\.\d+)?(?:\s+\d+)?(?:/\d+)?)`)

// noSpaceUnitRe matches numbers directly followed by a unit: "8oz", "16oz"
var noSpaceUnitRe = regexp.MustCompile(`^(\d+(?:\.\d+)?)(` + strings.Join(unitKeys(), "|") + `)$`)

func unitKeys() []string {
	keys := make([]string, 0, len(knownUnits))
	for k := range knownUnits {
		keys = append(keys, k)
	}
	return keys
}

func parseQuantityUnit(content string) (string, string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ""
	}
	// Canonical Cooklang separator
	if idx := strings.Index(content, "%"); idx >= 0 {
		return strings.TrimSpace(content[:idx]), strings.TrimSpace(content[idx+1:])
	}
	// No-space unit: "8oz", "16oz"
	if m := noSpaceUnitRe.FindStringSubmatch(content); m != nil {
		return m[1], m[2]
	}
	// Extract leading quantity, then check if the next word is a known unit.
	if m := qtyRe.FindString(content); m != "" {
		rest := strings.TrimSpace(content[len(m):])
		if rest == "" {
			return m, ""
		}
		// Check if the next word is a known unit
		nextWord := rest
		if idx := strings.IndexAny(rest, " ,;("); idx >= 0 {
			nextWord = rest[:idx]
		}
		if knownUnits[strings.ToLower(nextWord)] {
			return strings.TrimSpace(m), strings.ToLower(nextWord)
		}
		// Not a known unit — return the full content as quantity (it's a description)
		return content, ""
	}
	return content, ""
}

func combineQuantities(existingQty, existingUnit, newQty, newUnit string) string {
	if existingQty == "" {
		return newQty
	}
	if newQty == "" {
		return existingQty
	}
	if existingUnit == newUnit {
		e := parseFloat(existingQty)
		n := parseFloat(newQty)
		if e > 0 && n > 0 {
			sum := e + n
			if sum == float64(int(sum)) {
				return fmt.Sprintf("%d", int(sum))
			}
			return fmt.Sprintf("%.1f", sum)
		}
	}
	return existingQty + " + " + newQty
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// ParseMinutes extracts minutes from duration strings.
// Supports compound formats like "1h 20m", "1h20m", "1 hour 20 minutes",
// simple formats like "2h", "45m", "90 minutes", "20",
// and fractional hours like "1.5 hours".
func ParseMinutes(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var totalMinutes float64
	matched := false

	// Match fractional or integer hours: "1.5 hours", "1.5h", "2 hours", "2h"
	hourRe := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*h(?:ours?)?`)
	for _, m := range hourRe.FindAllStringSubmatch(s, -1) {
		var h float64
		fmt.Sscanf(m[1], "%f", &h)
		totalMinutes += h * 60
		matched = true
	}

	// Match integer minutes: "20 minutes", "20m", "20 min"
	minRe := regexp.MustCompile(`(\d+)\s*m(?:in(?:utes?)?)?`)
	for _, m := range minRe.FindAllStringSubmatch(s, -1) {
		v, _ := strconv.Atoi(m[1])
		totalMinutes += float64(v)
		matched = true
	}

	if matched {
		result := int(math.Round(totalMinutes))
		return &result
	}

	// Fallback: bare number (e.g. "20")
	bareRe := regexp.MustCompile(`^(\d+)$`)
	if m := bareRe.FindStringSubmatch(s); m != nil {
		v, _ := strconv.Atoi(m[1])
		return &v
	}

	return nil
}

// ParseServings extracts a serving count from strings like "4", "4 servings".
func ParseServings(s string) *int {
	return ParseMinutes(s) // Same logic — extract first number
}
