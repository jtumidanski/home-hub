import { useQuery } from "@tanstack/react-query";
import { accountService } from "@/services/api/account";

export const contextKeys = {
  current: ["context", "current"] as const,
};

export function useAppContext(enabled: boolean = true) {
  return useQuery({
    queryKey: contextKeys.current,
    queryFn: () => accountService.getContext(),
    enabled,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });
}
