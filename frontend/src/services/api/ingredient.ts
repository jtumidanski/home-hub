import { BaseService } from "./base";
import type {
  CanonicalIngredientListItem,
  CanonicalIngredientDetail,
  CanonicalIngredientCreateAttributes,
  CanonicalIngredientUpdateAttributes,
} from "@/types/models/ingredient";
import type { Tenant } from "@/types/models/tenant";
import type { ApiListResponse } from "@/types/api/responses";
import { api } from "@/lib/api/client";

interface IngredientListParams {
  search?: string;
  page?: number;
  pageSize?: number;
}

export interface IngredientListResponse extends ApiListResponse<CanonicalIngredientListItem> {
  meta?: {
    total: number;
    page: number;
    pageSize: number;
  };
}

class IngredientService extends BaseService {
  constructor() {
    super("/ingredients");
  }

  listIngredients(tenant: Tenant, params?: IngredientListParams): Promise<IngredientListResponse> {
    this.setTenant(tenant);
    const query = new URLSearchParams();
    if (params?.search) query.set("search", params.search);
    if (params?.page) query.set("page[number]", String(params.page));
    if (params?.pageSize) query.set("page[size]", String(params.pageSize));
    const qs = query.toString();
    return api.get<IngredientListResponse>(`/ingredients${qs ? `?${qs}` : ""}`);
  }

  getIngredient(tenant: Tenant, id: string) {
    return this.getOne<CanonicalIngredientDetail>(tenant, `/ingredients/${id}`);
  }

  createIngredient(tenant: Tenant, attrs: CanonicalIngredientCreateAttributes) {
    return this.create<CanonicalIngredientDetail>(tenant, "/ingredients", {
      data: { type: "ingredients", attributes: attrs },
    });
  }

  updateIngredient(tenant: Tenant, id: string, attrs: CanonicalIngredientUpdateAttributes) {
    return this.update<CanonicalIngredientDetail>(tenant, `/ingredients/${id}`, {
      data: { type: "ingredients", id, attributes: attrs },
    });
  }

  deleteIngredient(tenant: Tenant, id: string) {
    return this.remove(tenant, `/ingredients/${id}`);
  }

  addAlias(tenant: Tenant, ingredientId: string, aliasName: string) {
    return this.create<CanonicalIngredientDetail>(tenant, `/ingredients/${ingredientId}/aliases`, {
      data: { type: "ingredient-aliases", attributes: { name: aliasName } },
    });
  }

  removeAlias(tenant: Tenant, ingredientId: string, aliasId: string) {
    return this.remove(tenant, `/ingredients/${ingredientId}/aliases/${aliasId}`);
  }

  getIngredientRecipes(tenant: Tenant, ingredientId: string, params?: { page?: number; pageSize?: number }) {
    this.setTenant(tenant);
    const query = new URLSearchParams();
    if (params?.page) query.set("page[number]", String(params.page));
    if (params?.pageSize) query.set("page[size]", String(params.pageSize));
    const qs = query.toString();
    return api.get<{ data: Array<{ recipeId: string; rawName: string }>; meta?: { total: number } }>(
      `/ingredients/${ingredientId}/recipes${qs ? `?${qs}` : ""}`,
    );
  }

  reassignAndDelete(tenant: Tenant, ingredientId: string, targetIngredientId: string) {
    return this.create<{ meta: { reassigned: number } }>(tenant, `/ingredients/${ingredientId}/reassign`, {
      data: { type: "ingredient-reassignments", attributes: { targetIngredientId } },
    });
  }
}

export const ingredientService = new IngredientService();
