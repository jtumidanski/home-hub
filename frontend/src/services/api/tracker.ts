import { api } from "@/lib/api/client";
import { BaseService } from "./base";
import type {
  Tracker,
  TrackerEntry,
  MonthSummaryResponse,
  MonthReportResponse,
  TodayResponse,
} from "@/types/models/tracker";
import type { ApiResponse, ApiListResponse } from "@/types/api/responses";

class TrackerService extends BaseService {
  constructor() {
    super("/trackers");
  }

  getTrackers(tenant: { id: string }) {
    return this.getList<Tracker>(tenant);
  }

  getTracker(tenant: { id: string }, id: string) {
    return this.getOne<Tracker>(tenant, `/trackers/${id}`);
  }

  createTracker(
    tenant: { id: string },
    attrs: {
      name: string;
      scale_type: string;
      scale_config: { min: number; max: number } | null;
      schedule: number[];
      color: string;
      sort_order?: number;
    }
  ) {
    return this.create<Tracker>(tenant, "/trackers", {
      data: {
        type: "trackers",
        attributes: attrs,
      },
    });
  }

  updateTracker(
    tenant: { id: string },
    id: string,
    attrs: {
      name?: string;
      color?: string;
      schedule?: number[];
      sort_order?: number;
      scale_config?: { min: number; max: number };
    }
  ) {
    return this.update<Tracker>(tenant, `/trackers/${id}`, {
      data: {
        type: "trackers",
        id,
        attributes: attrs,
      },
    });
  }

  deleteTracker(tenant: { id: string }, id: string) {
    return this.remove(tenant, `/trackers/${id}`);
  }

  getToday(tenant: { id: string }, date: string): Promise<TodayResponse> {
    this.setTenant(tenant);
    return api.get<TodayResponse>(`/trackers/today?date=${encodeURIComponent(date)}`);
  }

  putEntry(
    tenant: { id: string },
    itemId: string,
    date: string,
    today: string,
    value: unknown,
    note?: string | null
  ): Promise<ApiResponse<TrackerEntry>> {
    this.setTenant(tenant);
    return api.put<ApiResponse<TrackerEntry>>(
      `/trackers/${itemId}/entries/${date}?today=${encodeURIComponent(today)}`,
      {
        data: {
          type: "tracker-entries",
          attributes: {
            value,
            ...(note !== undefined ? { note } : {}),
          },
        },
      },
    );
  }

  deleteEntry(tenant: { id: string }, itemId: string, date: string) {
    return this.remove(tenant, `/trackers/${itemId}/entries/${date}`);
  }

  skipEntry(
    tenant: { id: string },
    itemId: string,
    date: string,
    today: string,
  ): Promise<ApiResponse<TrackerEntry>> {
    this.setTenant(tenant);
    return api.put<ApiResponse<TrackerEntry>>(
      `/trackers/${itemId}/entries/${date}/skip?today=${encodeURIComponent(today)}`,
      {},
    );
  }

  removeSkip(tenant: { id: string }, itemId: string, date: string) {
    return this.remove(tenant, `/trackers/${itemId}/entries/${date}/skip`);
  }

  getEntriesByMonth(tenant: { id: string }, month: string): Promise<ApiListResponse<TrackerEntry>> {
    this.setTenant(tenant);
    return api.get<ApiListResponse<TrackerEntry>>(`/trackers/entries?month=${month}`);
  }

  getMonthSummary(tenant: { id: string }, month: string, today: string): Promise<MonthSummaryResponse> {
    this.setTenant(tenant);
    return api.get<MonthSummaryResponse>(
      `/trackers/months/${month}?today=${encodeURIComponent(today)}`,
    );
  }

  getMonthReport(tenant: { id: string }, month: string, today: string): Promise<MonthReportResponse> {
    this.setTenant(tenant);
    return api.get<MonthReportResponse>(
      `/trackers/months/${month}/report?today=${encodeURIComponent(today)}`,
    );
  }
}

export const trackerService = new TrackerService();
