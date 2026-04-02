package invitation

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
)

type RestModel struct {
	Id        uuid.UUID `json:"-"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	HouseholdID uuid.UUID `json:"-"`
	InvitedByID uuid.UUID `json:"-"`
}

func (r RestModel) GetName() string { return "invitations" }
func (r RestModel) GetID() string   { return r.Id.String() }
func (r *RestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func (r RestModel) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{Type: "households", Name: "household", Relationship: jsonapi.ToOneRelationship},
		{Type: "users", Name: "invitedBy", Relationship: jsonapi.ToOneRelationship},
	}
}

func (r RestModel) GetReferencedIDs() []jsonapi.ReferenceID {
	return []jsonapi.ReferenceID{
		{ID: r.HouseholdID.String(), Type: "households", Name: "household", Relationship: jsonapi.ToOneRelationship},
		{ID: r.InvitedByID.String(), Type: "users", Name: "invitedBy", Relationship: jsonapi.ToOneRelationship},
	}
}

func Transform(m Model) (RestModel, error) {
	return RestModel{
		Id:          m.Id(),
		Email:       m.Email(),
		Role:        m.Role(),
		Status:      m.Status(),
		ExpiresAt:   m.ExpiresAt(),
		CreatedAt:   m.CreatedAt(),
		UpdatedAt:   m.UpdatedAt(),
		HouseholdID: m.HouseholdID(),
		InvitedByID: m.InvitedBy(),
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

type CreateRequest struct {
	Id    uuid.UUID `json:"-"`
	Email string    `json:"email"`
	Role  string    `json:"role"`

	HouseholdID uuid.UUID `json:"-"`
}

func (r CreateRequest) GetName() string { return "invitations" }
func (r CreateRequest) GetID() string   { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func (r CreateRequest) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{Type: "households", Name: "household", Relationship: jsonapi.ToOneRelationship},
	}
}

// MineRestModel extends RestModel with included household for the /invitations/mine endpoint.
type MineRestModel struct {
	RestModel
	Household *household.RestModel `json:"-"`
}

func (r MineRestModel) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	if r.Household != nil {
		return []jsonapi.MarshalIdentifier{*r.Household}
	}
	return nil
}

// TransformMineSlice converts invitation REST models and household models into MineRestModels
// with embedded household REST representations.
func TransformMineSlice(rest []RestModel, households []household.Model) ([]MineRestModel, error) {
	hhMap := make(map[string]*household.RestModel, len(households))
	for _, hh := range households {
		hhRest, err := household.Transform(hh)
		if err != nil {
			return nil, err
		}
		hhMap[hh.Id().String()] = &hhRest
	}

	result := make([]MineRestModel, len(rest))
	for i, rm := range rest {
		result[i] = MineRestModel{RestModel: rm, Household: hhMap[rm.HouseholdID.String()]}
	}
	return result, nil
}

func (r *CreateRequest) SetToOneReferenceID(name, ID string) error {
	if name == "household" {
		id, err := uuid.Parse(ID)
		if err != nil {
			return err
		}
		r.HouseholdID = id
	}
	return nil
}
