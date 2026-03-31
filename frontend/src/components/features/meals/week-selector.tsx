import { useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

interface WeekSelectorProps {
  startsOn: Date;
  onWeekChange: (newStartsOn: Date) => void;
}

function getMonday(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  d.setDate(d.getDate() + diff);
  d.setHours(0, 0, 0, 0);
  return d;
}

export function WeekSelector({ startsOn, onWeekChange }: WeekSelectorProps) {
  const [open, setOpen] = useState(false);

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

  const handleCalendarSelect = (date: Date | undefined) => {
    if (!date) return;
    onWeekChange(getMonday(date));
    setOpen(false);
  };

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="icon" onClick={goToPreviousWeek}>
        <ChevronLeft className="h-4 w-4" />
      </Button>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger
          className="text-sm font-medium min-w-[180px] text-center hover:underline cursor-pointer"
        >
          {formatDate(startsOn)} – {formatDate(endDate)}, {endDate.getFullYear()}
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="center">
          <Calendar
            mode="single"
            selected={startsOn}
            onSelect={handleCalendarSelect}
            defaultMonth={startsOn}
          />
        </PopoverContent>
      </Popover>
      <Button variant="outline" size="icon" onClick={goToNextWeek}>
        <ChevronRight className="h-4 w-4" />
      </Button>
    </div>
  );
}
