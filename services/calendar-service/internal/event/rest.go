package event

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

var recurringInstancePattern = regexp.MustCompile(`_\d{8}(T\d{6}Z)?$`)

type CreateEventRequest struct {
	Id          uuid.UUID `json:"-"`
	Title       string    `json:"title"`
	Start       string    `json:"start"`
	End         string    `json:"end"`
	AllDay      bool      `json:"allDay"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Recurrence  []string  `json:"recurrence"`
	TimeZone    string    `json:"timeZone"`
}

func (r CreateEventRequest) GetName() string { return "calendar-events" }
func (r CreateEventRequest) GetID() string   { return r.Id.String() }
func (r *CreateEventRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type UpdateEventRequest struct {
	Id          uuid.UUID `json:"-"`
	Title       *string   `json:"title"`
	Start       *string   `json:"start"`
	End         *string   `json:"end"`
	AllDay      *bool     `json:"allDay"`
	Location    *string   `json:"location"`
	Description *string   `json:"description"`
	Scope       string    `json:"scope"`
	TimeZone    *string   `json:"timeZone"`
	Recurrence  *[]string `json:"recurrence"`
}

func (r UpdateEventRequest) GetName() string { return "calendar-events" }
func (r UpdateEventRequest) GetID() string   { return r.Id.String() }
func (r *UpdateEventRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type RestModel struct {
	Id              uuid.UUID  `json:"-"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	StartTime       time.Time  `json:"startTime"`
	EndTime         time.Time  `json:"endTime"`
	AllDay          bool       `json:"allDay"`
	Location        *string    `json:"location"`
	Visibility      string     `json:"visibility"`
	UserDisplayName string     `json:"userDisplayName"`
	UserColor       string     `json:"userColor"`
	IsOwner         bool       `json:"isOwner"`
	SourceId        string     `json:"sourceId"`
	ConnectionId    string     `json:"connectionId"`
	IsRecurring     bool       `json:"isRecurring"`
}

func (r RestModel) GetName() string { return "calendar-events" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return TransformWithPrivacy(m, uuid.Nil)
}

func TransformSlice(models []Model) ([]RestModel, error) {
	return TransformSliceWithPrivacy(models, uuid.Nil)
}

func TransformWithPrivacy(m Model, requesterUserID uuid.UUID) (RestModel, error) {
	isOwner := m.UserID() == requesterUserID
	rm := RestModel{
		Id:              m.Id(),
		StartTime:       m.StartTime(),
		EndTime:         m.EndTime(),
		AllDay:          m.AllDay(),
		Visibility:      m.Visibility(),
		UserDisplayName: m.UserDisplayName(),
		UserColor:       m.UserColor(),
		IsOwner:         isOwner,
		SourceId:        m.SourceID().String(),
		ConnectionId:    m.ConnectionID().String(),
		IsRecurring:     recurringInstancePattern.MatchString(m.ExternalID()),
	}

	if m.IsPrivate() && !isOwner {
		rm.Title = "Busy"
		rm.Description = nil
		rm.Location = nil
	} else {
		rm.Title = m.Title()
		desc := m.Description()
		if desc != "" {
			rm.Description = &desc
		}
		loc := m.Location()
		if loc != "" {
			rm.Location = &loc
		}
	}
	return rm, nil
}

func TransformSliceWithPrivacy(models []Model, requesterUserID uuid.UUID) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := TransformWithPrivacy(m, requesterUserID)
		if err != nil {
			return nil, err
		}
		result[i] = rm
	}
	return result, nil
}
