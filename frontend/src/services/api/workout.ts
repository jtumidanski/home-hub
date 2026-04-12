import { api } from "@/lib/api/client";
import { BaseService } from "./base";
import type {
  Exercise,
  PerformanceStatus,
  Region,
  SummaryDocument,
  Theme,
  TodayDocument,
  WeekDocument,
  WeightType,
  WeightUnit,
  WorkoutKind,
} from "@/types/models/workout";
import type { ApiListResponse, ApiResponse } from "@/types/api/responses";

class WorkoutService extends BaseService {
  constructor() {
    super("/workouts");
  }

  // --- themes ----------------------------------------------------------------

  listThemes(tenant: { id: string }) {
    return this.getList<Theme>(tenant, "/workouts/themes");
  }
  createTheme(tenant: { id: string }, attrs: { name: string; sortOrder?: number }) {
    return this.create<Theme>(tenant, "/workouts/themes", {
      data: { type: "themes", attributes: attrs },
    });
  }
  updateTheme(tenant: { id: string }, id: string, attrs: { name?: string; sortOrder?: number }) {
    return this.update<Theme>(tenant, `/workouts/themes/${id}`, {
      data: { type: "themes", id, attributes: attrs },
    });
  }
  deleteTheme(tenant: { id: string }, id: string) {
    return this.remove(tenant, `/workouts/themes/${id}`);
  }

  // --- regions ---------------------------------------------------------------

  listRegions(tenant: { id: string }) {
    return this.getList<Region>(tenant, "/workouts/regions");
  }
  createRegion(tenant: { id: string }, attrs: { name: string; sortOrder?: number }) {
    return this.create<Region>(tenant, "/workouts/regions", {
      data: { type: "regions", attributes: attrs },
    });
  }
  updateRegion(tenant: { id: string }, id: string, attrs: { name?: string; sortOrder?: number }) {
    return this.update<Region>(tenant, `/workouts/regions/${id}`, {
      data: { type: "regions", id, attributes: attrs },
    });
  }
  deleteRegion(tenant: { id: string }, id: string) {
    return this.remove(tenant, `/workouts/regions/${id}`);
  }

  // --- exercises -------------------------------------------------------------

  listExercises(tenant: { id: string }, opts?: { themeId?: string; regionId?: string }) {
    this.setTenant(tenant);
    const params = new URLSearchParams();
    if (opts?.themeId) params.set("themeId", opts.themeId);
    if (opts?.regionId) params.set("regionId", opts.regionId);
    const qs = params.toString();
    return api.get<ApiListResponse<Exercise>>(`/workouts/exercises${qs ? `?${qs}` : ""}`);
  }
  createExercise(
    tenant: { id: string },
    attrs: {
      name: string;
      kind: WorkoutKind;
      weightType?: WeightType;
      themeId: string;
      regionId: string;
      secondaryRegionIds?: string[];
      defaultSets?: number | null;
      defaultReps?: number | null;
      defaultWeight?: number | null;
      defaultWeightUnit?: WeightUnit | null;
      defaultDurationSeconds?: number | null;
      defaultDistance?: number | null;
      defaultDistanceUnit?: string | null;
      notes?: string | null;
    }
  ) {
    return this.create<Exercise>(tenant, "/workouts/exercises", {
      data: { type: "exercises", attributes: attrs },
    });
  }
  updateExercise(
    tenant: { id: string },
    id: string,
    attrs: {
      name?: string;
      themeId?: string;
      regionId?: string;
      secondaryRegionIds?: string[];
      defaultSets?: number | null;
      defaultReps?: number | null;
      defaultWeight?: number | null;
      defaultWeightUnit?: WeightUnit | null;
      defaultDurationSeconds?: number | null;
      defaultDistance?: number | null;
      defaultDistanceUnit?: string | null;
      notes?: string | null;
    }
  ) {
    return this.update<Exercise>(tenant, `/workouts/exercises/${id}`, {
      data: { type: "exercises", id, attributes: attrs },
    });
  }
  deleteExercise(tenant: { id: string }, id: string) {
    return this.remove(tenant, `/workouts/exercises/${id}`);
  }

  // --- weeks -----------------------------------------------------------------

  getWeek(tenant: { id: string }, weekStart: string): Promise<WeekDocument> {
    this.setTenant(tenant);
    return api.get<WeekDocument>(`/workouts/weeks/${weekStart}`);
  }
  patchWeek(tenant: { id: string }, weekStart: string, restDayFlags: number[]): Promise<WeekDocument> {
    this.setTenant(tenant);
    return api.patch<WeekDocument>(`/workouts/weeks/${weekStart}`, {
      data: { type: "weeks", attributes: { restDayFlags } },
    });
  }
  copyWeek(tenant: { id: string }, weekStart: string, mode: "planned" | "actual"): Promise<WeekDocument> {
    this.setTenant(tenant);
    return api.post<WeekDocument>(`/workouts/weeks/${weekStart}/copy`, {
      data: { type: "weeks", attributes: { mode } },
    });
  }

  // --- planned items ---------------------------------------------------------

  addPlannedItem(
    tenant: { id: string },
    weekStart: string,
    attrs: {
      exerciseId: string;
      dayOfWeek: number;
      position?: number;
      planned?: Record<string, unknown>;
      notes?: string | null;
    }
  ): Promise<WeekDocument> {
    this.setTenant(tenant);
    return api.post<WeekDocument>(`/workouts/weeks/${weekStart}/items`, {
      data: { type: "planned-items", attributes: attrs },
    });
  }
  updatePlannedItem(
    tenant: { id: string },
    weekStart: string,
    itemId: string,
    attrs: Record<string, unknown>
  ): Promise<WeekDocument> {
    this.setTenant(tenant);
    return api.patch<WeekDocument>(`/workouts/weeks/${weekStart}/items/${itemId}`, {
      data: { type: "planned-items", attributes: attrs },
    });
  }
  deletePlannedItem(tenant: { id: string }, weekStart: string, itemId: string): Promise<void> {
    this.setTenant(tenant);
    return api.delete(`/workouts/weeks/${weekStart}/items/${itemId}`);
  }
  reorderPlannedItems(
    tenant: { id: string },
    weekStart: string,
    items: Array<{ itemId: string; dayOfWeek: number; position: number }>
  ): Promise<WeekDocument> {
    this.setTenant(tenant);
    return api.post<WeekDocument>(`/workouts/weeks/${weekStart}/items/reorder`, {
      data: { type: "planned-items", attributes: { items } },
    });
  }

  // --- performance -----------------------------------------------------------

  patchPerformance(
    tenant: { id: string },
    weekStart: string,
    itemId: string,
    attrs: {
      status?: PerformanceStatus;
      weightUnit?: WeightUnit;
      actualSets?: number | null;
      actualReps?: number | null;
      actualWeight?: number | null;
      actualDurationSeconds?: number | null;
      actualDistance?: number | null;
      actualDistanceUnit?: string | null;
      notes?: string | null;
    }
  ): Promise<ApiResponse<unknown>> {
    this.setTenant(tenant);
    return api.patch<ApiResponse<unknown>>(`/workouts/weeks/${weekStart}/items/${itemId}/performance`, {
      data: { type: "performances", attributes: attrs },
    });
  }
  putPerformanceSets(
    tenant: { id: string },
    weekStart: string,
    itemId: string,
    weightUnit: WeightUnit,
    sets: Array<{ reps: number; weight: number }>
  ): Promise<ApiResponse<unknown>> {
    this.setTenant(tenant);
    return api.put<ApiResponse<unknown>>(`/workouts/weeks/${weekStart}/items/${itemId}/performance/sets`, {
      data: { type: "performances", attributes: { weightUnit, sets } },
    });
  }
  collapsePerformanceSets(tenant: { id: string }, weekStart: string, itemId: string): Promise<void> {
    this.setTenant(tenant);
    return api.delete(`/workouts/weeks/${weekStart}/items/${itemId}/performance/sets`);
  }

  // --- composite reads -------------------------------------------------------

  getToday(tenant: { id: string }): Promise<TodayDocument> {
    this.setTenant(tenant);
    return api.get<TodayDocument>("/workouts/today");
  }
  getWeekSummary(tenant: { id: string }, weekStart: string): Promise<SummaryDocument> {
    this.setTenant(tenant);
    return api.get<SummaryDocument>(`/workouts/weeks/${weekStart}/summary`);
  }
}

export const workoutService = new WorkoutService();
