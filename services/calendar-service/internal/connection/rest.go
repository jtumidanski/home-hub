package connection

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id                  uuid.UUID  `json:"-"`
	Provider            string     `json:"provider"`
	Status              string     `json:"status"`
	Email               string     `json:"email"`
	UserDisplayName     string     `json:"userDisplayName"`
	UserColor           string     `json:"userColor"`
	WriteAccess         bool       `json:"writeAccess"`
	LastSyncAt          *time.Time `json:"lastSyncAt"`
	LastSyncAttemptAt   *time.Time `json:"lastSyncAttemptAt"`
	LastSyncEventCount  int        `json:"lastSyncEventCount"`
	ErrorCode           *string    `json:"errorCode"`
	ErrorMessage        *string    `json:"errorMessage"`
	LastErrorAt         *time.Time `json:"lastErrorAt"`
	ConsecutiveFailures int        `json:"consecutiveFailures"`
	CreatedAt           time.Time  `json:"createdAt"`
}

func (r RestModel) GetName() string { return "calendar-connections" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:                  m.Id(),
		Provider:            m.Provider(),
		Status:              m.Status(),
		Email:               m.Email(),
		UserDisplayName:     m.UserDisplayName(),
		UserColor:           m.UserColor(),
		WriteAccess:         m.WriteAccess(),
		LastSyncAt:          m.LastSyncAt(),
		LastSyncAttemptAt:   m.LastSyncAttemptAt(),
		LastSyncEventCount:  m.LastSyncEventCount(),
		ErrorCode:           m.ErrorCode(),
		ErrorMessage:        m.ErrorMessage(),
		LastErrorAt:         m.LastErrorAt(),
		ConsecutiveFailures: m.ConsecutiveFailures(),
		CreatedAt:           m.CreatedAt(),
	}, nil
}

func TransformSlice(models []Model) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := Transform(m)
		if err != nil {
			return nil, err
		}
		result[i] = rm
	}
	return result, nil
}

type AuthorizeRequest struct {
	Id          uuid.UUID `json:"-"`
	RedirectUri string    `json:"redirectUri"`
	Reauthorize bool      `json:"reauthorize"`
}

func (r AuthorizeRequest) GetName() string { return "calendar-authorization-requests" }
func (r AuthorizeRequest) GetID() string   { return r.Id.String() }
func (r *AuthorizeRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type TriggerSyncRequest struct {
	Id uuid.UUID `json:"-"`
}

func (r TriggerSyncRequest) GetName() string { return "calendar-sync-triggers" }
func (r TriggerSyncRequest) GetID() string   { return r.Id.String() }
func (r *TriggerSyncRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type AuthorizeResponse struct {
	Id           uuid.UUID `json:"-"`
	AuthorizeUrl string    `json:"authorizeUrl"`
}

func (r AuthorizeResponse) GetName() string { return "calendar-authorization-responses" }
func (r AuthorizeResponse) GetID() string   { return r.Id.String() }
func (r *AuthorizeResponse) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}
