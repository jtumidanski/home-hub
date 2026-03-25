import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { accountService } from "@/services/api/account";
import { getErrorMessage } from "@/lib/api/errors";
import type { Tenant } from "@/types/models/tenant";

// --- Key factory ---
// Pattern C (tenant-agnostic): context is the bootstrapping query that
// establishes tenant context, so it is inherently pre-tenant.

export const contextKeys = {
  all: ["context"] as const,
  current: () => [...contextKeys.all, "current"] as const,
};

// --- Query hooks ---

export function useAppContext(enabled: boolean = true) {
  return useQuery({
    queryKey: contextKeys.current(),
    queryFn: () => accountService.getContext({ maxRetries: 1, retryDelay: 500 }),
    enabled,
    retry: false,
    staleTime: 5 * 60 * 1000,
    gcTime: 5 * 60 * 1000,
  });
}

// --- Mutation hooks ---

export function useUpdatePreferenceTheme() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { tenant: Tenant; preferenceId: string; theme: "light" | "dark" }) =>
      accountService.updatePreferenceTheme(args.tenant, args.preferenceId, args.theme),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: contextKeys.current() });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to save theme preference"));
    },
  });
}

export function useCreateTenant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => accountService.createTenant({ name }),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: contextKeys.current() });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create tenant"));
    },
  });
}

export function useOnboardingCreateHousehold() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { tenant: Tenant; name: string; timezone: string; units: "imperial" | "metric" }) =>
      accountService.createHousehold(args.tenant, { name: args.name, timezone: args.timezone, units: args.units }),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: contextKeys.current() });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to create household"));
    },
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
