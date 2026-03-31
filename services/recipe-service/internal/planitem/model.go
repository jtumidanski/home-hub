package planitem

import (
	"time"

	"github.com/google/uuid"
)

const (
	SlotBreakfast = "breakfast"
	SlotLunch     = "lunch"
	SlotDinner    = "dinner"
	SlotSnack     = "snack"
	SlotSide      = "side"
)

var ValidSlots = []string{SlotBreakfast, SlotLunch, SlotDinner, SlotSnack, SlotSide}

type Model struct {
	id                uuid.UUID
	planWeekID        uuid.UUID
	day               time.Time
	slot              string
	recipeID          uuid.UUID
	servingMultiplier *float64
	plannedServings   *int
	notes             *string
	position          int
	createdAt         time.Time
	updatedAt         time.Time
}

func (m Model) Id() uuid.UUID              { return m.id }
func (m Model) PlanWeekID() uuid.UUID      { return m.planWeekID }
func (m Model) Day() time.Time             { return m.day }
func (m Model) Slot() string               { return m.slot }
func (m Model) RecipeID() uuid.UUID        { return m.recipeID }
func (m Model) ServingMultiplier() *float64 { return m.servingMultiplier }
func (m Model) PlannedServings() *int      { return m.plannedServings }
func (m Model) Notes() *string             { return m.notes }
func (m Model) Position() int              { return m.position }
func (m Model) CreatedAt() time.Time       { return m.createdAt }
func (m Model) UpdatedAt() time.Time       { return m.updatedAt }

func IsValidSlot(slot string) bool {
	for _, s := range ValidSlots {
		if s == slot {
			return true
		}
	}
	return false
}
