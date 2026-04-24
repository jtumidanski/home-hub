import { CalendarWidget } from "@/components/features/calendar/calendar-widget";

export interface CalendarAdapterConfig {
  horizonDays: 1 | 3 | 7;
  includeAllDay: boolean;
}

/**
 * Registry adapter around CalendarWidget. horizonDays / includeAllDay
 * are accepted on the config and forwarded — not yet wired through the
 * underlying query range; that's a follow-up.
 */
export function CalendarAdapter({ config }: { config: CalendarAdapterConfig }) {
  return (
    <CalendarWidget
      horizonDays={config.horizonDays}
      includeAllDay={config.includeAllDay}
    />
  );
}
