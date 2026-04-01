package item

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id                uuid.UUID
	listID            uuid.UUID
	name              string
	quantity          *string
	categoryID        *uuid.UUID
	categoryName      *string
	categorySortOrder *int
	checked           bool
	position          int
	createdAt         time.Time
	updatedAt         time.Time
}

func (m Model) Id() uuid.UUID            { return m.id }
func (m Model) ListID() uuid.UUID        { return m.listID }
func (m Model) Name() string             { return m.name }
func (m Model) Quantity() *string         { return m.quantity }
func (m Model) CategoryID() *uuid.UUID   { return m.categoryID }
func (m Model) CategoryName() *string    { return m.categoryName }
func (m Model) CategorySortOrder() *int  { return m.categorySortOrder }
func (m Model) Checked() bool            { return m.checked }
func (m Model) Position() int            { return m.position }
func (m Model) CreatedAt() time.Time     { return m.createdAt }
func (m Model) UpdatedAt() time.Time     { return m.updatedAt }
