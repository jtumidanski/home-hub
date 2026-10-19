import { cn } from "@/lib/utils";

/**
 * The Home Hub brand mark — "Window light": a dark house silhouette with a lit
 * window on a terracotta tile. Kept in sync with public/favicon.svg (the
 * browser-tab icon). Literal fills are the documented image-asset exception to
 * the no-colour-literal rule. Sized via className (defaults to size-8); the SVG
 * scales to fit.
 */
export function BrandMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      className={cn("size-8", className)}
      role="img"
      aria-label="Home Hub"
    >
      <rect width="32" height="32" rx="7.5" fill="oklch(0.53 0.13 45)" />
      <path
        d="M16 5 L27.5 14.8 L24.5 14.8 L24.5 26.5 L7.5 26.5 L7.5 14.8 L4.5 14.8 Z"
        fill="oklch(0.24 0.02 55)"
      />
      <rect x="11.5" y="15.5" width="9" height="7.5" rx="1.2" fill="oklch(0.80 0.13 80)" />
      <path
        d="M16 15.5 V23 M11.5 19.25 H20.5"
        stroke="oklch(0.24 0.02 55)"
        strokeWidth="1.3"
      />
    </svg>
  );
}
