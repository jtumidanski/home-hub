import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ingredientService } from "@/services/api/ingredient";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import { ingredientKeys, categoryKeys } from "./query-keys";
import type {
  IngredientCategoryCreateAttributes,
  IngredientCategoryUpdateAttributes,
} from "@/types/models/ingredient";

export { categoryKeys } from "./query-keys";

// --- Query hooks ---

export function useIngredientCategories() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: categoryKeys.lists(tenant, household),
    queryFn: () => ingredientService.listCategories(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateCategory() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: IngredientCategoryCreateAttributes) =>
      ingredientService.createCategory(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: categoryKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create category"));
    },
  });
}

export function useUpdateCategory() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: IngredientCategoryUpdateAttributes }) =>
      ingredientService.updateCategory(tenant!, id, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: categoryKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update category"));
    },
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => ingredientService.deleteCategory(tenant!, id),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: categoryKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: ingredientKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to delete category"));
    },
  });
}

export function useBulkCategorize() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ ingredientIds, categoryId }: { ingredientIds: string[]; categoryId: string }) =>
      ingredientService.bulkCategorize(tenant!, ingredientIds, categoryId),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ingredientKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: categoryKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to bulk categorize ingredients"));
    },
  });
}
