import type { CalendarConnection } from "@/types/models/calendar";

interface UserLegendProps {
  connections: CalendarConnection[];
}

export function UserLegend({ connections }: UserLegendProps) {
  if (connections.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-3 text-sm">
      {connections.map((conn) => (
        <div key={conn.id} className="flex items-center gap-1.5">
          <div
            className="w-3 h-3 rounded-full flex-shrink-0"
            style={{ backgroundColor: conn.attributes.userColor }}
          />
          <span className="text-muted-foreground">{conn.attributes.userDisplayName}</span>
        </div>
      ))}
    </div>
  );
}
