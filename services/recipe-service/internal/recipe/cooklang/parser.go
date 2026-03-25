package cooklang

import (
	"fmt"
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
	body = stripNotes(body, &metadata)
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

		// Section header: = Filling
		if strings.HasPrefix(text, "= ") || text == "=" {
			currentSection = strings.TrimSpace(text[1:])
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
	}
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

// stripNotes removes blockquote lines (> ...) and collects them as notes in metadata.
func stripNotes(source string, meta *Metadata) string {
	lines := strings.Split(source, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ">") {
			note := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			if note != "" {
				meta.Notes = append(meta.Notes, note)
			}
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
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
		if strings.HasPrefix(trimmed, ">") || strings.HasPrefix(trimmed, "= ") {
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

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if current.Len() > 0 {
				blocks = append(blocks, current.String())
				current.Reset()
			}
		} else {
			if current.Len() > 0 {
				// Section headers get their own block
				if strings.HasPrefix(trimmed, "= ") || trimmed == "=" {
					blocks = append(blocks, current.String())
					current.Reset()
				} else {
					current.WriteString(" ")
				}
			}
			current.WriteString(trimmed)
		}
	}
	if current.Len() > 0 {
		blocks = append(blocks, current.String())
	}
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

func parseQuantityUnit(content string) (string, string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ""
	}
	if idx := strings.Index(content, "%"); idx >= 0 {
		return strings.TrimSpace(content[:idx]), strings.TrimSpace(content[idx+1:])
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

// ParseMinutes extracts minutes from strings like "20 minutes", "35 min", or "20".
func ParseMinutes(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	re := regexp.MustCompile(`(\d+)`)
	match := re.FindString(s)
	if match == "" {
		return nil
	}
	v, err := strconv.Atoi(match)
	if err != nil {
		return nil
	}
	return &v
}

// ParseServings extracts a serving count from strings like "4", "4 servings".
func ParseServings(s string) *int {
	return ParseMinutes(s) // Same logic — extract first number
}
