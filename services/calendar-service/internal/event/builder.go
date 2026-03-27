package event

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExternalIDRequired = errors.New("external ID is required")
	ErrTitleRequired      = errors.New("title is required")
)

type Builder struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	connectionID    uuid.UUID
	sourceID        uuid.UUID
	userID          uuid.UUID
	externalID      string
	googleCalendarID string
	title           string
	description     string
	startTime       time.Time
	endTime         time.Time
	allDay          bool
	location        string
	visibility      string
	userDisplayName string
	userColor       string
	createdAt       time.Time
	updatedAt       time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder              { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder         { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder      { b.householdID = id; return b }
func (b *Builder) SetConnectionID(id uuid.UUID) *Builder     { b.connectionID = id; return b }
func (b *Builder) SetSourceID(id uuid.UUID) *Builder         { b.sourceID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder           { b.userID = id; return b }
func (b *Builder) SetExternalID(eid string) *Builder         { b.externalID = eid; return b }
func (b *Builder) SetGoogleCalendarID(id string) *Builder    { b.googleCalendarID = id; return b }
func (b *Builder) SetTitle(t string) *Builder                { b.title = t; return b }
func (b *Builder) SetDescription(d string) *Builder          { b.description = d; return b }
func (b *Builder) SetStartTime(t time.Time) *Builder         { b.startTime = t; return b }
func (b *Builder) SetEndTime(t time.Time) *Builder           { b.endTime = t; return b }
func (b *Builder) SetAllDay(a bool) *Builder                 { b.allDay = a; return b }
func (b *Builder) SetLocation(l string) *Builder             { b.location = l; return b }
func (b *Builder) SetVisibility(v string) *Builder           { b.visibility = v; return b }
func (b *Builder) SetUserDisplayName(n string) *Builder      { b.userDisplayName = n; return b }
func (b *Builder) SetUserColor(c string) *Builder            { b.userColor = c; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder         { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder         { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.externalID == "" {
		return Model{}, ErrExternalIDRequired
	}
	if b.title == "" {
		return Model{}, ErrTitleRequired
	}
	return Model{
		id:              b.id,
		tenantID:        b.tenantID,
		householdID:     b.householdID,
		connectionID:    b.connectionID,
		sourceID:        b.sourceID,
		userID:          b.userID,
		externalID:      b.externalID,
		googleCalendarID: b.googleCalendarID,
		title:           b.title,
		description:     b.description,
		startTime:       b.startTime,
		endTime:         b.endTime,
		allDay:          b.allDay,
		location:        b.location,
		visibility:      b.visibility,
		userDisplayName: b.userDisplayName,
		userColor:       b.userColor,
		createdAt:       b.createdAt,
		updatedAt:       b.updatedAt,
	}, nil
}
