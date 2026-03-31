import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";

interface WeekSelectorProps {
  startsOn: Date;
  onWeekChange: (newStartsOn: Date) => void;
}

export function WeekSelector({ startsOn, onWeekChange }: WeekSelectorProps) {
  const endDate = new Date(startsOn);
  endDate.setDate(endDate.getDate() + 6);

  const formatDate = (d: Date) =>
    d.toLocaleDateString("en-US", { month: "short", day: "numeric" });

  const goToPreviousWeek = () => {
    const prev = new Date(startsOn);
    prev.setDate(prev.getDate() - 7);
    onWeekChange(prev);
  };

  const goToNextWeek = () => {
    const next = new Date(startsOn);
    next.setDate(next.getDate() + 7);
    onWeekChange(next);
  };

  const goToCurrentWeek = () => {
    const today = new Date();
    const day = today.getDay();
    const diff = day === 0 ? -6 : 1 - day; // Monday start
    const monday = new Date(today);
    monday.setDate(today.getDate() + diff);
    monday.setHours(0, 0, 0, 0);
    onWeekChange(monday);
  };

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="icon" onClick={goToPreviousWeek}>
        <ChevronLeft className="h-4 w-4" />
      </Button>
      <button
        onClick={goToCurrentWeek}
        className="text-sm font-medium min-w-[180px] text-center hover:underline cursor-pointer"
      >
        {formatDate(startsOn)} – {formatDate(endDate)}, {endDate.getFullYear()}
      </button>
      <Button variant="outline" size="icon" onClick={goToNextWeek}>
        <ChevronRight className="h-4 w-4" />
      </Button>
    </div>
  );
}
