import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { authService } from "@/services/api/auth";
import { userKeys } from "@/lib/hooks/api/use-users";
import type { UserUpdateAttributes } from "@/types/models/user";

// --- Key factory ---
// Pattern C (tenant-agnostic): auth endpoints are not tenant-scoped.
// They return the same user/providers regardless of tenant context.

export const authKeys = {
  all: ["auth"] as const,
  me: () => [...authKeys.all, "me"] as const,
  providers: () => [...authKeys.all, "providers"] as const,
};

// --- Query hooks ---

export function useMe() {
  return useQuery({
    queryKey: authKeys.me(),
    queryFn: () => authService.getMe({ maxRetries: 1, retryDelay: 500 }),
    retry: false,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

export function useProviders() {
  return useQuery({
    queryKey: authKeys.providers(),
    queryFn: () => authService.getProviders(),
    staleTime: 10 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useUpdateMe() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (attributes: UserUpdateAttributes) =>
      authService.updateMe(attributes),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: authKeys.me() });
      qc.invalidateQueries({ queryKey: userKeys.all });
    },
  });
}

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
