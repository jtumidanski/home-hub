import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export type CalendarViewMode = "week" | "month";

interface ViewModeToggleProps {
  mode: CalendarViewMode;
  onChange: (mode: CalendarViewMode) => void;
}

export function ViewModeToggle({ mode, onChange }: ViewModeToggleProps) {
  return (
    <div className="flex items-center border rounded-md" role="group" aria-label="Calendar view mode">
      <Button
        variant="ghost"
        size="sm"
        aria-pressed={mode === "week"}
        className={cn(mode === "week" && "bg-accent text-accent-foreground")}
        onClick={() => onChange("week")}
      >
        Week
      </Button>
      <Button
        variant="ghost"
        size="sm"
        aria-pressed={mode === "month"}
        className={cn(mode === "month" && "bg-accent text-accent-foreground")}
        onClick={() => onChange("month")}
      >
        Month
      </Button>
    </div>
  );
}
