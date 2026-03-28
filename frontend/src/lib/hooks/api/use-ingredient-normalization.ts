import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { recipeService } from "@/services/api/recipe";
import { useTenant } from "@/context/tenant-context";
import { recipeKeys } from "@/lib/hooks/api/use-recipes";
import { ingredientKeys } from "@/lib/hooks/api/use-ingredients";
import { getErrorMessage } from "@/lib/api/errors";

export function useResolveIngredient() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({
      recipeId,
      ingredientId,
      canonicalIngredientId,
      saveAsAlias,
    }: {
      recipeId: string;
      ingredientId: string;
      canonicalIngredientId: string;
      saveAsAlias: boolean;
    }) =>
      recipeService.resolveIngredient(tenant!, recipeId, ingredientId, canonicalIngredientId, saveAsAlias),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: recipeKeys.detail(tenant, household, variables.recipeId) });
      qc.invalidateQueries({ queryKey: recipeKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: ingredientKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to resolve ingredient"));
    },
  });
}

export function useRenormalize() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (recipeId: string) => recipeService.renormalize(tenant!, recipeId),
    onSuccess: () => {
      toast.success("Ingredients re-normalized");
    },
    onSettled: (_data, _err, recipeId) => {
      qc.invalidateQueries({ queryKey: recipeKeys.detail(tenant, household, recipeId) });
      qc.invalidateQueries({ queryKey: recipeKeys.lists(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to re-normalize"));
    },
  });
}
