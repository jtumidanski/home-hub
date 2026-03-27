package household

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/api2go/jsonapi"
	"gorm.io/gorm"
)

type MemberView struct {
	MembershipID uuid.UUID
	UserID       uuid.UUID
	HouseholdID  uuid.UUID
	Role         string
	DisplayName  string
}

func getMembersByHousehold(db *gorm.DB, householdID uuid.UUID) ([]MemberView, error) {
	var members []MemberView
	err := db.Raw(`
		SELECT m.id AS membership_id, m.user_id, m.household_id, m.role, COALESCE(u.display_name, u.email) AS display_name
		FROM account.memberships m
		JOIN auth.users u ON u.id = m.user_id
		WHERE m.household_id = ?
	`, householdID).Scan(&members).Error
	return members, err
}

type MemberRestModel struct {
	Id          uuid.UUID `json:"-"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role"`

	UserID      uuid.UUID `json:"-"`
	HouseholdID uuid.UUID `json:"-"`
}

func (r MemberRestModel) GetName() string { return "members" }
func (r MemberRestModel) GetID() string   { return r.Id.String() }
func (r *MemberRestModel) SetID(id string) error {
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

func (r MemberRestModel) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{Type: "users", Name: "user", Relationship: jsonapi.ToOneRelationship},
		{Type: "households", Name: "household", Relationship: jsonapi.ToOneRelationship},
	}
}

func (r MemberRestModel) GetReferencedIDs() []jsonapi.ReferenceID {
	return []jsonapi.ReferenceID{
		{ID: r.UserID.String(), Type: "users", Name: "user", Relationship: jsonapi.ToOneRelationship},
		{ID: r.HouseholdID.String(), Type: "households", Name: "household", Relationship: jsonapi.ToOneRelationship},
	}
}

func TransformMember(v MemberView) MemberRestModel {
	return MemberRestModel{
		Id:          v.MembershipID,
		DisplayName: v.DisplayName,
		Role:        v.Role,
		UserID:      v.UserID,
		HouseholdID: v.HouseholdID,
	}
}

func TransformMembers(views []MemberView) []MemberRestModel {
	result := make([]MemberRestModel, len(views))
	for i, v := range views {
		result[i] = TransformMember(v)
	}
	return result
}
