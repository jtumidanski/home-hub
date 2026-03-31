import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ingredientService } from "@/services/api/ingredient";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import { ingredientKeys, categoryKeys } from "./query-keys";
import type {
  CanonicalIngredientCreateAttributes,
  CanonicalIngredientUpdateAttributes,
} from "@/types/models/ingredient";
export { ingredientKeys } from "./query-keys";

// --- Query hooks ---

interface UseIngredientsParams {
  search?: string;
  page?: number;
  pageSize?: number;
  categoryId?: string;
}

export function useIngredients(params?: UseIngredientsParams) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: [...ingredientKeys.lists(tenant, household), params] as const,
    queryFn: () => ingredientService.listIngredients(tenant!, params),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useIngredient(id: string) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: ingredientKeys.detail(tenant, household, id),
    queryFn: () => ingredientService.getIngredient(tenant!, id),
    enabled: !!tenant?.id && !!household?.id && !!id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useIngredientRecipes(id: string, params?: { page?: number; pageSize?: number }) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: [...ingredientKeys.recipes(tenant, household, id), params] as const,
    queryFn: () => ingredientService.getIngredientRecipes(tenant!, id, params),
    enabled: !!tenant?.id && !!household?.id && !!id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateIngredient() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: CanonicalIngredientCreateAttributes) =>
      ingredientService.createIngredient(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ingredientKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create ingredient"));
    },
  });
}

export function useUpdateIngredient() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: CanonicalIngredientUpdateAttributes }) =>
      ingredientService.updateIngredient(tenant!, id, attrs),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: ingredientKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: ingredientKeys.detail(tenant, household, variables.id) });
      qc.invalidateQueries({ queryKey: categoryKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update ingredient"));
    },
  });
}

export function useDeleteIngredient() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => ingredientService.deleteIngredient(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ingredientKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: categoryKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to delete ingredient"));
    },
  });
}

export function useReassignIngredient() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, targetId }: { id: string; targetId: string }) =>
      ingredientService.reassignAndDelete(tenant!, id, targetId),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ingredientKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to reassign ingredient"));
    },
  });
}

export function useAddAlias() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ ingredientId, aliasName }: { ingredientId: string; aliasName: string }) =>
      ingredientService.addAlias(tenant!, ingredientId, aliasName),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: ingredientKeys.detail(tenant, household, variables.ingredientId) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to add alias"));
    },
  });
}

export function useRemoveAlias() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ ingredientId, aliasId }: { ingredientId: string; aliasId: string }) =>
      ingredientService.removeAlias(tenant!, ingredientId, aliasId),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: ingredientKeys.detail(tenant, household, variables.ingredientId) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to remove alias"));
    },
  });
}
