import { readableForeground } from "@/lib/color-contrast";

/** Sentinel userColor for package ETA events; resolves to the calendar-package token pair. */
export const PACKAGE_EVENT_COLOR = "var(--calendar-package)";

export interface ChipColors {
  backgroundColor: string;
  color: string;
}

/**
 * Background + readable text colour for an event chip. Busy placeholders and
 * package events use token pairs; user-coloured events pick ink or paper by contrast.
 */
export function eventChipColors(attrs: { title: string; isOwner: boolean; userColor: string }): ChipColors {
  if (attrs.title === "Busy" && !attrs.isOwner) {
    return { backgroundColor: "var(--calendar-busy)", color: "var(--calendar-busy-foreground)" };
  }
  if (attrs.userColor === PACKAGE_EVENT_COLOR) {
    return { backgroundColor: PACKAGE_EVENT_COLOR, color: "var(--calendar-package-foreground)" };
  }
  return {
    backgroundColor: attrs.userColor,
    color: readableForeground(attrs.userColor) === "ink" ? "var(--on-color-ink)" : "var(--on-color-paper)",
  };
}
