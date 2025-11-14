import { useEffect, useRef, useCallback } from 'react';

interface UsePollingOptions {
  interval?: number;
  enabled?: boolean;
  pauseOnHidden?: boolean;
}

/**
 * Hook for polling a callback function at a regular interval
 *
 * @param callback - Function to execute on each poll
 * @param options - Configuration options
 * @param options.interval - Polling interval in milliseconds (default: 20000ms / 20s)
 * @param options.enabled - Whether polling is enabled (default: true)
 * @param options.pauseOnHidden - Pause polling when tab is hidden (default: true)
 */
export function usePolling(
  callback: () => void | Promise<void>,
  options: UsePollingOptions = {}
) {
  const {
    interval = 20000,
    enabled = true,
    pauseOnHidden = true,
  } = options;

  const savedCallback = useRef(callback);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const isVisible = useRef(true);

  // Update saved callback on each render
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  // Handle visibility change
  useEffect(() => {
    if (!pauseOnHidden) return;

    const handleVisibilityChange = () => {
      isVisible.current = !document.hidden;

      // Clear interval when hidden
      if (document.hidden && intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }

      // Restart interval when visible
      if (!document.hidden && enabled) {
        // Execute callback immediately when becoming visible
        savedCallback.current();

        intervalRef.current = setInterval(() => {
          savedCallback.current();
        }, interval);
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [pauseOnHidden, enabled, interval]);

  // Main polling effect
  useEffect(() => {
    if (!enabled) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

    // Don't start if hidden and pause is enabled
    if (pauseOnHidden && document.hidden) {
      return;
    }

    // Execute callback immediately on mount/enable
    savedCallback.current();

    // Set up interval
    intervalRef.current = setInterval(() => {
      savedCallback.current();
    }, interval);

    // Cleanup
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [interval, enabled, pauseOnHidden]);
}
