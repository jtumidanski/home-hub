package wishlist

import (
	"time"

	"github.com/google/uuid"
)

type RestModel struct {
	Id               uuid.UUID `json:"-"`
	Name             string    `json:"name"`
	PurchaseLocation *string   `json:"purchase_location"`
	Urgency          string    `json:"urgency"`
	VoteCount        int       `json:"vote_count"`
	CreatedBy        uuid.UUID `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (r RestModel) GetName() string         { return "wish-items" }
func (r RestModel) GetID() string           { return r.Id.String() }
func (r *RestModel) SetID(id string) error  { var err error; r.Id, err = uuid.Parse(id); return err }

type CreateRequest struct {
	Id               uuid.UUID `json:"-"`
	Name             string    `json:"name"`
	PurchaseLocation *string   `json:"purchase_location,omitempty"`
	Urgency          *string   `json:"urgency,omitempty"`
}

func (r CreateRequest) GetName() string { return "wish-items" }
func (r CreateRequest) GetID() string   { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

type UpdateRequest struct {
	Id               uuid.UUID `json:"-"`
	Name             *string   `json:"name,omitempty"`
	PurchaseLocation *string   `json:"purchase_location,omitempty"`
	Urgency          *string   `json:"urgency,omitempty"`
	// VoteCount must NEVER be modifiable through PATCH. We accept it as a
	// pointer purely so we can detect and reject the attempt with a clear
	// 400.
	VoteCount *int `json:"vote_count,omitempty"`
}

func (r UpdateRequest) GetName() string         { return "wish-items" }
func (r UpdateRequest) GetID() string           { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error  { var err error; r.Id, err = uuid.Parse(id); return err }

type VoteRequest struct {
	Id uuid.UUID `json:"-"`
}

func (r VoteRequest) GetName() string { return "wish-items" }
func (r VoteRequest) GetID() string   { return r.Id.String() }
func (r *VoteRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:               m.Id(),
		Name:             m.Name(),
		PurchaseLocation: m.PurchaseLocation(),
		Urgency:          m.Urgency(),
		VoteCount:        m.VoteCount(),
		CreatedBy:        m.CreatedBy(),
		CreatedAt:        m.CreatedAt(),
		UpdatedAt:        m.UpdatedAt(),
	}, nil
}

func TransformSlice(models []Model) ([]RestModel, error) {
	out := make([]RestModel, len(models))
	for i, m := range models {
		r, err := Transform(m)
		if err != nil {
			return nil, err
		}
		out[i] = r
	}
	return out, nil
}
