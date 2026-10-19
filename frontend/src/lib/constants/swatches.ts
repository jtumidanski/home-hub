import type { TrackerColor } from "@/types/models/tracker";

// `satisfies Record<TrackerColor, true>` makes TypeScript reject a missing or misspelled colour name.
const COLOR_KEYS = {
  red: true, orange: true, amber: true, yellow: true,
  lime: true, green: true, emerald: true, teal: true,
  cyan: true, blue: true, indigo: true, violet: true,
  purple: true, fuchsia: true, pink: true, rose: true,
} satisfies Record<TrackerColor, true>;

export const TRACKER_COLORS: readonly TrackerColor[] = Object.keys(COLOR_KEYS) as TrackerColor[];

export type SwatchRole = "solid" | "soft" | "borderLeft" | "border";

// Tailwind's scanner only sees whole literals, so every class is spelled out here.
export const SWATCH_CLASSES: Record<SwatchRole, Record<TrackerColor, string>> = {
  solid: {
    red: "bg-swatch-red", orange: "bg-swatch-orange", amber: "bg-swatch-amber", yellow: "bg-swatch-yellow",
    lime: "bg-swatch-lime", green: "bg-swatch-green", emerald: "bg-swatch-emerald", teal: "bg-swatch-teal",
    cyan: "bg-swatch-cyan", blue: "bg-swatch-blue", indigo: "bg-swatch-indigo", violet: "bg-swatch-violet",
    purple: "bg-swatch-purple", fuchsia: "bg-swatch-fuchsia", pink: "bg-swatch-pink", rose: "bg-swatch-rose",
  },
  soft: {
    red: "bg-swatch-red-soft", orange: "bg-swatch-orange-soft", amber: "bg-swatch-amber-soft", yellow: "bg-swatch-yellow-soft",
    lime: "bg-swatch-lime-soft", green: "bg-swatch-green-soft", emerald: "bg-swatch-emerald-soft", teal: "bg-swatch-teal-soft",
    cyan: "bg-swatch-cyan-soft", blue: "bg-swatch-blue-soft", indigo: "bg-swatch-indigo-soft", violet: "bg-swatch-violet-soft",
    purple: "bg-swatch-purple-soft", fuchsia: "bg-swatch-fuchsia-soft", pink: "bg-swatch-pink-soft", rose: "bg-swatch-rose-soft",
  },
  borderLeft: {
    red: "border-l-swatch-red", orange: "border-l-swatch-orange", amber: "border-l-swatch-amber", yellow: "border-l-swatch-yellow",
    lime: "border-l-swatch-lime", green: "border-l-swatch-green", emerald: "border-l-swatch-emerald", teal: "border-l-swatch-teal",
    cyan: "border-l-swatch-cyan", blue: "border-l-swatch-blue", indigo: "border-l-swatch-indigo", violet: "border-l-swatch-violet",
    purple: "border-l-swatch-purple", fuchsia: "border-l-swatch-fuchsia", pink: "border-l-swatch-pink", rose: "border-l-swatch-rose",
  },
  border: {
    red: "border-swatch-red", orange: "border-swatch-orange", amber: "border-swatch-amber", yellow: "border-swatch-yellow",
    lime: "border-swatch-lime", green: "border-swatch-green", emerald: "border-swatch-emerald", teal: "border-swatch-teal",
    cyan: "border-swatch-cyan", blue: "border-swatch-blue", indigo: "border-swatch-indigo", violet: "border-swatch-violet",
    purple: "border-swatch-purple", fuchsia: "border-swatch-fuchsia", pink: "border-swatch-pink", rose: "border-swatch-rose",
  },
};

const NEUTRAL_FALLBACK: Record<SwatchRole, string> = {
  solid: "bg-muted-foreground",
  soft: "bg-muted",
  borderLeft: "border-l-muted-foreground",
  border: "border-muted-foreground",
};

// Unknown colours (including prototype keys such as "constructor") get the neutral fallback.
export function swatchClass(color: string, role: SwatchRole): string {
  const map = SWATCH_CLASSES[role];
  if (!Object.prototype.hasOwnProperty.call(map, color)) return NEUTRAL_FALLBACK[role];
  return map[color as TrackerColor];
}
