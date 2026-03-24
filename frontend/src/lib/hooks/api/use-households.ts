import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import type { JsonApiListResponse } from "@/types/api/responses";
import type { Household } from "@/types/models/household";

export const householdKeys = {
  list: ["households"] as const,
};

export function useHouseholds(enabled: boolean = true) {
  return useQuery({
    queryKey: householdKeys.list,
    queryFn: () => api.get<JsonApiListResponse<Household>>("/households"),
    enabled,
    staleTime: 5 * 60 * 1000,
  });
}
