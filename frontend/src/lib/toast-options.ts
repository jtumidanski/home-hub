import type { CSSProperties } from "react";

// Sonner injects unlayered CSS that outranks Tailwind's layered utilities, so
// the base surface is themed through Sonner's own CSS variables (token refs).
export const toasterStyle = {
  "--normal-bg": "var(--popover)",
  "--normal-text": "var(--popover-foreground)",
  "--normal-border": "var(--border)",
  "--border-radius": "var(--radius)",
} as CSSProperties;

// Per-type borders and the close button need `!` to beat Sonner's unlayered rules.
export const toastOptions = {
  classNames: {
    toast: "shadow-overlay!",
    success: "border-success!",
    error: "border-destructive!",
    warning: "border-warning!",
    info: "border-info!",
    closeButton: "size-6! coarse:size-11!",
  },
};
