import { BaseService } from "./base";
import { api } from "@/lib/api/client";
import type {
  PlanListItem,
  PlanDetail,
  PlanCreateAttributes,
  PlanUpdateAttributes,
  PlanDuplicateAttributes,
  PlanItemResponse,
  PlanItemCreateAttributes,
  PlanItemUpdateAttributes,
  PlanIngredient,
} from "@/types/models/meal-plan";
import type { Tenant } from "@/types/models/tenant";
import type { ApiListResponse, ApiResponse } from "@/types/api/responses";

export interface PlanListResponse extends ApiListResponse<PlanListItem> {
  meta?: {
    total: number;
    page: number;
    pageSize: number;
  };
}

interface PlanListParams {
  starts_on?: string;
  page?: number;
  pageSize?: number;
}

class MealsService extends BaseService {
  constructor() {
    super("/meals");
  }

  // --- Plans ---

  listPlans(tenant: Tenant, params?: PlanListParams): Promise<PlanListResponse> {
    this.setTenant(tenant);
    const query = new URLSearchParams();
    if (params?.starts_on) query.set("starts_on", params.starts_on);
    if (params?.page) query.set("page[number]", String(params.page));
    if (params?.pageSize) query.set("page[size]", String(params.pageSize));
    const qs = query.toString();
    return api.get<PlanListResponse>(`/meals/plans${qs ? `?${qs}` : ""}`);
  }

  getPlan(tenant: Tenant, id: string) {
    return this.getOne<PlanDetail>(tenant, `/meals/plans/${id}`);
  }

  createPlan(tenant: Tenant, attrs: PlanCreateAttributes) {
    return this.create<PlanDetail>(tenant, "/meals/plans", {
      data: { type: "plans", attributes: attrs },
    });
  }

  updatePlan(tenant: Tenant, id: string, attrs: PlanUpdateAttributes) {
    return this.update<PlanDetail>(tenant, `/meals/plans/${id}`, {
      data: { type: "plans", id, attributes: attrs },
    });
  }

  lockPlan(tenant: Tenant, id: string) {
    this.setTenant(tenant);
    return api.post<ApiResponse<PlanDetail>>(`/meals/plans/${id}/lock`, {});
  }

  unlockPlan(tenant: Tenant, id: string) {
    this.setTenant(tenant);
    return api.post<ApiResponse<PlanDetail>>(`/meals/plans/${id}/unlock`, {});
  }

  duplicatePlan(tenant: Tenant, id: string, attrs: PlanDuplicateAttributes) {
    return this.create<PlanDetail>(tenant, `/meals/plans/${id}/duplicate`, {
      data: { type: "plans", attributes: attrs },
    });
  }

  // --- Plan Items ---

  addItem(tenant: Tenant, planId: string, attrs: PlanItemCreateAttributes) {
    return this.create<PlanItemResponse>(tenant, `/meals/plans/${planId}/items`, {
      data: { type: "plan-items", attributes: attrs },
    });
  }

  updateItem(tenant: Tenant, planId: string, itemId: string, attrs: PlanItemUpdateAttributes) {
    return this.update<PlanItemResponse>(tenant, `/meals/plans/${planId}/items/${itemId}`, {
      data: { type: "plan-items", id: itemId, attributes: attrs },
    });
  }

  removeItem(tenant: Tenant, planId: string, itemId: string) {
    return this.remove(tenant, `/meals/plans/${planId}/items/${itemId}`);
  }

  // --- Export ---

  exportMarkdown(tenant: Tenant, planId: string): Promise<string> {
    this.setTenant(tenant);
    return api.getText(`/meals/plans/${planId}/export/markdown`, {
      headers: { Accept: "text/markdown" },
    });
  }

  getIngredients(tenant: Tenant, planId: string) {
    return this.getList<PlanIngredient>(tenant, `/meals/plans/${planId}/ingredients`);
  }
}

export const mealsService = new MealsService();
