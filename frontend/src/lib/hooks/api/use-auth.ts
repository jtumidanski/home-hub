import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { authService } from "@/services/api/auth";

// --- Key factory ---

export const authKeys = {
  all: ["auth"] as const,
  me: () => [...authKeys.all, "me"] as const,
  providers: () => [...authKeys.all, "providers"] as const,
};

// --- Query hooks ---

export function useMe() {
  return useQuery({
    queryKey: authKeys.me(),
    queryFn: () => authService.getMe(),
    retry: false,
    staleTime: 5 * 60 * 1000,
  });
}

export function useProviders() {
  return useQuery({
    queryKey: authKeys.providers(),
    queryFn: () => authService.getProviders(),
    staleTime: 10 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useLogout() {
  return useMutation({
    mutationFn: () => authService.logout(),
    onSettled: () => {
      window.location.href = "/login";
    },
  });
}

// --- Invalidation helper ---

export function useInvalidateAuth() {
  const qc = useQueryClient();

  return {
    invalidateAll: () =>
      qc.invalidateQueries({ queryKey: authKeys.all }),
    invalidateMe: () =>
      qc.invalidateQueries({ queryKey: authKeys.me() }),
    invalidateProviders: () =>
      qc.invalidateQueries({ queryKey: authKeys.providers() }),
  };
}

// --- Prefetch helper ---

export function usePrefetchAuth() {
  const qc = useQueryClient();

  return {
    prefetchMe: () =>
      qc.prefetchQuery({
        queryKey: authKeys.me(),
        queryFn: () => authService.getMe(),
        staleTime: 5 * 60 * 1000,
      }),
    prefetchProviders: () =>
      qc.prefetchQuery({
        queryKey: authKeys.providers(),
        queryFn: () => authService.getProviders(),
        staleTime: 10 * 60 * 1000,
      }),
  };
}
