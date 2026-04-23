package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/layout"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Processor orchestrates dashboard CRUD with scope/visibility rules.
type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// ErrInvalidScope is returned when a caller provides a scope outside {household, user}.
var ErrInvalidScope = errors.New("invalid scope")

// CreateAttrs carries Create inputs for a dashboard.
type CreateAttrs struct {
	Name      string
	Scope     string // "household" | "user"
	Layout    json.RawMessage
	SortOrder *int
}

// Create inserts a new dashboard under the requested scope after validating
// name and layout.
func (p *Processor) Create(tenantID, householdID, callerUserID uuid.UUID, attrs CreateAttrs) (Model, error) {
	var userID *uuid.UUID
	switch attrs.Scope {
	case "household":
		userID = nil
	case "user":
		u := callerUserID
		userID = &u
	default:
		return Model{}, ErrInvalidScope
	}

	layoutBytes := attrs.Layout
	if len(layoutBytes) == 0 {
		layoutBytes = json.RawMessage(`{"version":1,"widgets":[]}`)
	}
	if _, err := layout.Validate(layoutBytes); err != nil {
		return Model{}, err
	}

	sortOrder := 0
	if attrs.SortOrder != nil {
		sortOrder = *attrs.SortOrder
	} else {
		max, err := maxSortOrderInScope(p.db.WithContext(p.ctx), tenantID, householdID, userID)
		if err != nil {
			return Model{}, err
		}
		sortOrder = max + 1
	}

	name := trimName(attrs.Name)
	if err := validateNameLen(name); err != nil {
		return Model{}, err
	}

	e := Entity{
		TenantId:      tenantID,
		HouseholdId:   householdID,
		UserId:        userID,
		Name:          name,
		SortOrder:     sortOrder,
		Layout:        datatypes.JSON(layoutBytes),
		SchemaVersion: 1,
	}
	saved, err := insert(p.db.WithContext(p.ctx), e)
	if err != nil {
		return Model{}, err
	}
	return Make(saved)
}

func trimName(s string) string {
	return strings.TrimSpace(s)
}

func validateNameLen(s string) error {
	if s == "" {
		return ErrNameRequired
	}
	if len(s) > 80 {
		return ErrNameTooLong
	}
	return nil
}
