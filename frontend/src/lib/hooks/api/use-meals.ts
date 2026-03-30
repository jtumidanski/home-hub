import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { mealsService } from "@/services/api/meals";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import type {
  PlanCreateAttributes,
  PlanUpdateAttributes,
  PlanDuplicateAttributes,
  PlanItemCreateAttributes,
  PlanItemUpdateAttributes,
} from "@/types/models/meal-plan";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

// --- Key factory ---

export const mealKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["meals", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  plans: (tenant: Tenant | null, household: Household | null) =>
    [...mealKeys.all(tenant, household), "plans"] as const,
  planDetail: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...mealKeys.plans(tenant, household), id] as const,
  ingredients: (tenant: Tenant | null, household: Household | null, planId: string) =>
    [...mealKeys.planDetail(tenant, household, planId), "ingredients"] as const,
};

// --- Query hooks ---

interface UsePlansParams {
  starts_on?: string;
  page?: number;
  pageSize?: number;
}

export function usePlans(params?: UsePlansParams) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: [...mealKeys.plans(tenant, household), params] as const,
    queryFn: () => mealsService.listPlans(tenant!, params),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function usePlan(id: string | null) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: mealKeys.planDetail(tenant, household, id ?? ""),
    queryFn: () => mealsService.getPlan(tenant!, id!),
    enabled: !!tenant?.id && !!household?.id && !!id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function usePlanIngredients(planId: string | null) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: mealKeys.ingredients(tenant, household, planId ?? ""),
    queryFn: () => mealsService.getIngredients(tenant!, planId!),
    enabled: !!tenant?.id && !!household?.id && !!planId,
    staleTime: 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreatePlan() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: PlanCreateAttributes) =>
      mealsService.createPlan(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: mealKeys.plans(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create plan"));
    },
  });
}

export function useUpdatePlan() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: PlanUpdateAttributes }) =>
      mealsService.updatePlan(tenant!, id, attrs),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: mealKeys.plans(tenant, household) });
      qc.invalidateQueries({ queryKey: mealKeys.planDetail(tenant, household, variables.id) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update plan"));
    },
  });
}

export function useLockPlan() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => mealsService.lockPlan(tenant!, id),
    onSettled: (_data, _err, id) => {
      qc.invalidateQueries({ queryKey: mealKeys.plans(tenant, household) });
      qc.invalidateQueries({ queryKey: mealKeys.planDetail(tenant, household, id) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to lock plan"));
    },
  });
}

export function useUnlockPlan() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => mealsService.unlockPlan(tenant!, id),
    onSettled: (_data, _err, id) => {
      qc.invalidateQueries({ queryKey: mealKeys.plans(tenant, household) });
      qc.invalidateQueries({ queryKey: mealKeys.planDetail(tenant, household, id) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to unlock plan"));
    },
  });
}

export function useDuplicatePlan() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: PlanDuplicateAttributes }) =>
      mealsService.duplicatePlan(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: mealKeys.plans(tenant, household) });
    },
    onSuccess: () => {
      toast.success("Plan duplicated successfully");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to duplicate plan"));
    },
  });
}

export function useAddPlanItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ planId, attrs }: { planId: string; attrs: PlanItemCreateAttributes }) =>
      mealsService.addItem(tenant!, planId, attrs),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: mealKeys.planDetail(tenant, household, variables.planId) });
      qc.invalidateQueries({ queryKey: mealKeys.ingredients(tenant, household, variables.planId) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to add item"));
    },
  });
}

export function useUpdatePlanItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ planId, itemId, attrs }: { planId: string; itemId: string; attrs: PlanItemUpdateAttributes }) =>
      mealsService.updateItem(tenant!, planId, itemId, attrs),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: mealKeys.planDetail(tenant, household, variables.planId) });
      qc.invalidateQueries({ queryKey: mealKeys.ingredients(tenant, household, variables.planId) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update item"));
    },
  });
}

export function useRemovePlanItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ planId, itemId }: { planId: string; itemId: string }) =>
      mealsService.removeItem(tenant!, planId, itemId),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: mealKeys.planDetail(tenant, household, variables.planId) });
      qc.invalidateQueries({ queryKey: mealKeys.ingredients(tenant, household, variables.planId) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to remove item"));
    },
  });
}

export function useExportMarkdown() {
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (planId: string) => mealsService.exportMarkdown(tenant!, planId),
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to export plan"));
    },
  });
}
