import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { wishListService } from "@/services/api/wish-list";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import { wishListKeys } from "./query-keys";
import type {
  WishListItem,
  WishListItemCreateAttributes,
  WishListItemUpdateAttributes,
} from "@/types/models/wish-list";
import type { ApiListResponse } from "@/types/api/responses";

export function useWishListItems() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: wishListKeys.items(tenant, household),
    queryFn: () => wishListService.list(tenant!),
    enabled: !!tenant?.id && !!household?.id,
  });
}

export function useCreateWishListItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: WishListItemCreateAttributes) =>
      wishListService.create(tenant!, attrs),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: wishListKeys.items(tenant, household) });
      toast.success("Wish list item added");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to add wish list item"));
    },
  });
}

export function useUpdateWishListItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({ id, attrs }: { id: string; attrs: WishListItemUpdateAttributes }) =>
      wishListService.update(tenant!, id, attrs),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: wishListKeys.items(tenant, household) });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update wish list item"));
    },
  });
}

export function useDeleteWishListItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => wishListService.remove(tenant!, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: wishListKeys.items(tenant, household) });
      toast.success("Wish list item deleted");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to delete wish list item"));
    },
  });
}

export function useVoteWishListItem() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  const queryKey = wishListKeys.items(tenant, household);
  return useMutation({
    mutationFn: (id: string) => wishListService.vote(tenant!, id),
    onMutate: async (id: string) => {
      await qc.cancelQueries({ queryKey });
      const previous = qc.getQueryData<ApiListResponse<WishListItem>>(queryKey);
      if (previous) {
        const updated: ApiListResponse<WishListItem> = {
          ...previous,
          data: previous.data.map((item) =>
            item.id === id
              ? {
                  ...item,
                  attributes: {
                    ...item.attributes,
                    vote_count: item.attributes.vote_count + 1,
                  },
                }
              : item,
          ),
        };
        qc.setQueryData(queryKey, updated);
      }
      return { previous };
    },
    onError: (error, _id, context) => {
      if (context?.previous) {
        qc.setQueryData(queryKey, context.previous);
      }
      toast.error(getErrorMessage(error, "Failed to vote"));
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey });
    },
  });
}
