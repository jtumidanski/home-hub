import { api, type RequestOptions } from "@/lib/api/client";
import type { ApiResponse, ApiListResponse } from "@/types/api/responses";
import type { Tenant } from "@/types/models/tenant";

export interface RetentionCategoryView {
  days: number;
  source: "default" | "household" | "user" | "tenant" | string;
}

export interface RetentionPolicyScope {
  id: string;
  categories: Record<string, RetentionCategoryView>;
}

export interface RetentionPolicyAttributes {
  household?: RetentionPolicyScope;
  user?: RetentionPolicyScope;
}

export interface RetentionPolicy {
  type: "retention-policies";
  id: string;
  attributes: RetentionPolicyAttributes;
}

export interface RetentionRunAttributes {
  service: string;
  category: string;
  scope: string;
  scope_id: string;
  trigger: "scheduled" | "manual";
  dry_run: boolean;
  scanned: number;
  deleted: number;
  started_at: string;
  finished_at?: string;
  error?: string;
}

export interface RetentionRun {
  type: "retention-runs";
  id: string;
  attributes: RetentionRunAttributes;
}

export interface RetentionPurgeAttributes {
  category: string;
  scope: "household" | "user";
  scope_id: string;
  status: string;
  scanned: number;
  deleted: number;
  dry_run: boolean;
}

export interface RetentionPurge {
  type: "retention-purges";
  id: string;
  attributes: RetentionPurgeAttributes;
}

class RetentionService {
  getPolicies(tenant: Tenant, options?: RequestOptions) {
    api.setTenant(tenant);
    return api.get<ApiResponse<RetentionPolicy>>("/retention-policies", options);
  }

  patchHousehold(tenant: Tenant, householdId: string, categories: Record<string, number | null>) {
    api.setTenant(tenant);
    return api.patch<ApiResponse<RetentionPolicy>>(
      `/retention-policies/household/${householdId}`,
      {
        data: {
          type: "retention-policies",
          attributes: { categories },
        },
      },
    );
  }

  patchUser(tenant: Tenant, categories: Record<string, number | null>) {
    api.setTenant(tenant);
    return api.patch<ApiResponse<RetentionPolicy>>("/retention-policies/user", {
      data: {
        type: "retention-policies",
        attributes: { categories },
      },
    });
  }

  purge(tenant: Tenant, category: string, scope: "household" | "user", dryRun: boolean = false) {
    api.setTenant(tenant);
    return api.post<ApiResponse<RetentionPurge>>("/retention-policies/purge", {
      data: {
        type: "retention-purges",
        attributes: { category, scope, dry_run: dryRun },
      },
    });
  }

  listRuns(tenant: Tenant, params: { category?: string; trigger?: string; limit?: number } = {}) {
    api.setTenant(tenant);
    const search = new URLSearchParams();
    if (params.category) search.set("category", params.category);
    if (params.trigger) search.set("trigger", params.trigger);
    if (params.limit) search.set("limit", String(params.limit));
    const qs = search.toString();
    return api.get<ApiListResponse<RetentionRun>>(`/retention-runs${qs ? `?${qs}` : ""}`);
  }
}

export const retentionService = new RetentionService();
