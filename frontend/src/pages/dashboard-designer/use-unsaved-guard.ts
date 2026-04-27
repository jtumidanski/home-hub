import { useEffect } from "react";

/**
 * Browser-level guard: while `dirty` is true, install a `beforeunload`
 * listener so the browser surfaces its native "leave this page?" prompt on
 * tab-close / reload / address-bar navigation. The listener is removed the
 * moment `dirty` returns to false.
 *
 * The React Router v7 `useBlocker` hook would be the natural way to guard
 * in-app navigations, but it requires a data router — this app uses the
 * standard `BrowserRouter`. The designer page instead triggers its own
 * discard-confirmation dialog from the Discard button and any explicit
 * navigate() call site.
 */
export function useUnsavedGuard(dirty: boolean): void {
  useEffect(() => {
    if (!dirty) return;
    const handler = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      // Required by legacy Chrome to show the browser-default prompt.
      event.returnValue = "";
      return "";
    };
    window.addEventListener("beforeunload", handler);
    return () => {
      window.removeEventListener("beforeunload", handler);
    };
  }, [dirty]);
}
