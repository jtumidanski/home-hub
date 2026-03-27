package normalization

type UnitIdentity struct {
	Canonical string
	Family    string
}

var unitRegistry = map[string]UnitIdentity{
	// Count family
	"each":   {Canonical: "each", Family: "count"},
	"piece":  {Canonical: "piece", Family: "count"},
	"pcs":    {Canonical: "piece", Family: "count"},
	"count":  {Canonical: "each", Family: "count"},
	"whole":  {Canonical: "whole", Family: "count"},
	"clove":  {Canonical: "clove", Family: "count"},
	"cloves": {Canonical: "clove", Family: "count"},
	"head":   {Canonical: "head", Family: "count"},
	"heads":  {Canonical: "head", Family: "count"},
	"bunch":  {Canonical: "bunch", Family: "count"},
	"bunches": {Canonical: "bunch", Family: "count"},
	"sprig":  {Canonical: "sprig", Family: "count"},
	"sprigs": {Canonical: "sprig", Family: "count"},
	"stalk":  {Canonical: "stalk", Family: "count"},
	"stalks": {Canonical: "stalk", Family: "count"},
	"slice":  {Canonical: "slice", Family: "count"},
	"slices": {Canonical: "slice", Family: "count"},
	"pinch":  {Canonical: "pinch", Family: "count"},
	"pinches": {Canonical: "pinch", Family: "count"},
	"dash":   {Canonical: "dash", Family: "count"},
	"dashes": {Canonical: "dash", Family: "count"},

	// Weight family
	"g":          {Canonical: "gram", Family: "weight"},
	"gram":       {Canonical: "gram", Family: "weight"},
	"grams":      {Canonical: "gram", Family: "weight"},
	"kg":         {Canonical: "kilogram", Family: "weight"},
	"kilogram":   {Canonical: "kilogram", Family: "weight"},
	"kilograms":  {Canonical: "kilogram", Family: "weight"},
	"oz":         {Canonical: "ounce", Family: "weight"},
	"ounce":      {Canonical: "ounce", Family: "weight"},
	"ounces":     {Canonical: "ounce", Family: "weight"},
	"lb":         {Canonical: "pound", Family: "weight"},
	"pound":      {Canonical: "pound", Family: "weight"},
	"pounds":     {Canonical: "pound", Family: "weight"},

	// Volume family
	"ml":              {Canonical: "milliliter", Family: "volume"},
	"milliliter":      {Canonical: "milliliter", Family: "volume"},
	"milliliters":     {Canonical: "milliliter", Family: "volume"},
	"l":               {Canonical: "liter", Family: "volume"},
	"liter":           {Canonical: "liter", Family: "volume"},
	"liters":          {Canonical: "liter", Family: "volume"},
	"tsp":             {Canonical: "teaspoon", Family: "volume"},
	"teaspoon":        {Canonical: "teaspoon", Family: "volume"},
	"teaspoons":       {Canonical: "teaspoon", Family: "volume"},
	"tbsp":            {Canonical: "tablespoon", Family: "volume"},
	"tablespoon":      {Canonical: "tablespoon", Family: "volume"},
	"tablespoons":     {Canonical: "tablespoon", Family: "volume"},
	"cup":             {Canonical: "cup", Family: "volume"},
	"cups":            {Canonical: "cup", Family: "volume"},
	"fl oz":           {Canonical: "fluid ounce", Family: "volume"},
	"fluid ounce":     {Canonical: "fluid ounce", Family: "volume"},
	"fluid ounces":    {Canonical: "fluid ounce", Family: "volume"},
}

func LookupUnit(raw string) (UnitIdentity, bool) {
	identity, ok := unitRegistry[raw]
	return identity, ok
}
