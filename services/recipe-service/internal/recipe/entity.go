package recipe

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TenantId        uuid.UUID  `gorm:"type:uuid;not null;index:idx_recipe_tenant_household"`
	HouseholdId     uuid.UUID  `gorm:"type:uuid;not null;index:idx_recipe_tenant_household"`
	Title           string     `gorm:"type:varchar(255);not null"`
	Description     *string    `gorm:"type:text"`
	Source          string     `gorm:"type:text;not null"`
	Servings        *int       `gorm:"type:int"`
	PrepTimeMinutes *int       `gorm:"type:int"`
	CookTimeMinutes *int       `gorm:"type:int"`
	SourceURL       *string    `gorm:"type:varchar(2048)"`
	Tags            []TagEntity `gorm:"foreignKey:RecipeId;constraint:OnDelete:CASCADE"`
	DeletedAt       *time.Time `gorm:"index:idx_recipe_soft_delete"`
	CreatedAt       time.Time  `gorm:"not null"`
	UpdatedAt       time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "recipes" }

type TagEntity struct {
	Id       uuid.UUID `gorm:"type:uuid;primaryKey"`
	RecipeId uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_tag_recipe_unique"`
	Tag      string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_tag_recipe_unique;index:idx_tag_value"`
}

func (TagEntity) TableName() string { return "recipe_tags" }

type RestorationEntity struct {
	Id         uuid.UUID `gorm:"type:uuid;primaryKey"`
	RecipeId   uuid.UUID `gorm:"type:uuid;not null"`
	RestoredAt time.Time `gorm:"not null"`
}

func (RestorationEntity) TableName() string { return "recipe_restorations" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{}, &TagEntity{}, &RestorationEntity{})
}

func (m Model) ToEntity() Entity {
	var desc *string
	if m.description != "" {
		desc = &m.description
	}
	var srcURL *string
	if m.sourceURL != "" {
		srcURL = &m.sourceURL
	}
	tags := make([]TagEntity, len(m.tags))
	for i, t := range m.tags {
		tags[i] = TagEntity{Id: uuid.New(), RecipeId: m.id, Tag: strings.ToLower(strings.TrimSpace(t))}
	}
	return Entity{
		Id: m.id, TenantId: m.tenantID, HouseholdId: m.householdID,
		Title: m.title, Description: desc, Source: m.source,
		Servings: m.servings, PrepTimeMinutes: m.prepTimeMinutes, CookTimeMinutes: m.cookTimeMinutes,
		SourceURL: srcURL, Tags: tags,
		DeletedAt: m.deletedAt, CreatedAt: m.createdAt, UpdatedAt: m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	desc := ""
	if e.Description != nil {
		desc = *e.Description
	}
	srcURL := ""
	if e.SourceURL != nil {
		srcURL = *e.SourceURL
	}
	tags := make([]string, len(e.Tags))
	for i, t := range e.Tags {
		tags[i] = t.Tag
	}
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetHouseholdID(e.HouseholdId).
		SetTitle(e.Title).
		SetDescription(desc).
		SetSource(e.Source).
		SetServings(e.Servings).
		SetPrepTimeMinutes(e.PrepTimeMinutes).
		SetCookTimeMinutes(e.CookTimeMinutes).
		SetSourceURL(srcURL).
		SetTags(tags).
		SetDeletedAt(e.DeletedAt).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
