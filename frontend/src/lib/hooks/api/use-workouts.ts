import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { workoutService } from "@/services/api/workout";
import { useTenant } from "@/context/tenant-context";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";
import type {
  ExerciseDefaults,
  PerformanceStatus,
  WeightType,
  WeightUnit,
  WorkoutKind,
} from "@/types/models/workout";

// Query-key registry. The household is included for parity with other domain
// hooks even though workout-service has no household scope — it keeps the
// tenant cache invalidation behavior consistent with the rest of the app.
export const workoutKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["workouts", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  themes: (tenant: Tenant | null, household: Household | null) =>
    [...workoutKeys.all(tenant, household), "themes"] as const,
  regions: (tenant: Tenant | null, household: Household | null) =>
    [...workoutKeys.all(tenant, household), "regions"] as const,
  exercises: (tenant: Tenant | null, household: Household | null, themeId?: string, regionId?: string) =>
    [...workoutKeys.all(tenant, household), "exercises", themeId ?? "all", regionId ?? "all"] as const,
  week: (tenant: Tenant | null, household: Household | null, weekStart: string) =>
    [...workoutKeys.all(tenant, household), "week", weekStart] as const,
  today: (tenant: Tenant | null, household: Household | null) =>
    [...workoutKeys.all(tenant, household), "today"] as const,
  summary: (tenant: Tenant | null, household: Household | null, weekStart: string) =>
    [...workoutKeys.all(tenant, household), "summary", weekStart] as const,
};

// --- themes -----------------------------------------------------------------

export function useWorkoutThemes() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: workoutKeys.themes(tenant, household),
    queryFn: () => workoutService.listThemes(tenant!),
    enabled: !!tenant?.id,
    staleTime: 60 * 1000,
  });
}

export function useCreateWorkoutTheme() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: { name: string; sortOrder?: number }) => workoutService.createTheme(tenant!, attrs),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.themes(tenant, household) }),
  });
}

export function useUpdateWorkoutTheme() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: { name?: string; sortOrder?: number } }) =>
      workoutService.updateTheme(tenant!, id, attrs),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.themes(tenant, household) }),
  });
}

export function useDeleteWorkoutTheme() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => workoutService.deleteTheme(tenant!, id),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.themes(tenant, household) }),
  });
}

// --- regions ----------------------------------------------------------------

export function useWorkoutRegions() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: workoutKeys.regions(tenant, household),
    queryFn: () => workoutService.listRegions(tenant!),
    enabled: !!tenant?.id,
    staleTime: 60 * 1000,
  });
}

export function useCreateWorkoutRegion() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: { name: string; sortOrder?: number }) => workoutService.createRegion(tenant!, attrs),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.regions(tenant, household) }),
  });
}

export function useUpdateWorkoutRegion() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: { name?: string; sortOrder?: number } }) =>
      workoutService.updateRegion(tenant!, id, attrs),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.regions(tenant, household) }),
  });
}

export function useDeleteWorkoutRegion() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => workoutService.deleteRegion(tenant!, id),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.regions(tenant, household) }),
  });
}

// --- exercises --------------------------------------------------------------

export function useWorkoutExercises(opts?: { themeId?: string; regionId?: string }) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: workoutKeys.exercises(tenant, household, opts?.themeId, opts?.regionId),
    queryFn: () => workoutService.listExercises(tenant!, opts),
    enabled: !!tenant?.id,
    staleTime: 60 * 1000,
  });
}

export interface CreateExerciseAttrs {
  name: string;
  kind: WorkoutKind;
  weightType?: WeightType;
  themeId: string;
  regionId: string;
  secondaryRegionIds?: string[];
  defaults: ExerciseDefaults;
  notes?: string | null;
}

export function useCreateWorkoutExercise() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: CreateExerciseAttrs) => workoutService.createExercise(tenant!, attrs),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.all(tenant, household) }),
  });
}

export function useUpdateWorkoutExercise() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: Partial<CreateExerciseAttrs> }) =>
      workoutService.updateExercise(tenant!, id, attrs),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.all(tenant, household) }),
  });
}

export function useDeleteWorkoutExercise() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => workoutService.deleteExercise(tenant!, id),
    onSuccess: () => qc.invalidateQueries({ queryKey: workoutKeys.all(tenant, household) }),
  });
}

// --- weeks & today & summary ------------------------------------------------

export function useWorkoutWeek(weekStart: string) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: workoutKeys.week(tenant, household, weekStart),
    queryFn: () => workoutService.getWeek(tenant!, weekStart),
    enabled: !!tenant?.id && !!weekStart,
    retry: false, // 404 is the empty-week signal — surface it immediately
    staleTime: 30 * 1000,
  });
}

export function usePatchWorkoutWeek() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ weekStart, restDayFlags }: { weekStart: string; restDayFlags: number[] }) =>
      workoutService.patchWeek(tenant!, weekStart, restDayFlags),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: workoutKeys.week(tenant, household, vars.weekStart) });
    },
  });
}

export function useCopyWorkoutWeek() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ weekStart, mode }: { weekStart: string; mode: "planned" | "actual" }) =>
      workoutService.copyWeek(tenant!, weekStart, mode),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: workoutKeys.week(tenant, household, vars.weekStart) });
    },
  });
}

export function useAddPlannedItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({
      weekStart,
      attrs,
    }: {
      weekStart: string;
      attrs: {
        exerciseId: string;
        dayOfWeek: number;
        position?: number;
        planned?: Record<string, unknown>;
        notes?: string | null;
      };
    }) => workoutService.addPlannedItem(tenant!, weekStart, attrs),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: workoutKeys.week(tenant, household, vars.weekStart) });
      qc.invalidateQueries({ queryKey: workoutKeys.today(tenant, household) });
    },
  });
}

export function useUpdatePlannedItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({
      weekStart,
      itemId,
      attrs,
    }: {
      weekStart: string;
      itemId: string;
      attrs: Record<string, unknown>;
    }) => workoutService.updatePlannedItem(tenant!, weekStart, itemId, attrs),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: workoutKeys.week(tenant, household, vars.weekStart) });
      qc.invalidateQueries({ queryKey: workoutKeys.today(tenant, household) });
    },
  });
}

export function useDeletePlannedItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ weekStart, itemId }: { weekStart: string; itemId: string }) =>
      workoutService.deletePlannedItem(tenant!, weekStart, itemId),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: workoutKeys.week(tenant, household, vars.weekStart) });
      qc.invalidateQueries({ queryKey: workoutKeys.today(tenant, household) });
    },
  });
}

export function usePatchPerformance() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({
      weekStart,
      itemId,
      attrs,
    }: {
      weekStart: string;
      itemId: string;
      attrs: {
        status?: PerformanceStatus;
        weightUnit?: WeightUnit;
        actuals?: Record<string, unknown>;
        notes?: string | null;
      };
    }) => workoutService.patchPerformance(tenant!, weekStart, itemId, attrs),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: workoutKeys.week(tenant, household, vars.weekStart) });
      qc.invalidateQueries({ queryKey: workoutKeys.today(tenant, household) });
    },
  });
}

export function useWorkoutToday() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: workoutKeys.today(tenant, household),
    queryFn: () => workoutService.getToday(tenant!),
    enabled: !!tenant?.id,
    staleTime: 30 * 1000,
  });
}

export function useWorkoutWeekSummary(weekStart: string) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: workoutKeys.summary(tenant, household, weekStart),
    queryFn: () => workoutService.getWeekSummary(tenant!, weekStart),
    enabled: !!tenant?.id && !!weekStart,
    staleTime: 30 * 1000,
  });
}
