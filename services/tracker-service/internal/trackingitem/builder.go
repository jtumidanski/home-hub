package trackingitem

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired       = errors.New("tracking item name is required")
	ErrNameTooLong        = errors.New("tracking item name must not exceed 100 characters")
	ErrInvalidScaleType   = errors.New("scale type must be one of: sentiment, numeric, range")
	ErrInvalidColor       = errors.New("invalid color; must be from the allowed palette")
	ErrInvalidSortOrder   = errors.New("sort order must be non-negative")
	ErrRangeConfigRequired = errors.New("range scale type requires scale_config with min and max")
	ErrInvalidRangeConfig = errors.New("range min must be less than max")
	ErrInvalidScheduleDay = errors.New("schedule days must be integers 0-6 (Sun-Sat)")
)

var validScaleTypes = map[string]bool{
	"sentiment": true,
	"numeric":   true,
	"range":     true,
}

var validColors = map[string]bool{
	"red": true, "orange": true, "amber": true, "yellow": true,
	"lime": true, "green": true, "emerald": true, "teal": true,
	"cyan": true, "blue": true, "indigo": true, "violet": true,
	"purple": true, "fuchsia": true, "pink": true, "rose": true,
}

type RangeConfig struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	userID      uuid.UUID
	name        string
	scaleType   string
	scaleConfig json.RawMessage
	color       string
	sortOrder   int
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder              { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder         { b.tenantID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder           { b.userID = id; return b }
func (b *Builder) SetName(name string) *Builder              { b.name = name; return b }
func (b *Builder) SetScaleType(st string) *Builder           { b.scaleType = st; return b }
func (b *Builder) SetScaleConfig(sc json.RawMessage) *Builder { b.scaleConfig = sc; return b }
func (b *Builder) SetColor(c string) *Builder                { b.color = c; return b }
func (b *Builder) SetSortOrder(o int) *Builder               { b.sortOrder = o; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder         { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder         { b.updatedAt = t; return b }
func (b *Builder) SetDeletedAt(t *time.Time) *Builder        { b.deletedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if len(b.name) > 100 {
		return Model{}, ErrNameTooLong
	}
	if !validScaleTypes[b.scaleType] {
		return Model{}, ErrInvalidScaleType
	}
	if !validColors[b.color] {
		return Model{}, ErrInvalidColor
	}
	if b.sortOrder < 0 {
		return Model{}, ErrInvalidSortOrder
	}
	if b.scaleType == "range" {
		if len(b.scaleConfig) == 0 || string(b.scaleConfig) == "null" {
			return Model{}, ErrRangeConfigRequired
		}
		var rc RangeConfig
		if err := json.Unmarshal(b.scaleConfig, &rc); err != nil {
			return Model{}, ErrRangeConfigRequired
		}
		if rc.Min >= rc.Max {
			return Model{}, ErrInvalidRangeConfig
		}
	}
	return Model{
		id:          b.id,
		tenantID:    b.tenantID,
		userID:      b.userID,
		name:        b.name,
		scaleType:   b.scaleType,
		scaleConfig: b.scaleConfig,
		color:       b.color,
		sortOrder:   b.sortOrder,
		createdAt:   b.createdAt,
		updatedAt:   b.updatedAt,
		deletedAt:   b.deletedAt,
	}, nil
}

func ValidateSchedule(days []int) error {
	for _, d := range days {
		if d < 0 || d > 6 {
			return ErrInvalidScheduleDay
		}
	}
	return nil
}
