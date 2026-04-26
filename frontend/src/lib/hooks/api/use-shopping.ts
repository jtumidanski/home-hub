import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { shoppingService } from "@/services/api/shopping";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import { shoppingKeys } from "./query-keys";
import type {
  ShoppingItemCreateAttributes,
  ShoppingItemUpdateAttributes,
} from "@/types/models/shopping";

// --- List Query Hooks ---

export function useShoppingLists(status: "active" | "archived" = "active") {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: shoppingKeys.lists(tenant, household, status),
    queryFn: () => shoppingService.listLists(tenant!, status),
    enabled: !!tenant?.id && !!household?.id,
  });
}

export function useShoppingList(id: string | null) {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: shoppingKeys.detail(tenant, household, id ?? ""),
    queryFn: () => shoppingService.getListDetail(tenant!, id!),
    enabled: !!tenant?.id && !!household?.id && !!id,
  });
}

// --- List Mutation Hooks ---

export function useCreateShoppingList() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (name: string) => shoppingService.createList(tenant!, name),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.lists(tenant, household, "active") });
      toast.success("Shopping list created");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create shopping list"));
    },
  });
}

export function useUpdateShoppingList() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) =>
      shoppingService.updateList(tenant!, id, name),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.all(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update shopping list"));
    },
  });
}

export function useDeleteShoppingList() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => shoppingService.deleteList(tenant!, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.all(tenant, household) });
      toast.success("Shopping list deleted");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to delete shopping list"));
    },
  });
}

export function useArchiveShoppingList() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => shoppingService.archiveList(tenant!, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.all(tenant, household) });
      toast.success("Shopping list archived");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to archive shopping list"));
    },
  });
}

export function useUnarchiveShoppingList() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => shoppingService.unarchiveList(tenant!, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.all(tenant, household) });
      toast.success("Shopping list reactivated");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to unarchive shopping list"));
    },
  });
}

// --- Item Mutation Hooks ---

export function useAddShoppingItem(listId: string) {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: ShoppingItemCreateAttributes) =>
      shoppingService.addItem(tenant!, listId, attrs),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.detail(tenant, household, listId) });
      qc.invalidateQueries({ queryKey: shoppingKeys.lists(tenant, household, "active") });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to add item"));
    },
  });
}

export function useUpdateShoppingItem(listId: string) {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ itemId, attrs }: { itemId: string; attrs: ShoppingItemUpdateAttributes }) =>
      shoppingService.updateItem(tenant!, listId, itemId, attrs),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.detail(tenant, household, listId) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update item"));
    },
  });
}

export function useRemoveShoppingItem(listId: string) {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (itemId: string) => shoppingService.removeItem(tenant!, listId, itemId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.detail(tenant, household, listId) });
      qc.invalidateQueries({ queryKey: shoppingKeys.lists(tenant, household, "active") });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to remove item"));
    },
  });
}

export function useCheckShoppingItem(listId: string) {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ itemId, checked }: { itemId: string; checked: boolean }) =>
      shoppingService.checkItem(tenant!, listId, itemId, { checked }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.detail(tenant, household, listId) });
      qc.invalidateQueries({ queryKey: shoppingKeys.lists(tenant, household, "active") });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update item"));
    },
  });
}

export function useUncheckAllItems(listId: string) {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: () => shoppingService.uncheckAll(tenant!, listId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: shoppingKeys.detail(tenant, household, listId) });
      qc.invalidateQueries({ queryKey: shoppingKeys.lists(tenant, household, "active") });
      toast.success("All items unchecked");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to uncheck items"));
    },
  });
}

export function useImportMealPlan(listId: string) {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (planId: string) => shoppingService.importMealPlan(tenant!, listId, planId),
    onSuccess: (data) => {
      const count = data.data.attributes.imported_count ?? 0;
      qc.invalidateQueries({ queryKey: shoppingKeys.detail(tenant, household, listId) });
      qc.invalidateQueries({ queryKey: shoppingKeys.lists(tenant, household, "active") });
      if (count === 0) {
        toast.success("Meal plan had no ingredients to import");
      } else {
        toast.success(`Added ${count} ${count === 1 ? "item" : "items"} from meal plan`);
      }
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to import from meal plan"));
    },
  });
}
