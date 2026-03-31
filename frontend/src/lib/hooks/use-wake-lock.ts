import { useState, useEffect, useRef } from "react";

export function useWakeLock(enabled: boolean) {
  const [isActive, setIsActive] = useState(false);
  const sentinelRef = useRef<WakeLockSentinel | null>(null);
  const isSupported = typeof navigator !== "undefined" && "wakeLock" in navigator;

  /* eslint-disable react-hooks/set-state-in-effect -- wake lock lifecycle requires synchronizing active state with external API */
  useEffect(() => {
    if (!enabled || !isSupported) {
      // Release if we had one
      const sentinel = sentinelRef.current;
      if (sentinel) {
        sentinel.release().then(() => {
          sentinelRef.current = null;
        });
      }
      setIsActive(false);
      return;
    }

    let cancelled = false;

    async function acquire() {
      try {
        const sentinel = await navigator.wakeLock.request("screen");
        if (cancelled) {
          await sentinel.release();
          return;
        }
        sentinelRef.current = sentinel;
        setIsActive(true);
        sentinel.addEventListener("release", () => {
          if (!cancelled) setIsActive(false);
        });
      } catch {
        if (!cancelled) setIsActive(false);
      }
    }

    acquire();

    // Re-acquire on tab refocus
    const onVisibilityChange = () => {
      if (document.visibilityState === "visible" && !cancelled) {
        acquire();
      }
    };
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      cancelled = true;
      document.removeEventListener("visibilitychange", onVisibilityChange);
      const sentinel = sentinelRef.current;
      if (sentinel) {
        sentinel.release();
        sentinelRef.current = null;
      }
    };
  }, [enabled, isSupported]);
  /* eslint-enable react-hooks/set-state-in-effect */

  return { isSupported, isActive };
}
