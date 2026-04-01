package item

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ListId            uuid.UUID  `gorm:"type:uuid;not null;index:idx_shopping_item_list_checked_sort"`
	Name              string     `gorm:"type:varchar(255);not null"`
	Quantity          *string    `gorm:"type:varchar(100)"`
	CategoryId        *uuid.UUID `gorm:"type:uuid"`
	CategoryName      *string    `gorm:"type:varchar(100)"`
	CategorySortOrder *int       `gorm:"index:idx_shopping_item_list_checked_sort"`
	Checked           bool       `gorm:"not null;default:false;index:idx_shopping_item_list_checked_sort"`
	Position          int        `gorm:"not null;default:0;index:idx_shopping_item_list_checked_sort"`
	CreatedAt         time.Time  `gorm:"not null"`
	UpdatedAt         time.Time  `gorm:"not null"`
}

func (Entity) TableName() string { return "shopping_items" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:                m.id,
		ListId:            m.listID,
		Name:              m.name,
		Quantity:          m.quantity,
		CategoryId:        m.categoryID,
		CategoryName:      m.categoryName,
		CategorySortOrder: m.categorySortOrder,
		Checked:           m.checked,
		Position:          m.position,
		CreatedAt:         m.createdAt,
		UpdatedAt:         m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	b := NewBuilder().
		SetId(e.Id).
		SetListID(e.ListId).
		SetName(e.Name).
		SetChecked(e.Checked).
		SetPosition(e.Position).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt)
	if e.Quantity != nil {
		b.SetQuantity(e.Quantity)
	}
	if e.CategoryId != nil {
		b.SetCategoryID(e.CategoryId)
	}
	if e.CategoryName != nil {
		b.SetCategoryName(e.CategoryName)
	}
	if e.CategorySortOrder != nil {
		b.SetCategorySortOrder(e.CategorySortOrder)
	}
	return b.Build()
}
