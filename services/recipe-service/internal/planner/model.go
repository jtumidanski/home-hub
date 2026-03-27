package planner

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	id                 uuid.UUID
	recipeID           uuid.UUID
	classification     string
	servingsYield      *int
	eatWithinDays      *int
	minGapDays         *int
	maxConsecutiveDays *int
	createdAt          time.Time
	updatedAt          time.Time
}

func (m Model) Id() uuid.UUID            { return m.id }
func (m Model) RecipeID() uuid.UUID      { return m.recipeID }
func (m Model) Classification() string   { return m.classification }
func (m Model) ServingsYield() *int      { return m.servingsYield }
func (m Model) EatWithinDays() *int      { return m.eatWithinDays }
func (m Model) MinGapDays() *int         { return m.minGapDays }
func (m Model) MaxConsecutiveDays() *int { return m.maxConsecutiveDays }
func (m Model) CreatedAt() time.Time     { return m.createdAt }
func (m Model) UpdatedAt() time.Time     { return m.updatedAt }
