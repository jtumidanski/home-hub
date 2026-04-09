import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { locationsOfInterestService } from "@/services/api/locations-of-interest";
import { useTenant } from "@/context/tenant-context";
import { getErrorMessage } from "@/lib/api/errors";
import { locationsOfInterestKeys } from "./query-keys";
import { weatherKeys } from "./use-weather";
import type {
  LocationOfInterestCreateAttributes,
  LocationOfInterestUpdateAttributes,
} from "@/types/models/location-of-interest";

export function useLocationsOfInterest() {
  const { tenant, household } = useTenant();
  return useQuery({
    queryKey: locationsOfInterestKeys.list(tenant, household),
    queryFn: () => locationsOfInterestService.list(tenant!),
    enabled: !!tenant?.id && !!household?.id,
  });
}

export function useCreateLocationOfInterest() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (attrs: LocationOfInterestCreateAttributes) =>
      locationsOfInterestService.create(tenant!, attrs),
    onSuccess: (response) => {
      qc.invalidateQueries({
        queryKey: locationsOfInterestKeys.list(tenant, household),
      });
      // The new row was synchronously cache-warmed on the server. Invalidate
      // the weather queries for that location id so the next switch fetches
      // the freshly populated cache row.
      const newId = response.data.id;
      qc.invalidateQueries({
        queryKey: weatherKeys.current(household?.id, newId),
      });
      qc.invalidateQueries({
        queryKey: weatherKeys.forecast(household?.id, newId),
      });
      toast.success("Location saved");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to save location"));
    },
  });
}

export function useUpdateLocationOfInterest() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: ({
      id,
      attrs,
    }: {
      id: string;
      attrs: LocationOfInterestUpdateAttributes;
    }) => locationsOfInterestService.update(tenant!, id, attrs),
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: locationsOfInterestKeys.list(tenant, household),
      });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to update location"));
    },
  });
}

export function useDeleteLocationOfInterest() {
  const qc = useQueryClient();
  const { tenant, household } = useTenant();
  return useMutation({
    mutationFn: (id: string) => locationsOfInterestService.remove(tenant!, id),
    onSuccess: (_void, id) => {
      qc.invalidateQueries({
        queryKey: locationsOfInterestKeys.list(tenant, household),
      });
      qc.invalidateQueries({
        queryKey: weatherKeys.current(household?.id, id),
      });
      qc.invalidateQueries({
        queryKey: weatherKeys.forecast(household?.id, id),
      });
      toast.success("Location removed");
    },
    onError: (error) => {
      toast.error(getErrorMessage(error, "Failed to remove location"));
    },
  });
}
