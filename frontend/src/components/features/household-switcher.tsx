import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAuth } from "@/components/providers/auth-provider";
import { useTenant } from "@/context/tenant-context";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { accountService } from "@/services/api/account";
import { contextKeys } from "@/lib/hooks/api/use-context";
import { taskKeys } from "@/lib/hooks/api/use-tasks";
import { reminderKeys } from "@/lib/hooks/api/use-reminders";
import { getErrorMessage } from "@/lib/api/errors";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export function HouseholdSwitcher() {
  const { appContext } = useAuth();
  const { tenantId, householdId: currentHouseholdId } = useTenant();
  const { data } = useHouseholds(!!appContext);
  const queryClient = useQueryClient();

  const households = data?.data ?? [];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const preferenceId = appContext?.relationships?.preference?.data?.id;

  if (households.length <= 1) return null;

  const handleChange = async (householdId: string | null) => {
    if (!preferenceId || !householdId || householdId === activeId || !tenantId) return;
    try {
      await accountService.setActiveHousehold(tenantId, preferenceId, householdId);
      await queryClient.invalidateQueries({ queryKey: contextKeys.current });
      await queryClient.invalidateQueries({ queryKey: taskKeys.all(currentHouseholdId) });
      await queryClient.invalidateQueries({ queryKey: reminderKeys.all(currentHouseholdId) });
      toast.success("Household switched");
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to switch household"));
    }
  };

  return (
    <Select value={activeId} onValueChange={handleChange}>
      <SelectTrigger className="w-full">
        <SelectValue placeholder="Select household" />
      </SelectTrigger>
      <SelectContent>
        {households.map((hh) => (
          <SelectItem key={hh.id} value={hh.id}>
            {hh.attributes.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
