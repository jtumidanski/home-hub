package googlecal

import "time"

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type CalendarListResponse struct {
	Items []CalendarListEntry `json:"items"`
}

type CalendarListEntry struct {
	ID              string `json:"id"`
	Summary         string `json:"summary"`
	Primary         bool   `json:"primary"`
	BackgroundColor string `json:"backgroundColor"`
}

type EventsResponse struct {
	Items         []Event `json:"items"`
	NextSyncToken string  `json:"nextSyncToken"`
	NextPageToken string  `json:"nextPageToken"`
}

type Event struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Summary     string     `json:"summary"`
	Description string     `json:"description"`
	Location    string     `json:"location"`
	Start       *EventTime `json:"start"`
	End         *EventTime `json:"end"`
	Visibility  string     `json:"visibility"`
}

type EventTime struct {
	DateTime *time.Time `json:"dateTime,omitempty"`
	Date     string     `json:"date,omitempty"`
}

type InsertEventRequest struct {
	Summary     string     `json:"summary"`
	Location    string     `json:"location,omitempty"`
	Description string     `json:"description,omitempty"`
	Start       *EventTime `json:"start"`
	End         *EventTime `json:"end"`
	Recurrence  []string   `json:"recurrence,omitempty"`
}

type UpdateEventRequest struct {
	Summary     *string    `json:"summary,omitempty"`
	Location    *string    `json:"location,omitempty"`
	Description *string    `json:"description,omitempty"`
	Start       *EventTime `json:"start,omitempty"`
	End         *EventTime `json:"end,omitempty"`
}

func (et *EventTime) IsAllDay() bool {
	return et != nil && et.Date != ""
}

func (et *EventTime) Time() time.Time {
	if et == nil {
		return time.Time{}
	}
	if et.Date != "" {
		t, _ := time.Parse("2006-01-02", et.Date)
		return t
	}
	if et.DateTime != nil {
		return *et.DateTime
	}
	return time.Time{}
}
