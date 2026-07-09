import { cn } from "@/lib/utils";

/**
 * The Home Hub brand mark — a filled house with a lit doorway on an indigo
 * tile. Kept in sync with public/favicon.svg (the browser-tab icon). Sized via
 * className (defaults to size-8); the SVG scales to fit.
 */
export function BrandMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      className={cn("size-8", className)}
      role="img"
      aria-label="Home Hub"
    >
      <defs>
        <linearGradient id="hh-tile" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#6366f1" />
          <stop offset="1" stopColor="#4338ca" />
        </linearGradient>
      </defs>
      <rect width="32" height="32" rx="7.5" fill="url(#hh-tile)" />
      <path d="M16 6.2 27 15.8 5 15.8 Z" fill="#ffffff" />
      <rect x="8.5" y="14.5" width="15" height="11.5" rx="1.6" fill="#ffffff" />
      <path d="M13.6 26 L13.6 21 Q13.6 19 16 19 Q18.4 19 18.4 21 L18.4 26 Z" fill="#f59e0b" />
    </svg>
  );
}
