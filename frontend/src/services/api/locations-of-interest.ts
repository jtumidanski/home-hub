import { api } from "@/lib/api/client";
import type {
  LocationOfInterest,
  LocationOfInterestCreateAttributes,
  LocationOfInterestUpdateAttributes,
} from "@/types/models/location-of-interest";
import type { Tenant } from "@/types/models/tenant";
import type { ApiListResponse, ApiResponse } from "@/types/api/responses";

class LocationsOfInterestService {
  private setTenant(tenant: Tenant) {
    api.setTenant(tenant);
  }

  list(tenant: Tenant): Promise<ApiListResponse<LocationOfInterest>> {
    this.setTenant(tenant);
    return api.get<ApiListResponse<LocationOfInterest>>("/locations-of-interest");
  }

  create(
    tenant: Tenant,
    attrs: LocationOfInterestCreateAttributes,
  ): Promise<ApiResponse<LocationOfInterest>> {
    this.setTenant(tenant);
    return api.post<ApiResponse<LocationOfInterest>>("/locations-of-interest", {
      data: { type: "location-of-interest", attributes: attrs },
    });
  }

  update(
    tenant: Tenant,
    id: string,
    attrs: LocationOfInterestUpdateAttributes,
  ): Promise<ApiResponse<LocationOfInterest>> {
    this.setTenant(tenant);
    return api.patch<ApiResponse<LocationOfInterest>>(
      `/locations-of-interest/${id}`,
      {
        data: { type: "location-of-interest", id, attributes: attrs },
      },
    );
  }

  remove(tenant: Tenant, id: string): Promise<void> {
    this.setTenant(tenant);
    return api.delete(`/locations-of-interest/${id}`);
  }
}

export const locationsOfInterestService = new LocationsOfInterestService();
