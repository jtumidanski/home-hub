import { toast } from "sonner";
import { useAuth } from "@/components/providers/auth-provider";
import { useTenant } from "@/context/tenant-context";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { useInvalidateTasks } from "@/lib/hooks/api/use-tasks";
import { useInvalidateReminders } from "@/lib/hooks/api/use-reminders";
import { createErrorFromUnknown } from "@/lib/api/errors";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export function HouseholdSwitcher() {
  const { appContext } = useAuth();
  const { setActiveHousehold } = useTenant();
  const { data } = useHouseholds(!!appContext);
  const invalidateTasks = useInvalidateTasks();
  const invalidateReminders = useInvalidateReminders();

  const households = data?.data ?? [];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;

  if (households.length <= 1) return null;

  const handleChange = async (newHouseholdId: string | null) => {
    if (!newHouseholdId || newHouseholdId === activeId) return;
    try {
      await setActiveHousehold(newHouseholdId);
      invalidateTasks.invalidateAll();
      invalidateReminders.invalidateAll();
      toast.success("Household switched");
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to switch household").message);
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
