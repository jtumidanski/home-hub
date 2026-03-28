import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { recipeService } from "@/services/api/recipe";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import type { RecipeCreateAttributes, RecipeUpdateAttributes } from "@/types/models/recipe";
import type { Tenant } from "@/types/models/tenant";
import type { Household } from "@/types/models/household";

// --- Key factory ---

export const recipeKeys = {
  all: (tenant: Tenant | null, household: Household | null) =>
    ["recipes", tenant?.id ?? "no-tenant", household?.id ?? "no-household"] as const,
  lists: (tenant: Tenant | null, household: Household | null) =>
    [...recipeKeys.all(tenant, household), "list"] as const,
  details: (tenant: Tenant | null, household: Household | null) =>
    [...recipeKeys.all(tenant, household), "detail"] as const,
  detail: (tenant: Tenant | null, household: Household | null, id: string) =>
    [...recipeKeys.details(tenant, household), id] as const,
  tags: (tenant: Tenant | null, household: Household | null) =>
    [...recipeKeys.all(tenant, household), "tags"] as const,
};

// --- Query hooks ---

interface UseRecipesParams {
  search?: string | undefined;
  tags?: string[] | undefined;
  page?: number | undefined;
  pageSize?: number | undefined;
  plannerReady?: boolean | undefined;
  classification?: string | undefined;
  normalizationStatus?: string | undefined;
}

export function useRecipes(params?: UseRecipesParams) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: [...recipeKeys.lists(tenant, household), params] as const,
    queryFn: () => recipeService.listRecipes(tenant!, params),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useRecipe(id: string) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: recipeKeys.detail(tenant, household, id),
    queryFn: () => recipeService.getRecipe(tenant!, id),
    enabled: !!tenant?.id && !!household?.id && !!id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useRecipeTags() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: recipeKeys.tags(tenant, household),
    queryFn: () => recipeService.listTags(tenant!),
    enabled: !!tenant?.id && !!household?.id,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useCreateRecipe() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: RecipeCreateAttributes) =>
      recipeService.createRecipe(tenant!, attrs),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: recipeKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: recipeKeys.tags(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create recipe"));
    },
  });
}

export function useUpdateRecipe() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: RecipeUpdateAttributes }) =>
      recipeService.updateRecipe(tenant!, id, attrs),
    onSettled: (_data, _err, variables) => {
      qc.invalidateQueries({ queryKey: recipeKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: recipeKeys.detail(tenant, household, variables.id) });
      qc.invalidateQueries({ queryKey: recipeKeys.tags(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update recipe"));
    },
  });
}

export function useDeleteRecipe() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => recipeService.deleteRecipe(tenant!, id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: recipeKeys.lists(tenant, household) });
      const previous = qc.getQueryData(recipeKeys.lists(tenant, household));
      if (previous) {
        qc.setQueryData(recipeKeys.lists(tenant, household), {
          ...(previous as Record<string, unknown>),
          data: ((previous as { data: Array<{ id: string }> }).data ?? []).filter(
            (item) => item.id !== id,
          ),
        });
      }
      return { previous };
    },
    onError: (error, _id, context) => {
      if (context?.previous) {
        qc.setQueryData(recipeKeys.lists(tenant, household), context.previous);
      }
      toast.error(getErrorMessage(error, "Failed to delete recipe"));
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: recipeKeys.lists(tenant, household) });
      qc.invalidateQueries({ queryKey: recipeKeys.tags(tenant, household) });
    },
  });
}

// --- Parse hook for live preview ---

export function useParseRecipe() {
  const { tenant } = useTenant();
  return useMutation({
    mutationFn: (source: string) => recipeService.parseSource(tenant!, source),
  });
}

// --- Invalidation helper ---

export function useInvalidateRecipes() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: recipeKeys.all(tenant, household) }),
    invalidateLists: () =>
      qc.invalidateQueries({ queryKey: recipeKeys.lists(tenant, household) }),
    invalidateRecipe: (id: string) =>
      qc.invalidateQueries({ queryKey: recipeKeys.detail(tenant, household, id) }),
  };
}
