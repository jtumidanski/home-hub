package connection

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id                 uuid.UUID  `json:"-"`
	Provider           string     `json:"provider"`
	Status             string     `json:"status"`
	Email              string     `json:"email"`
	UserDisplayName    string     `json:"userDisplayName"`
	UserColor          string     `json:"userColor"`
	LastSyncAt         *time.Time `json:"lastSyncAt"`
	LastSyncEventCount int        `json:"lastSyncEventCount"`
	CreatedAt          time.Time  `json:"createdAt"`
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
		Id:                 m.Id(),
		Provider:           m.Provider(),
		Status:             m.Status(),
		Email:              m.Email(),
		UserDisplayName:    m.UserDisplayName(),
		UserColor:          m.UserColor(),
		LastSyncAt:         m.LastSyncAt(),
		LastSyncEventCount: m.LastSyncEventCount(),
		CreatedAt:          m.CreatedAt(),
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
