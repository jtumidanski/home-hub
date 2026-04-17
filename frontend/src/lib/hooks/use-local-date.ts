import { useCallback, useSyncExternalStore } from "react";
import { getLocalTodayStr } from "@/lib/date-utils";

const POLL_MS = 60_000;

/**
 * Returns today's calendar date as `YYYY-MM-DD` in the given IANA timezone
 * (or browser default if `tz` is omitted) and re-renders only when the local
 * date string actually changes. Polls every 60 seconds — enough to catch the
 * midnight transition for a tab left open across days.
 */
export function useLocalDate(tz?: string): string {
  const subscribe = useCallback(
    (notify: () => void) => {
      const id = window.setInterval(notify, POLL_MS);
      return () => window.clearInterval(id);
    },
    [],
  );
  const getSnapshot = useCallback(() => getLocalTodayStr(tz), [tz]);
  return useSyncExternalStore(subscribe, getSnapshot);
}
