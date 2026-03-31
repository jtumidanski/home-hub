import { useState, useEffect, useRef, useCallback } from "react";

export function useWakeLock(enabled: boolean) {
  const [isActive, setIsActive] = useState(false);
  const sentinelRef = useRef<WakeLockSentinel | null>(null);
  const isSupported = typeof navigator !== "undefined" && "wakeLock" in navigator;

  const acquire = useCallback(async () => {
    if (!isSupported || !enabled) return;
    try {
      sentinelRef.current = await navigator.wakeLock.request("screen");
      setIsActive(true);
      sentinelRef.current.addEventListener("release", () => {
        setIsActive(false);
      });
    } catch {
      setIsActive(false);
    }
  }, [isSupported, enabled]);

  const release = useCallback(async () => {
    if (sentinelRef.current) {
      await sentinelRef.current.release();
      sentinelRef.current = null;
      setIsActive(false);
    }
  }, []);

  // Acquire/release based on enabled flag
  useEffect(() => {
    if (enabled) {
      acquire();
    } else {
      release();
    }
    return () => {
      release();
    };
  }, [enabled, acquire, release]);

  // Re-acquire on visibility change (tab refocus)
  useEffect(() => {
    if (!enabled || !isSupported) return;
    const handler = () => {
      if (document.visibilityState === "visible") {
        acquire();
      }
    };
    document.addEventListener("visibilitychange", handler);
    return () => document.removeEventListener("visibilitychange", handler);
  }, [enabled, isSupported, acquire]);

  return { isSupported, isActive };
}
