package recipe

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planner"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe/cooklang"
)

// RestModel is the JSON:API list representation (no source/parsed data).
type RestModel struct {
	Id                  uuid.UUID `json:"-"`
	Title               string    `json:"title"`
	Description         string    `json:"description,omitempty"`
	Servings            *int      `json:"servings,omitempty"`
	PrepTimeMinutes     *int      `json:"prepTimeMinutes,omitempty"`
	CookTimeMinutes     *int      `json:"cookTimeMinutes,omitempty"`
	Tags                []string  `json:"tags"`
	PlannerReady        bool      `json:"plannerReady"`
	Classification      string    `json:"classification,omitempty"`
	ResolvedIngredients int       `json:"resolvedIngredients"`
	TotalIngredients    int       `json:"totalIngredients"`
	LastUsedDate        *string   `json:"lastUsedDate,omitempty"`
	UsageCount          int64     `json:"usageCount,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

func (r RestModel) GetName() string       { return "recipes" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

// RestPlannerConfigModel is the JSON representation for planner config.
type RestPlannerConfigModel struct {
	Classification     string `json:"classification,omitempty"`
	ServingsYield      *int   `json:"servingsYield,omitempty"`
	EatWithinDays      *int   `json:"eatWithinDays,omitempty"`
	MinGapDays         *int   `json:"minGapDays,omitempty"`
	MaxConsecutiveDays *int   `json:"maxConsecutiveDays,omitempty"`
}

// RestDetailModel is the JSON:API detail representation with source and parsed data.
type RestDetailModel struct {
	Id              uuid.UUID                          `json:"-"`
	Title           string                             `json:"title"`
	Description     string                             `json:"description,omitempty"`
	Servings        *int                               `json:"servings,omitempty"`
	PrepTimeMinutes *int                               `json:"prepTimeMinutes,omitempty"`
	CookTimeMinutes *int                               `json:"cookTimeMinutes,omitempty"`
	SourceURL       string                             `json:"sourceUrl,omitempty"`
	Tags            []string                           `json:"tags"`
	Source          string                             `json:"source"`
	Ingredients     []normalization.RestIngredientModel `json:"ingredients"`
	Steps           []cooklang.Step                    `json:"steps"`
	PlannerConfig   *RestPlannerConfigModel            `json:"plannerConfig,omitempty"`
	PlannerReady    bool                               `json:"plannerReady"`
	PlannerIssues   []string                           `json:"plannerIssues"`
	CreatedAt       time.Time                          `json:"createdAt"`
	UpdatedAt       time.Time                          `json:"updatedAt"`
}

func (r RestDetailModel) GetName() string       { return "recipes" }
func (r RestDetailModel) GetID() string          { return r.Id.String() }
func (r *RestDetailModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

type ListEnrichment struct {
	PlannerReady        bool
	Classification      string
	ResolvedIngredients int
	TotalIngredients    int
	LastUsedDate        *string
	UsageCount          int64
}

func TransformList(m Model, enrichment ListEnrichment) RestModel {
	tags := m.Tags()
	if tags == nil {
		tags = []string{}
	}
	return RestModel{
		Id: m.Id(), Title: m.Title(), Description: m.Description(),
		Servings: m.Servings(), PrepTimeMinutes: m.PrepTimeMinutes(), CookTimeMinutes: m.CookTimeMinutes(),
		Tags: tags,
		PlannerReady: enrichment.PlannerReady, Classification: enrichment.Classification,
		ResolvedIngredients: enrichment.ResolvedIngredients, TotalIngredients: enrichment.TotalIngredients,
		LastUsedDate: enrichment.LastUsedDate, UsageCount: enrichment.UsageCount,
		CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt(),
	}
}

func TransformListSlice(models []Model, enrichments []ListEnrichment) []RestModel {
	result := make([]RestModel, len(models))
	for i, m := range models {
		result[i] = TransformList(m, enrichments[i])
	}
	return result
}

type DetailEnrichment struct {
	Ingredients []normalization.RestIngredientModel
	PlannerConfig *RestPlannerConfigModel
	Readiness     planner.Readiness
}

func TransformDetail(m Model, parsed cooklang.ParseResult, enrichment DetailEnrichment) RestDetailModel {
	tags := m.Tags()
	if tags == nil {
		tags = []string{}
	}
	ingredients := enrichment.Ingredients
	if ingredients == nil {
		ingredients = []normalization.RestIngredientModel{}
	}
	steps := parsed.Steps
	if steps == nil {
		steps = []cooklang.Step{}
	}
	return RestDetailModel{
		Id: m.Id(), Title: m.Title(), Description: m.Description(),
		Servings: m.Servings(), PrepTimeMinutes: m.PrepTimeMinutes(), CookTimeMinutes: m.CookTimeMinutes(),
		SourceURL: m.SourceURL(), Tags: tags, Source: m.Source(),
		Ingredients: ingredients, Steps: steps,
		PlannerConfig: enrichment.PlannerConfig, PlannerReady: enrichment.Readiness.Ready, PlannerIssues: enrichment.Readiness.Issues,
		CreatedAt: m.CreatedAt(), UpdatedAt: m.UpdatedAt(),
	}
}

// RestTagModel is the JSON:API representation for a tag with count.
type RestTagModel struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

func (r RestTagModel) GetName() string       { return "recipe-tags" }
func (r RestTagModel) GetID() string          { return r.Tag }
func (r *RestTagModel) SetID(id string) error { r.Tag = id; return nil }

// RestParseModel is the JSON:API representation for a parse result.
type RestParseModel struct {
	Ingredients   []cooklang.Ingredient             `json:"ingredients"`
	Steps         []cooklang.Step                   `json:"steps"`
	Metadata      cooklang.Metadata                 `json:"metadata"`
	Errors        []cooklang.ParseError             `json:"errors,omitempty"`
	Normalization []normalization.PreviewResult      `json:"normalization,omitempty"`
}

func (r RestParseModel) GetName() string       { return "recipe-parse" }
func (r RestParseModel) GetID() string          { return "parse" }
func (r *RestParseModel) SetID(_ string) error  { return nil }

// CreateRequest is the JSON:API request body for creating a recipe.
type CreateRequest struct {
	Id              uuid.UUID               `json:"-"`
	Title           string                  `json:"title"`
	Description     string                  `json:"description"`
	Source          string                  `json:"source"`
	Servings        *int                    `json:"servings,omitempty"`
	PrepTimeMinutes *int                    `json:"prepTimeMinutes,omitempty"`
	CookTimeMinutes *int                    `json:"cookTimeMinutes,omitempty"`
	SourceURL       string                  `json:"sourceUrl"`
	Tags            []string                `json:"tags"`
	PlannerConfig   *RestPlannerConfigModel `json:"plannerConfig,omitempty"`
}

func (r CreateRequest) GetName() string       { return "recipes" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// UpdateRequest is the JSON:API request body for updating a recipe.
type UpdateRequest struct {
	Id              uuid.UUID               `json:"-"`
	Title           string                  `json:"title,omitempty"`
	Description     string                  `json:"description,omitempty"`
	Source          string                  `json:"source,omitempty"`
	Servings        *int                    `json:"servings,omitempty"`
	PrepTimeMinutes *int                    `json:"prepTimeMinutes,omitempty"`
	CookTimeMinutes *int                    `json:"cookTimeMinutes,omitempty"`
	SourceURL       string                  `json:"sourceUrl,omitempty"`
	Tags            []string                `json:"tags,omitempty"`
	PlannerConfig   *RestPlannerConfigModel `json:"plannerConfig,omitempty"`
}

func (r UpdateRequest) GetName() string       { return "recipes" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

// ParseRequest is the JSON:API request body for parsing Cooklang source.
type ParseRequest struct {
	Id     uuid.UUID `json:"-"`
	Source string    `json:"source"`
}

func (r ParseRequest) GetName() string       { return "recipe-parse" }
func (r ParseRequest) GetID() string          { return r.Id.String() }
func (r *ParseRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// RestorationRequest is the JSON:API request body for restoring a recipe.
type RestorationRequest struct {
	Id       uuid.UUID `json:"-"`
	RecipeId string    `json:"recipeId"`
}

func (r RestorationRequest) GetName() string       { return "recipe-restorations" }
func (r RestorationRequest) GetID() string          { return r.Id.String() }
func (r *RestorationRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
