package recipe

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe/cooklang"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("recipe not found")
	ErrNotDeleted    = errors.New("recipe is not deleted")
	ErrRestoreWindow = errors.New("restore window expired")
)

const restoreWindowDays = 3

type CreateAttrs struct {
	Title           string
	Description     string
	Source          string
	Servings        *int
	PrepTimeMinutes *int
	CookTimeMinutes *int
	SourceURL       string
	Tags            []string
}

type UpdateAttrs struct {
	Title           *string
	Description     *string
	Source          *string
	Servings        *int
	PrepTimeMinutes *int
	CookTimeMinutes *int
	SourceURL       *string
	Tags            *[]string
}

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))
}

func (p *Processor) Create(tenantID, householdID uuid.UUID, attrs CreateAttrs) (Model, cooklang.ParseResult, error) {
	if _, err := NewBuilder().SetTitle(attrs.Title).SetSource(attrs.Source).Build(); err != nil {
		return Model{}, cooklang.ParseResult{}, err
	}

	parseErrors := cooklang.Validate(attrs.Source)
	if len(parseErrors) > 0 {
		return Model{}, cooklang.ParseResult{Errors: parseErrors}, errors.New("invalid cooklang syntax")
	}

	parsed := cooklang.Parse(attrs.Source)

	// Derive tags from Cooklang metadata, merging with any explicitly provided tags
	tags := normalizeTags(mergeStringSlices(parsed.Metadata.Tags, attrs.Tags))

	// Derive fields from Cooklang metadata if not explicitly provided
	sourceURL := attrs.SourceURL
	if sourceURL == "" && parsed.Metadata.Source != "" {
		sourceURL = parsed.Metadata.Source
	}
	servings := attrs.Servings
	if servings == nil && parsed.Metadata.Servings != "" {
		servings = cooklang.ParseServings(parsed.Metadata.Servings)
	}
	prepTime := attrs.PrepTimeMinutes
	if prepTime == nil && parsed.Metadata.PrepTime != "" {
		prepTime = cooklang.ParseMinutes(parsed.Metadata.PrepTime)
	}
	cookTime := attrs.CookTimeMinutes
	if cookTime == nil && parsed.Metadata.CookTime != "" {
		cookTime = cooklang.ParseMinutes(parsed.Metadata.CookTime)
	}
	title := attrs.Title
	if title == "" && parsed.Metadata.Title != "" {
		title = parsed.Metadata.Title
	}

	now := time.Now().UTC()
	id := uuid.New()

	tagEntities := make([]TagEntity, len(tags))
	for i, t := range tags {
		tagEntities[i] = TagEntity{Id: uuid.New(), RecipeId: id, Tag: t}
	}

	var desc *string
	if attrs.Description != "" {
		desc = &attrs.Description
	}
	var srcURL *string
	if sourceURL != "" {
		srcURL = &sourceURL
	}

	e := &Entity{
		Id: id, TenantId: tenantID, HouseholdId: householdID,
		Title: title, Description: desc, Source: attrs.Source,
		Servings: servings, PrepTimeMinutes: prepTime, CookTimeMinutes: cookTime,
		SourceURL: srcURL, Tags: tagEntities,
		CreatedAt: now, UpdatedAt: now,
	}

	if err := create(p.db.WithContext(p.ctx), e); err != nil {
		return Model{}, cooklang.ParseResult{}, err
	}

	m, err := Make(*e)
	if err != nil {
		return Model{}, cooklang.ParseResult{}, err
	}
	return m, parsed, nil
}

func (p *Processor) Get(id uuid.UUID) (Model, cooklang.ParseResult, error) {
	m, err := p.ByIDProvider(id)()
	if err != nil {
		return Model{}, cooklang.ParseResult{}, ErrNotFound
	}
	return m, cooklang.Parse(m.Source()), nil
}

func (p *Processor) List(search string, tags []string, page, pageSize int) ([]Model, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	entities, total, err := getAll(search, tags, page, pageSize)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, 0, err
	}

	models := make([]Model, 0, len(entities))
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, 0, err
		}
		models = append(models, m)
	}
	return models, total, nil
}

func (p *Processor) Update(id uuid.UUID, attrs UpdateAttrs) (Model, cooklang.ParseResult, error) {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, cooklang.ParseResult{}, ErrNotFound
	}

	if attrs.Title != nil {
		e.Title = *attrs.Title
	}
	if attrs.Description != nil {
		e.Description = attrs.Description
	}
	if attrs.Source != nil {
		parseErrors := cooklang.Validate(*attrs.Source)
		if len(parseErrors) > 0 {
			return Model{}, cooklang.ParseResult{Errors: parseErrors}, errors.New("invalid cooklang syntax")
		}
		e.Source = *attrs.Source

		// Re-derive tags and source URL from updated Cooklang metadata
		parsed := cooklang.Parse(*attrs.Source)
		if len(parsed.Metadata.Tags) > 0 {
			derivedTags := normalizeTags(parsed.Metadata.Tags)
			if err := replaceTags(p.db.WithContext(p.ctx), id, derivedTags); err != nil {
				return Model{}, cooklang.ParseResult{}, err
			}
		}
		if parsed.Metadata.Source != "" && (e.SourceURL == nil || *e.SourceURL == "") {
			e.SourceURL = &parsed.Metadata.Source
		}
	}
	if attrs.Servings != nil {
		e.Servings = attrs.Servings
	}
	if attrs.PrepTimeMinutes != nil {
		e.PrepTimeMinutes = attrs.PrepTimeMinutes
	}
	if attrs.CookTimeMinutes != nil {
		e.CookTimeMinutes = attrs.CookTimeMinutes
	}
	if attrs.SourceURL != nil {
		e.SourceURL = attrs.SourceURL
	}
	e.UpdatedAt = time.Now().UTC()

	if err := save(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, cooklang.ParseResult{}, err
	}

	if attrs.Tags != nil && attrs.Source == nil {
		// Only replace tags from explicit attrs if source wasn't changed
		// (source changes derive tags from metadata above)
		if err := replaceTags(p.db.WithContext(p.ctx), id, *attrs.Tags); err != nil {
			return Model{}, cooklang.ParseResult{}, err
		}
	}

	// Re-fetch to get updated tags
	m, err := p.ByIDProvider(id)()
	if err != nil {
		return Model{}, cooklang.ParseResult{}, err
	}
	return m, cooklang.Parse(m.Source()), nil
}

func (p *Processor) Delete(id uuid.UUID) error {
	return softDelete(p.db.WithContext(p.ctx), id)
}

func (p *Processor) Restore(id uuid.UUID) (Model, cooklang.ParseResult, error) {
	m, err := model.Map(Make)(getDeletedByID(id)(p.db.WithContext(p.ctx)))()
	if err != nil {
		return Model{}, cooklang.ParseResult{}, ErrNotFound
	}
	if !m.IsDeleted() {
		return Model{}, cooklang.ParseResult{}, ErrNotDeleted
	}
	if time.Since(*m.DeletedAt()) > restoreWindowDays*24*time.Hour {
		return Model{}, cooklang.ParseResult{}, ErrRestoreWindow
	}

	if err := restoreByID(p.db.WithContext(p.ctx), id); err != nil {
		return Model{}, cooklang.ParseResult{}, err
	}

	re := RestorationEntity{Id: uuid.New(), RecipeId: id, RestoredAt: time.Now().UTC()}
	p.db.WithContext(p.ctx).Create(&re)

	restored, err := p.ByIDProvider(id)()
	if err != nil {
		return Model{}, cooklang.ParseResult{}, err
	}
	return restored, cooklang.Parse(restored.Source()), nil
}

func (p *Processor) ListTags() ([]TagCount, error) {
	return getAllTags(p.db.WithContext(p.ctx))
}

func (p *Processor) ParseSource(source string) cooklang.ParseResult {
	result := cooklang.Parse(source)
	result.Errors = cooklang.Validate(source)
	return result
}

func mergeStringSlices(slices ...[]string) []string {
	var result []string
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, t := range tags {
		normalized := strings.ToLower(strings.TrimSpace(t))
		if normalized != "" && !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}
	return result
}
