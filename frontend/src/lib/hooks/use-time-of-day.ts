import { useCallback, useSyncExternalStore } from "react";

const POLL_MS = 60_000;

export type TimeOfDay = "morning" | "afternoon" | "evening";

/**
 * Returns the current time-of-day bucket in the given IANA timezone (or
 * browser default) and re-renders only when the bucket changes. Polls every
 * 60 seconds so an always-on display moves from morning to afternoon to
 * evening without a reload.
 */
export function useTimeOfDay(tz?: string): TimeOfDay {
  const subscribe = useCallback((notify: () => void) => {
    const id = window.setInterval(notify, POLL_MS);
    return () => window.clearInterval(id);
  }, []);
  const getSnapshot = useCallback((): TimeOfDay => {
    const hour = Number(
      new Intl.DateTimeFormat("en-US", { hour: "numeric", hourCycle: "h23", timeZone: tz }).format(new Date()),
    );
    return hour < 12 ? "morning" : hour < 18 ? "afternoon" : "evening";
  }, [tz]);
  return useSyncExternalStore(subscribe, getSnapshot);
}
