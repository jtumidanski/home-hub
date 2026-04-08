import { api } from "@/lib/api/client";
import type {
  WishListItem,
  WishListItemCreateAttributes,
  WishListItemUpdateAttributes,
} from "@/types/models/wish-list";
import type { Tenant } from "@/types/models/tenant";
import type { ApiListResponse, ApiResponse } from "@/types/api/responses";

class WishListService {
  private setTenant(tenant: Tenant) {
    api.setTenant(tenant);
  }

  list(tenant: Tenant): Promise<ApiListResponse<WishListItem>> {
    this.setTenant(tenant);
    return api.get<ApiListResponse<WishListItem>>("/shopping/wish-list/items");
  }

  create(
    tenant: Tenant,
    attrs: WishListItemCreateAttributes,
  ): Promise<ApiResponse<WishListItem>> {
    this.setTenant(tenant);
    return api.post<ApiResponse<WishListItem>>("/shopping/wish-list/items", {
      data: { type: "wish-items", attributes: attrs },
    });
  }

  update(
    tenant: Tenant,
    id: string,
    attrs: WishListItemUpdateAttributes,
  ): Promise<ApiResponse<WishListItem>> {
    this.setTenant(tenant);
    return api.patch<ApiResponse<WishListItem>>(`/shopping/wish-list/items/${id}`, {
      data: { type: "wish-items", id, attributes: attrs },
    });
  }

  remove(tenant: Tenant, id: string): Promise<void> {
    this.setTenant(tenant);
    return api.delete(`/shopping/wish-list/items/${id}`);
  }

  vote(tenant: Tenant, id: string): Promise<ApiResponse<WishListItem>> {
    this.setTenant(tenant);
    return api.post<ApiResponse<WishListItem>>(`/shopping/wish-list/items/${id}/vote`, {
      data: { type: "wish-items", id, attributes: {} },
    });
  }
}

export const wishListService = new WishListService();
