import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAuth } from "@/components/providers/auth-provider";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { accountService } from "@/services/api/account";
import { contextKeys } from "@/lib/hooks/api/use-context";
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
  const { data } = useHouseholds(!!appContext);
  const queryClient = useQueryClient();

  const households = data?.data ?? [];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const preferenceId = appContext?.relationships?.preference?.data?.id;

  if (households.length <= 1) return null;

  const handleChange = async (householdId: string | null) => {
    if (!preferenceId || !householdId || householdId === activeId) return;
    try {
      await accountService.setActiveHousehold(preferenceId, householdId);
      await queryClient.invalidateQueries({ queryKey: contextKeys.current });
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
