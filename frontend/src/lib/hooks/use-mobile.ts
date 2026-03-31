import { useSyncExternalStore } from "react";

const MOBILE_BREAKPOINT = 767;

function subscribe(callback: () => void) {
  const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT}px)`);
  mql.addEventListener("change", callback);
  return () => mql.removeEventListener("change", callback);
}

function getSnapshot() {
  return window.innerWidth <= MOBILE_BREAKPOINT;
}

function getServerSnapshot() {
  return false;
}

export function useMobile(): boolean {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
