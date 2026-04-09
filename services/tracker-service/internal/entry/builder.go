package entry

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTrackingItemRequired = errors.New("tracking item ID is required")
	ErrDateRequired         = errors.New("date is required")
	ErrFutureDate           = errors.New("cannot create entries for future dates")
	ErrNoteTooLong          = errors.New("note must not exceed 500 characters")
	ErrInvalidSentiment     = errors.New("sentiment value must be positive, neutral, or negative")
	ErrInvalidNumeric       = errors.New("numeric value must be a non-negative integer")
	ErrInvalidRange         = errors.New("range value is out of bounds")
	ErrValueRequired        = errors.New("value is required for non-skipped entries")
)

type Builder struct {
	id             uuid.UUID
	tenantID       uuid.UUID
	userID         uuid.UUID
	trackingItemID uuid.UUID
	date           time.Time
	value          json.RawMessage
	skipped        bool
	note           *string
	createdAt      time.Time
	updatedAt      time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder                { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder           { b.tenantID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder             { b.userID = id; return b }
func (b *Builder) SetTrackingItemID(id uuid.UUID) *Builder     { b.trackingItemID = id; return b }
func (b *Builder) SetDate(d time.Time) *Builder                { b.date = d; return b }
func (b *Builder) SetValue(v json.RawMessage) *Builder         { b.value = v; return b }
func (b *Builder) SetSkipped(s bool) *Builder                  { b.skipped = s; return b }
func (b *Builder) SetNote(n *string) *Builder                  { b.note = n; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder           { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder           { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.trackingItemID == uuid.Nil {
		return Model{}, ErrTrackingItemRequired
	}
	if b.date.IsZero() {
		return Model{}, ErrDateRequired
	}
	if b.note != nil && len(*b.note) > 500 {
		return Model{}, ErrNoteTooLong
	}
	return Model{
		id:             b.id,
		tenantID:       b.tenantID,
		userID:         b.userID,
		trackingItemID: b.trackingItemID,
		date:           b.date,
		value:          b.value,
		skipped:        b.skipped,
		note:           b.note,
		createdAt:      b.createdAt,
		updatedAt:      b.updatedAt,
	}, nil
}

type SentimentValue struct {
	Rating string `json:"rating"`
}

type NumericValue struct {
	Count int `json:"count"`
}

type RangeValue struct {
	Value int `json:"value"`
}

func ValidateValue(scaleType string, value json.RawMessage, scaleConfig json.RawMessage) error {
	if len(value) == 0 || string(value) == "null" {
		return ErrValueRequired
	}

	switch scaleType {
	case "sentiment":
		var sv SentimentValue
		if err := json.Unmarshal(value, &sv); err != nil {
			return ErrInvalidSentiment
		}
		if sv.Rating != "positive" && sv.Rating != "neutral" && sv.Rating != "negative" {
			return ErrInvalidSentiment
		}
	case "numeric":
		var nv NumericValue
		if err := json.Unmarshal(value, &nv); err != nil {
			return ErrInvalidNumeric
		}
		if nv.Count < 0 {
			return ErrInvalidNumeric
		}
	case "range":
		var rv RangeValue
		if err := json.Unmarshal(value, &rv); err != nil {
			return ErrInvalidRange
		}
		if len(scaleConfig) > 0 && string(scaleConfig) != "null" {
			var rc struct {
				Min int `json:"min"`
				Max int `json:"max"`
			}
			if err := json.Unmarshal(scaleConfig, &rc); err == nil {
				if rv.Value < rc.Min || rv.Value > rc.Max {
					return ErrInvalidRange
				}
			}
		}
	}
	return nil
}
