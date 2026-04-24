import { api } from "@/lib/api/client";
import { BaseService } from "./base";
import type { Layout } from "@/lib/dashboard/schema";
import type {
  Dashboard,
  DashboardCreateAttributes,
  DashboardOrderEntry,
  DashboardUpdateAttributes,
} from "@/types/models/dashboard";
import type { ApiListResponse, ApiResponse } from "@/types/api/responses";

class DashboardService extends BaseService {
  constructor() {
    super("/dashboards");
  }

  listDashboards(tenant: { id: string }) {
    return this.getList<Dashboard>(tenant, "/dashboards");
  }

  getDashboard(tenant: { id: string }, id: string) {
    return this.getOne<Dashboard>(tenant, `/dashboards/${id}`);
  }

  createDashboard(tenant: { id: string }, attrs: DashboardCreateAttributes) {
    return this.create<Dashboard>(tenant, "/dashboards", {
      data: {
        type: "dashboards",
        attributes: attrs,
      },
    });
  }

  updateDashboard(tenant: { id: string }, id: string, attrs: DashboardUpdateAttributes) {
    return this.update<Dashboard>(tenant, `/dashboards/${id}`, {
      data: {
        type: "dashboards",
        id,
        attributes: attrs,
      },
    });
  }

  deleteDashboard(tenant: { id: string }, id: string) {
    return this.remove(tenant, `/dashboards/${id}`);
  }

  /**
   * Bulk reorder. Plain JSON body, NOT JSON:API — the endpoint takes
   * `{ data: [{ id, sortOrder }] }`.
   */
  reorderDashboards(tenant: { id: string }, entries: DashboardOrderEntry[]): Promise<ApiListResponse<Dashboard>> {
    this.setTenant(tenant);
    return api.patch<ApiListResponse<Dashboard>>(`/dashboards/order`, { data: entries });
  }

  promoteDashboard(tenant: { id: string }, id: string) {
    return this.create<Dashboard>(tenant, `/dashboards/${id}/promote`, {});
  }

  copyDashboardToMine(tenant: { id: string }, id: string) {
    return this.create<Dashboard>(tenant, `/dashboards/${id}/copy-to-mine`, {});
  }

  /**
   * Idempotent seed. On first call the backend returns `201` with a
   * single-resource body; on subsequent calls it returns `200` with a list.
   * Callers should accept either shape.
   */
  seedDashboard(
    tenant: { id: string },
    name: string,
    layout: Layout,
  ): Promise<ApiResponse<Dashboard> | ApiListResponse<Dashboard>> {
    this.setTenant(tenant);
    return api.post<ApiResponse<Dashboard> | ApiListResponse<Dashboard>>(`/dashboards/seed`, {
      data: {
        type: "dashboards",
        attributes: { name, layout },
      },
    });
  }
}

export const dashboardService = new DashboardService();
