import { useEffect, useRef } from 'react';

/**
 * Hook to prevent screen from sleeping on kiosk displays
 * Uses Screen Wake Lock API if available, falls back to keep-alive pings
 */
export function useScreenWakeLock() {
  const wakeLockRef = useRef<WakeLockSentinel | null>(null);

  useEffect(() => {
    let intervalId: NodeJS.Timeout | null = null;

    const requestWakeLock = async () => {
      try {
        // Try to use Screen Wake Lock API (modern browsers)
        if ('wakeLock' in navigator) {
          wakeLockRef.current = await navigator.wakeLock.request('screen');
          console.log('[ScreenWakeLock] Wake lock active');

          // Re-request wake lock if it's released (e.g., tab becomes hidden)
          wakeLockRef.current.addEventListener('release', () => {
            console.log('[ScreenWakeLock] Wake lock released');
          });
        } else {
          console.log('[ScreenWakeLock] Wake Lock API not supported, using fallback');
          // Fallback: periodic activity to prevent sleep
          intervalId = setInterval(() => {
            // Small no-op activity to prevent sleep
            document.body.style.transform = 'translateZ(0)';
            setTimeout(() => {
              document.body.style.transform = '';
            }, 100);
          }, 30000); // Every 30 seconds
        }
      } catch (err) {
        console.error('[ScreenWakeLock] Failed to acquire wake lock:', err);
      }
    };

    // Request wake lock on mount
    requestWakeLock();

    // Re-request wake lock when page becomes visible
    const handleVisibilityChange = () => {
      if (!document.hidden && wakeLockRef.current !== null) {
        requestWakeLock();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);

    // Cleanup
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);

      if (wakeLockRef.current) {
        wakeLockRef.current.release();
        wakeLockRef.current = null;
      }

      if (intervalId) {
        clearInterval(intervalId);
      }
    };
  }, []);
}
