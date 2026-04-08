import { BaseService } from "./base";
import { api } from "@/lib/api/client";
import type {
  ShoppingList,
  ShoppingItemCreateAttributes,
  ShoppingItemUpdateAttributes,
  ShoppingItemCheckAttributes,
  ShoppingItem,
} from "@/types/models/shopping";
import type { Tenant } from "@/types/models/tenant";
import type { ApiListResponse, ApiResponse } from "@/types/api/responses";

class ShoppingService extends BaseService {
  constructor() {
    super("/shopping");
  }

  listLists(tenant: Tenant, status: "active" | "archived" = "active"): Promise<ApiListResponse<ShoppingList>> {
    this.setTenant(tenant);
    return api.get<ApiListResponse<ShoppingList>>(`/shopping/lists?status=${status}`);
  }

  getListDetail(tenant: Tenant, id: string): Promise<ApiResponse<ShoppingList>> {
    return this.getOne<ShoppingList>(tenant, `/shopping/lists/${id}`);
  }

  createList(tenant: Tenant, name: string): Promise<ApiResponse<ShoppingList>> {
    return this.create<ShoppingList>(tenant, "/shopping/lists", {
      data: { type: "shopping-lists", attributes: { name } },
    });
  }

  updateList(tenant: Tenant, id: string, name: string): Promise<ApiResponse<ShoppingList>> {
    return this.update<ShoppingList>(tenant, `/shopping/lists/${id}`, {
      data: { type: "shopping-lists", id, attributes: { name } },
    });
  }

  deleteList(tenant: Tenant, id: string): Promise<void> {
    return this.remove(tenant, `/shopping/lists/${id}`);
  }

  archiveList(tenant: Tenant, id: string): Promise<ApiResponse<ShoppingList>> {
    this.setTenant(tenant);
    // Body must be a JSON:API resource — the backend handler is wired
    // through server.RegisterInputHandler[ArchiveRequest] which expects a
    // typed envelope. A bare {} will be rejected with "Could not parse
    // request body".
    return api.post<ApiResponse<ShoppingList>>(`/shopping/lists/${id}/archive`, {
      data: { type: "shopping-lists", id, attributes: {} },
    });
  }

  unarchiveList(tenant: Tenant, id: string): Promise<ApiResponse<ShoppingList>> {
    this.setTenant(tenant);
    return api.post<ApiResponse<ShoppingList>>(`/shopping/lists/${id}/unarchive`, {
      data: { type: "shopping-lists", id, attributes: {} },
    });
  }

  addItem(tenant: Tenant, listId: string, attrs: ShoppingItemCreateAttributes): Promise<ApiResponse<ShoppingItem>> {
    return this.create<ShoppingItem>(tenant, `/shopping/lists/${listId}/items`, {
      data: { type: "shopping-items", attributes: attrs },
    });
  }

  updateItem(tenant: Tenant, listId: string, itemId: string, attrs: ShoppingItemUpdateAttributes): Promise<ApiResponse<ShoppingItem>> {
    return this.update<ShoppingItem>(tenant, `/shopping/lists/${listId}/items/${itemId}`, {
      data: { type: "shopping-items", id: itemId, attributes: attrs },
    });
  }

  removeItem(tenant: Tenant, listId: string, itemId: string): Promise<void> {
    return this.remove(tenant, `/shopping/lists/${listId}/items/${itemId}`);
  }

  checkItem(tenant: Tenant, listId: string, itemId: string, attrs: ShoppingItemCheckAttributes): Promise<ApiResponse<ShoppingItem>> {
    return this.update<ShoppingItem>(tenant, `/shopping/lists/${listId}/items/${itemId}/check`, {
      data: { type: "shopping-items", id: itemId, attributes: attrs },
    });
  }

  uncheckAll(tenant: Tenant, listId: string): Promise<ApiResponse<ShoppingList>> {
    this.setTenant(tenant);
    return api.post<ApiResponse<ShoppingList>>(`/shopping/lists/${listId}/items/uncheck-all`, {});
  }

  importMealPlan(tenant: Tenant, listId: string, planId: string): Promise<ApiResponse<ShoppingList>> {
    this.setTenant(tenant);
    return api.post<ApiResponse<ShoppingList>>(`/shopping/lists/${listId}/import/meal-plan`, {
      data: { type: "shopping-list-imports", attributes: { plan_id: planId } },
    });
  }
}

export const shoppingService = new ShoppingService();
