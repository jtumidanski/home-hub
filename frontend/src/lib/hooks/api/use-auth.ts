import { useQuery } from "@tanstack/react-query";
import { authService } from "@/services/api/auth";

export const authKeys = {
  me: ["auth", "me"] as const,
  providers: ["auth", "providers"] as const,
};

export function useMe() {
  return useQuery({
    queryKey: authKeys.me,
    queryFn: () => authService.getMe(),
    retry: false,
    staleTime: 5 * 60 * 1000,
  });
}

export function useProviders() {
  return useQuery({
    queryKey: authKeys.providers,
    queryFn: () => authService.getProviders(),
    staleTime: 10 * 60 * 1000,
  });
}
