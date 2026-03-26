import { Loader2 } from "lucide-react";
import type { CalendarSource } from "@/types/models/calendar";
import { useToggleCalendarSource } from "@/lib/hooks/api/use-calendar";

interface CalendarSelectionPanelProps {
  connectionId: string;
  sources: CalendarSource[];
}

export function CalendarSelectionPanel({ connectionId, sources }: CalendarSelectionPanelProps) {
  const toggle = useToggleCalendarSource();

  if (sources.length === 0) {
    return <p className="text-sm text-muted-foreground">No calendars found.</p>;
  }

  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium">Calendars</h4>
      {sources.map((source) => (
        <label
          key={source.id}
          className="flex items-center gap-2 text-sm cursor-pointer hover:bg-muted/50 rounded px-2 py-1"
        >
          <input
            type="checkbox"
            checked={source.attributes.visible}
            onChange={() =>
              toggle.mutate({
                connectionId,
                calId: source.id,
                visible: !source.attributes.visible,
              })
            }
            disabled={toggle.isPending}
            className="rounded"
          />
          <div
            className="w-3 h-3 rounded-full flex-shrink-0"
            style={{ backgroundColor: source.attributes.color || "#888" }}
          />
          <span>{source.attributes.name}</span>
          {source.attributes.primary && (
            <span className="text-xs text-muted-foreground">(primary)</span>
          )}
          {toggle.isPending && <Loader2 className="h-3 w-3 animate-spin ml-auto" />}
        </label>
      ))}
    </div>
  );
}
