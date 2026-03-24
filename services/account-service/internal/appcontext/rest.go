package appcontext

import (
	"github.com/jtumidanski/home-hub/services/account-service/internal/household"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
	"github.com/jtumidanski/home-hub/services/account-service/internal/preference"
	"github.com/jtumidanski/home-hub/services/account-service/internal/tenant"
	"github.com/jtumidanski/api2go/jsonapi"
)

// RestModel represents the /contexts/current resource in JSON:API format.
// It implements api2go's MarshalLinkedRelations and MarshalIncludedRelations
// so that relationships and ?include= are handled automatically.
type RestModel struct {
	ResolvedTheme      string `json:"resolvedTheme"`
	ResolvedRole       string `json:"resolvedRole"`
	CanCreateHousehold bool   `json:"canCreateHousehold"`

	// Related resources (not serialized as attributes)
	Tenant          tenant.RestModel       `json:"-"`
	Preference      preference.RestModel   `json:"-"`
	ActiveHousehold *household.RestModel   `json:"-"`
	Memberships     []membership.RestModel `json:"-"`
}

func (r RestModel) GetName() string { return "contexts" }
func (r RestModel) GetID() string   { return "current" }
func (r *RestModel) SetID(_ string) error {
	return nil
}

// GetReferences declares the possible relationships on this resource.
func (r RestModel) GetReferences() []jsonapi.Reference {
	refs := []jsonapi.Reference{
		{Type: "tenants", Name: "tenant", Relationship: jsonapi.ToOneRelationship},
		{Type: "preferences", Name: "preference", Relationship: jsonapi.ToOneRelationship},
		{Type: "households", Name: "activeHousehold", Relationship: jsonapi.ToOneRelationship},
		{Type: "memberships", Name: "memberships", Relationship: jsonapi.ToManyRelationship},
	}
	return refs
}

// GetReferencedIDs returns the IDs of all related resources.
func (r RestModel) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{
		{ID: r.Tenant.GetID(), Type: "tenants", Name: "tenant", Relationship: jsonapi.ToOneRelationship},
		{ID: r.Preference.GetID(), Type: "preferences", Name: "preference", Relationship: jsonapi.ToOneRelationship},
	}

	if r.ActiveHousehold != nil {
		result = append(result, jsonapi.ReferenceID{
			ID: r.ActiveHousehold.GetID(), Type: "households", Name: "activeHousehold",
			Relationship: jsonapi.ToOneRelationship,
		})
	}

	for _, m := range r.Memberships {
		result = append(result, jsonapi.ReferenceID{
			ID: m.GetID(), Type: "memberships", Name: "memberships",
			Relationship: jsonapi.ToManyRelationship,
		})
	}

	return result
}

// GetReferencedStructs returns the actual related resources for inclusion
// in the JSON:API "included" section.
func (r RestModel) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	var result []jsonapi.MarshalIdentifier

	result = append(result, r.Tenant)
	result = append(result, r.Preference)

	if r.ActiveHousehold != nil {
		result = append(result, *r.ActiveHousehold)
	}

	for _, m := range r.Memberships {
		result = append(result, m)
	}

	return result
}

// TransformContext converts a Resolved domain object into a RestModel.
func TransformContext(resolved *Resolved) (RestModel, error) {
	tenantRest, err := tenant.Transform(resolved.Tenant)
	if err != nil {
		return RestModel{}, err
	}
	prefRest, err := preference.Transform(resolved.Preference)
	if err != nil {
		return RestModel{}, err
	}

	rm := RestModel{
		ResolvedTheme:      resolved.Preference.Theme(),
		ResolvedRole:       resolved.ResolvedRole,
		CanCreateHousehold: resolved.CanCreateHousehold,
		Tenant:             tenantRest,
		Preference:         prefRest,
	}

	if resolved.ActiveHousehold != nil {
		hhRest, err := household.Transform(*resolved.ActiveHousehold)
		if err != nil {
			return RestModel{}, err
		}
		rm.ActiveHousehold = &hhRest
	}

	memRest, err := membership.TransformSlice(resolved.Memberships)
	if err != nil {
		return RestModel{}, err
	}
	rm.Memberships = memRest

	return rm, nil
}
