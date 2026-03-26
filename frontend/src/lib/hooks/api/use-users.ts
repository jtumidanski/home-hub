import { useQuery } from "@tanstack/react-query";
import { authService } from "@/services/api/auth";

// --- Key factory ---

export const userKeys = {
  all: ["users"] as const,
  byIds: (ids: string[]) =>
    [...userKeys.all, "byIds", ...ids.sort()] as const,
};

// --- Query hooks ---

export function useUsersByIds(ids: string[]) {
  return useQuery({
    queryKey: userKeys.byIds(ids),
    queryFn: () => authService.getUsersByIds(ids),
    enabled: ids.length > 0,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
}
