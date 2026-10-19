import { useState } from "react";
import { toast } from "sonner";
import { ArrowLeft, Check, ChevronRight } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { useTenant } from "@/context/tenant-context";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { useInvalidateTasks } from "@/lib/hooks/api/use-tasks";
import { useInvalidateReminders } from "@/lib/hooks/api/use-reminders";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
import { Dialog, DialogClose, DialogContent, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

interface MobileHouseholdSelectorProps {
  onNavigate?: () => void;
}

export function MobileHouseholdSelector({ onNavigate }: MobileHouseholdSelectorProps) {
  const { appContext } = useAuth();
  const { setActiveHousehold } = useTenant();
  const { data } = useHouseholds(!!appContext);
  const invalidateTasks = useInvalidateTasks();
  const invalidateReminders = useInvalidateReminders();
  const [selectorOpen, setSelectorOpen] = useState(false);

  const households = data?.data ?? [];
  const activeId = appContext?.relationships?.activeHousehold?.data?.id;
  const activeHousehold = households.find((h) => h.id === activeId);

  if (households.length <= 1) return null;

  const handleSelect = async (householdId: string) => {
    if (householdId === activeId) {
      setSelectorOpen(false);
      return;
    }
    try {
      await setActiveHousehold(householdId);
      invalidateTasks.invalidateAll();
      invalidateReminders.invalidateAll();
      toast.success("Household switched");
      setSelectorOpen(false);
      onNavigate?.();
    } catch (error) {
      toast.error(createErrorFromUnknown(error, "Failed to switch household").message);
    }
  };

  return (
    <Dialog open={selectorOpen} onOpenChange={setSelectorOpen}>
      {/* Trigger button in drawer */}
      <DialogTrigger
        aria-haspopup="dialog"
        aria-expanded={selectorOpen}
        className="flex w-full items-center justify-between rounded-md border px-3 py-3 text-sm font-medium transition-colors hover:bg-sidebar-accent/50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring coarse:min-h-11"
      >
        <span className="truncate">{activeHousehold?.attributes.name ?? "Select household"}</span>
        <ChevronRight className="h-4 w-4 text-muted-foreground" />
      </DialogTrigger>

      {/* Full-screen selector */}
      <DialogContent
        showCloseButton={false}
        className="inset-0 top-0 left-0 h-dvh max-h-none w-full max-w-none translate-x-0 translate-y-0 grid-rows-[auto_1fr] content-start gap-0 rounded-none p-0 ring-0 sm:max-w-none"
      >
        {/* Header */}
        <div className="flex h-14 items-center border-b px-4">
          <DialogClose
            render={<Button variant="ghost" size="icon" aria-label="Back" />}
          >
            <ArrowLeft className="h-5 w-5" />
          </DialogClose>
          <DialogTitle className="ml-3 pr-0 text-lg font-semibold">Select Household</DialogTitle>
        </div>

        {/* Household list */}
        <div className="overflow-auto p-3">
          <div className="space-y-1">
            {households.map((household) => {
              const isActive = household.id === activeId;
              return (
                <button
                  key={household.id}
                  type="button"
                  onClick={() => handleSelect(household.id)}
                  aria-current={isActive ? "true" : undefined}
                  className={cn(
                    "flex w-full items-center justify-between rounded-md px-4 py-4 text-left text-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
                    isActive ? "bg-accent text-accent-foreground" : "hover:bg-muted",
                  )}
                >
                  <span>{household.attributes.name}</span>
                  {isActive && <Check aria-hidden="true" className="h-5 w-5" />}
                </button>
              );
            })}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
