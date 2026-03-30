import { X, AlertTriangle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { PlanItemAttributes, Slot } from "@/types/models/meal-plan";
import { SLOTS } from "@/types/models/meal-plan";

interface WeekGridProps {
  startsOn: Date;
  items: PlanItemAttributes[];
  locked: boolean;
  onCellClick: (day: string, slot: Slot) => void;
  onItemClick: (item: PlanItemAttributes) => void;
  onRemoveItem: (itemId: string) => void;
}

const SLOT_LABELS: Record<Slot, string> = {
  breakfast: "Breakfast",
  lunch: "Lunch",
  dinner: "Dinner",
  snack: "Snack",
  side: "Side",
};

function getDaysOfWeek(startsOn: Date): { date: Date; label: string; dateStr: string }[] {
  const days = [];
  for (let i = 0; i < 7; i++) {
    const d = new Date(startsOn);
    d.setDate(d.getDate() + i);
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    days.push({
      date: d,
      label: d.toLocaleDateString("en-US", { weekday: "short", month: "short", day: "numeric" }),
      dateStr: `${year}-${month}-${day}`,
    });
  }
  return days;
}

export function WeekGrid({ startsOn, items, locked, onCellClick, onItemClick, onRemoveItem }: WeekGridProps) {
  const days = getDaysOfWeek(startsOn);

  const getItemsForCell = (dateStr: string, slot: Slot) =>
    items
      .filter((item) => item.day === dateStr && item.slot === slot)
      .sort((a, b) => a.position - b.position);

  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr>
            <th className="p-2 text-left font-medium text-muted-foreground w-20">Slot</th>
            {days.map((day) => (
              <th key={day.dateStr} className="p-2 text-center font-medium text-muted-foreground min-w-[140px]">
                {day.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {SLOTS.map((slot) => (
            <tr key={slot} className="border-t">
              <td className="p-2 font-medium text-muted-foreground align-top">
                {SLOT_LABELS[slot]}
              </td>
              {days.map((day) => {
                const cellItems = getItemsForCell(day.dateStr, slot);
                return (
                  <td
                    key={`${day.dateStr}-${slot}`}
                    className={cn(
                      "p-1 align-top border-l min-h-[60px]",
                      locked ? "bg-muted/30" : "hover:bg-accent/50 cursor-pointer"
                    )}
                    onClick={() => {
                      if (!locked) onCellClick(day.dateStr, slot);
                    }}
                  >
                    <div className="space-y-1">
                      {cellItems.map((item) => (
                        <div
                          key={item.id}
                          className="group relative rounded bg-primary/10 p-1.5 text-xs cursor-pointer"
                          onClick={(e) => {
                            e.stopPropagation();
                            if (!locked) onItemClick(item);
                          }}
                        >
                          <div className="flex items-start justify-between gap-1">
                            <span className="font-medium leading-tight">
                              {item.recipe_deleted && (
                                <AlertTriangle className="inline h-3 w-3 text-yellow-500 mr-1" />
                              )}
                              {item.recipe_title}
                            </span>
                            {!locked && (
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-4 w-4 opacity-0 group-hover:opacity-100 shrink-0"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  onRemoveItem(item.id);
                                }}
                              >
                                <X className="h-3 w-3" />
                              </Button>
                            )}
                          </div>
                          {item.recipe_classification && (
                            <Badge variant="secondary" className="mt-0.5 text-[10px] px-1 py-0">
                              {item.recipe_classification}
                            </Badge>
                          )}
                          {(item.planned_servings || (item.serving_multiplier && item.serving_multiplier !== 1)) && (
                            <div className="text-muted-foreground mt-0.5">
                              {item.planned_servings
                                ? `serves ${item.planned_servings}`
                                : `×${item.serving_multiplier}`}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
