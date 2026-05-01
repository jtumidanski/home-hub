import { useCallback, useSyncExternalStore } from "react";
import { getLocalDateStrOffset } from "@/lib/date-utils";

const POLL_MS = 60_000;

/**
 * Returns the calendar date as YYYY-MM-DD `offsetDays` after today in the
 * given IANA timezone, polling every 60s to catch midnight transitions.
 * Mirrors `useLocalDate` but with an offset.
 */
export function useLocalDateOffset(tz: string | undefined, offsetDays: number): string {
  const subscribe = useCallback(
    (notify: () => void) => {
      const id = window.setInterval(notify, POLL_MS);
      return () => window.clearInterval(id);
    },
    [],
  );
  const getSnapshot = useCallback(
    () => getLocalDateStrOffset(tz, offsetDays),
    [tz, offsetDays],
  );
  return useSyncExternalStore(subscribe, getSnapshot);
}
