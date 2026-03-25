package cooklang

type SegmentType string

const (
	SegmentText       SegmentType = "text"
	SegmentIngredient SegmentType = "ingredient"
	SegmentCookware   SegmentType = "cookware"
	SegmentTimer      SegmentType = "timer"
	SegmentReference  SegmentType = "reference"
)

type Segment struct {
	Type     SegmentType `json:"type"`
	Value    string      `json:"value,omitempty"`
	Name     string      `json:"name,omitempty"`
	Quantity string      `json:"quantity,omitempty"`
	Unit     string      `json:"unit,omitempty"`
	Path     string      `json:"path,omitempty"`
}

type Step struct {
	Number   int       `json:"number"`
	Section  string    `json:"section,omitempty"`
	Segments []Segment `json:"segments"`
}

type Ingredient struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
	Unit     string `json:"unit"`
}

type Metadata struct {
	Tags            []string          `json:"tags,omitempty"`
	Source          string            `json:"source,omitempty"`
	Title           string            `json:"title,omitempty"`
	Servings        string            `json:"servings,omitempty"`
	PrepTime        string            `json:"prepTime,omitempty"`
	CookTime        string            `json:"cookTime,omitempty"`
	Notes           []string          `json:"notes,omitempty"`
	Extra           map[string]string `json:"extra,omitempty"`
}

type ParseResult struct {
	Ingredients []Ingredient `json:"ingredients"`
	Steps       []Step       `json:"steps"`
	Metadata    Metadata     `json:"metadata"`
	Errors      []ParseError `json:"errors,omitempty"`
}

type ParseError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

// MaxSourceSize is the maximum allowed source text size in bytes (64KB).
const MaxSourceSize = 64 * 1024
