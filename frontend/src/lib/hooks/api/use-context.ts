import { useQuery, useQueryClient } from "@tanstack/react-query";
import { accountService } from "@/services/api/account";

// --- Key factory ---

export const contextKeys = {
  all: ["context"] as const,
  current: () => [...contextKeys.all, "current"] as const,
};

// --- Query hooks ---

export function useAppContext(enabled: boolean = true) {
  return useQuery({
    queryKey: contextKeys.current(),
    queryFn: () => accountService.getContext(),
    enabled,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });
}

// --- Invalidation helper ---

export function useInvalidateContext() {
  const qc = useQueryClient();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: contextKeys.all }),
    invalidateCurrent: () =>
      qc.invalidateQueries({ queryKey: contextKeys.current() }),
  };
}

// --- Prefetch helper ---

export function usePrefetchContext() {
  const qc = useQueryClient();

  return {
    prefetch: () =>
      qc.prefetchQuery({
        queryKey: contextKeys.current(),
        queryFn: () => accountService.getContext(),
        staleTime: 5 * 60 * 1000,
      }),
  };
}
