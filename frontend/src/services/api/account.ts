import { api, type RequestOptions } from "@/lib/api/client";
import type { ApiResponse, ApiListResponse } from "@/types/api/responses";
import type { AppContext } from "@/types/models/context";
import type { Tenant } from "@/types/models/tenant";
import type { TenantCreateAttributes } from "@/types/models/tenant";
import type { Household, HouseholdUpdateAttributes, HouseholdCreateAttributes } from "@/types/models/household";
import type { Invitation, InvitationCreateAttributes } from "@/types/models/invitation";
import type { Membership } from "@/types/models/membership";
import type { Preference } from "@/types/models/preference";

class AccountService {
  getContext(options?: RequestOptions) {
    return api.get<ApiResponse<AppContext>>(
      "/contexts/current?include=tenant,activeHousehold,preference,memberships",
      options,
    );
  }

  createTenant(attrs: TenantCreateAttributes) {
    return api.post<ApiResponse<Tenant>>("/tenants", {
      data: { type: "tenants", attributes: attrs },
    });
  }

  listHouseholds(tenant: Tenant) {
    api.setTenant(tenant);
    return api.get<ApiListResponse<Household>>("/households");
  }

  createHousehold(tenant: Tenant, attrs: HouseholdCreateAttributes) {
    api.setTenant(tenant);
    return api.post<ApiResponse<Household>>("/households", {
      data: { type: "households", attributes: attrs },
    });
  }

  // Merge not needed: "theme" is the only mutable field on PreferenceAttributes
  // (createdAt and updatedAt are server-managed), so the patch payload is already complete.
  updatePreferenceTheme(tenant: Tenant, preferenceId: string, theme: "light" | "dark") {
    api.setTenant(tenant);
    return api.patch<ApiResponse<Preference>>(`/preferences/${preferenceId}`, {
      data: { type: "preferences", id: preferenceId, attributes: { theme } },
    });
  }

  updateHousehold(tenant: Tenant, householdId: string, attrs: HouseholdUpdateAttributes) {
    api.setTenant(tenant);
    return api.patch<ApiResponse<Household>>(`/households/${householdId}`, {
      data: { type: "households", id: householdId, attributes: attrs },
    });
  }

  // --- Invitation endpoints ---

  listInvitationsByHousehold(tenant: Tenant, householdId: string) {
    api.setTenant(tenant);
    return api.get<ApiListResponse<Invitation>>(
      `/invitations?filter[householdId]=${householdId}`,
    );
  }

  listMyInvitations() {
    return api.get<ApiListResponse<Invitation> & { included?: Household[] }>(
      "/invitations/mine?filter[status]=pending",
    );
  }

  createInvitation(tenant: Tenant, householdId: string, attrs: InvitationCreateAttributes) {
    api.setTenant(tenant);
    return api.post<ApiResponse<Invitation>>("/invitations", {
      data: {
        type: "invitations",
        attributes: { email: attrs.email, role: attrs.role ?? "viewer" },
        relationships: {
          household: { data: { type: "households", id: householdId } },
        },
      },
    });
  }

  revokeInvitation(tenant: Tenant, id: string) {
    api.setTenant(tenant);
    return api.delete(`/invitations/${id}`);
  }

  acceptInvitation(id: string) {
    return api.post<ApiResponse<Invitation>>(`/invitations/${id}/accept`);
  }

  declineInvitation(id: string) {
    return api.post<ApiResponse<Invitation>>(`/invitations/${id}/decline`);
  }

  // --- Membership endpoints ---

  listMembershipsByHousehold(tenant: Tenant, householdId: string) {
    api.setTenant(tenant);
    return api.get<ApiListResponse<Membership>>(
      `/memberships?filter[householdId]=${householdId}`,
    );
  }

  updateMembershipRole(tenant: Tenant, membershipId: string, role: string) {
    api.setTenant(tenant);
    return api.patch<ApiResponse<Membership>>(`/memberships/${membershipId}`, {
      data: { type: "memberships", id: membershipId, attributes: { role } },
    });
  }

  deleteMembership(tenant: Tenant, membershipId: string) {
    api.setTenant(tenant);
    return api.delete(`/memberships/${membershipId}`);
  }

  // --- Existing endpoints ---

  setActiveHousehold(tenant: Tenant, preferenceId: string, householdId: string) {
    api.setTenant(tenant);
    return api.patch<ApiResponse<Preference>>(`/preferences/${preferenceId}`, {
      data: {
        type: "preferences",
        id: preferenceId,
        relationships: {
          activeHousehold: {
            data: { type: "households", id: householdId },
          },
        },
      },
    });
  }
}

export const accountService = new AccountService();
