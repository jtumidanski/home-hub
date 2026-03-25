import { useState } from "react";
import { createPortal } from "react-dom";
import { toast } from "sonner";
import { ArrowLeft, Check, ChevronRight } from "lucide-react";
import { useAuth } from "@/components/providers/auth-provider";
import { useTenant } from "@/context/tenant-context";
import { useHouseholds } from "@/lib/hooks/api/use-households";
import { useInvalidateTasks } from "@/lib/hooks/api/use-tasks";
import { useInvalidateReminders } from "@/lib/hooks/api/use-reminders";
import { createErrorFromUnknown } from "@/lib/api/errors";
import { Button } from "@/components/ui/button";
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
    <>
      {/* Trigger button in drawer */}
      <button
        type="button"
        onClick={() => setSelectorOpen(true)}
        className="flex w-full items-center justify-between rounded-md border px-3 py-3 text-sm font-medium transition-colors hover:bg-sidebar-accent/50"
      >
        <span className="truncate">{activeHousehold?.attributes.name ?? "Select household"}</span>
        <ChevronRight className="h-4 w-4 text-muted-foreground" />
      </button>

      {/* Full-screen selector overlay */}
      {selectorOpen &&
        createPortal(
          <div className="fixed inset-0 z-[60] flex flex-col bg-background">
            {/* Header */}
            <div className="flex h-14 items-center border-b px-4">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setSelectorOpen(false)}
                aria-label="Back"
              >
                <ArrowLeft className="h-5 w-5" />
              </Button>
              <span className="ml-3 text-lg font-semibold">Select Household</span>
            </div>

            {/* Household list */}
            <div className="flex-1 overflow-auto p-3">
              <div className="space-y-1">
                {households.map((household) => (
                  <button
                    key={household.id}
                    type="button"
                    onClick={() => handleSelect(household.id)}
                    className={cn(
                      "flex w-full items-center justify-between rounded-md px-4 py-4 text-left text-sm font-medium transition-colors",
                      household.id === activeId
                        ? "bg-primary/10 text-primary"
                        : "hover:bg-muted",
                    )}
                  >
                    <span>{household.attributes.name}</span>
                    {household.id === activeId && <Check className="h-5 w-5" />}
                  </button>
                ))}
              </div>
            </div>
          </div>,
          document.body,
        )}
    </>
  );
}
