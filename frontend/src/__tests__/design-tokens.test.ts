import { describe, it, expect } from "vitest";
import { contrastRatio, parseColor, INK, PAPER } from "@/lib/color-contrast";
import { readIndexCss, tokenBlock, resolveToken } from "./css-tokens";

// Task 8 replaces this literal with an import from swatches.ts.
const TRACKER_COLORS = [
  "red", "orange", "amber", "yellow", "lime", "green", "emerald", "teal",
  "cyan", "blue", "indigo", "violet", "purple", "fuchsia", "pink", "rose",
];

const THEME_INVARIANT = ["--radius", "--duration-fast", "--duration-base", "--on-color-ink", "--on-color-paper"];

const TEXT_PAIRS: [fg: string, bg: string][] = [ // >= 4.5 in both themes
  ["--foreground", "--background"], ["--card-foreground", "--card"], ["--popover-foreground", "--popover"],
  ["--primary-foreground", "--primary"], ["--primary", "--background"], ["--secondary-foreground", "--secondary"],
  ["--muted-foreground", "--background"], ["--muted-foreground", "--muted"], ["--muted-foreground", "--card"],
  ["--accent-foreground", "--accent"], ["--destructive-foreground", "--destructive"], ["--destructive", "--background"],
  ["--success-foreground", "--success"], ["--success", "--background"], ["--warning-foreground", "--warning"],
  ["--warning-text", "--background"], ["--warning-text", "--card"], ["--info-foreground", "--info"], ["--info", "--background"],
  ["--sidebar-foreground", "--sidebar"], ["--sidebar-primary-foreground", "--sidebar-primary"], ["--sidebar-accent-foreground", "--sidebar-accent"],
  ["--calendar-busy-foreground", "--calendar-busy"], ["--calendar-package-foreground", "--calendar-package"],
  ...TRACKER_COLORS.map((n): [string, string] => ["--foreground", `--swatch-${n}-soft`]),
];

const NON_TEXT_PAIRS: [fg: string, bg: string][] = [ // >= 3.0 in both themes
  ["--input", "--background"], ["--ring", "--background"], ["--sidebar-ring", "--sidebar"],
  ["--chart-1", "--card"], ["--chart-2", "--card"], ["--chart-3", "--card"], ["--chart-4", "--card"], ["--chart-5", "--card"],
  ["--calendar-busy", "--card"], ["--calendar-package", "--card"],
  ...TRACKER_COLORS.map((n): [string, string] => [`--swatch-${n}`, "--card"]),
];

const css = readIndexCss();
const THEMES = ["light", "dark"] as const;

function ratio(theme: "light" | "dark", fg: string, bg: string): number {
  const f = resolveToken(css, theme, fg);
  const b = resolveToken(css, theme, bg);
  if (!parseColor(f) || !parseColor(b)) throw new Error(`unparseable colour in ${theme} ${fg} (${f}) on ${bg} (${b})`);
  return contrastRatio(f, b);
}

function checkPairs(pairs: [string, string][], threshold: number) {
  const failures: string[] = [];
  for (const theme of THEMES) {
    for (const [fg, bg] of pairs) {
      const r = ratio(theme, fg, bg);
      if (r < threshold) failures.push(`${theme}: ${fg} on ${bg} = ${r.toFixed(2)} (< ${threshold})`);
    }
  }
  return failures;
}

describe("design tokens", () => {
  it("defines the same tokens in :root and .dark (theme-invariant excepted)", () => {
    const keys = (sel: ":root" | ".dark") =>
      [...tokenBlock(css, sel).keys()].filter((k) => !THEME_INVARIANT.includes(k)).sort();
    expect(keys(".dark")).toEqual(keys(":root"));
  });

  it("text pairs meet 4.5:1 in both themes", () => {
    expect(checkPairs(TEXT_PAIRS, 4.5)).toEqual([]);
  });

  it("non-text pairs meet 3:1 in both themes", () => {
    expect(checkPairs(NON_TEXT_PAIRS, 3)).toEqual([]);
  });

  it("on-color ink and paper equal the helper constants", () => {
    expect(resolveToken(css, "light", "--on-color-ink")).toBe(INK);
    expect(resolveToken(css, "light", "--on-color-paper")).toBe(PAPER);
  });

  it("every --color-* in @theme inline points at a token defined in :root", () => {
    const theme = /@theme inline\s*\{([^}]*)\}/.exec(css);
    expect(theme).not.toBeNull();
    const root = tokenBlock(css, ":root");
    const missing: string[] = [];
    for (const m of (theme![1] as string).matchAll(/(--color-[\w-]+)\s*:\s*var\(\s*(--[\w-]+)\s*\)/g)) {
      if (!root.has(m[2] as string)) missing.push(`${m[1]} -> ${m[2]}`);
    }
    expect(missing).toEqual([]);
  });
});

describe("typography", () => {
  const css = readIndexCss();
  const fontFace = (family: string) =>
    new RegExp(`@font-face\\s*\\{[^}]*font-family:\\s*["']${family}["'][^}]*\\}`).exec(css)?.[0] ?? "";

  it("does not import Geist", () => {
    expect(css).not.toMatch(/geist/i);
  });

  it("imports Fraunces and Figtree variable fonts", () => {
    expect(css).toMatch(/@import\s+["']@fontsource-variable\/fraunces/);
    expect(css).toMatch(/@import\s+["']@fontsource-variable\/figtree/);
  });

  it("stacks lead with the variable face then its metric-matched fallback", () => {
    expect(css).toMatch(/--font-heading:\s*'Fraunces Variable',\s*'Fraunces Fallback'/);
    expect(css).toMatch(/--font-sans:\s*'Figtree Variable',\s*'Figtree Fallback'/);
  });

  it.each([
    ["Fraunces Fallback", 'local("Georgia")'],
    ["Figtree Fallback", 'local("Arial")'],
  ])("%s is a local() face with metric overrides", (family, src) => {
    const block = fontFace(family);
    expect(block).not.toBe("");
    expect(block).toContain(src);
    for (const prop of ["size-adjust", "ascent-override", "descent-override"]) {
      expect(block).toContain(`${prop}:`);
    }
  });

  it("sets the text scale in @theme", () => {
    const theme = /@theme\s*\{([^}]*)\}/.exec(css);
    expect(theme).not.toBeNull();
    const body = theme![1] as string;
    expect(body).toMatch(/--text-xl:\s*1\.375rem/);
    expect(body).toMatch(/--text-2xl:\s*1\.75rem/);
    expect(body).toMatch(/--text-3xl:\s*2\.25rem/);
    expect(body).toMatch(/--text-display:\s*3rem/);
  });
});

describe("base layer", () => {
  const css = readIndexCss();
  const squash = (s: string) => s.replace(/\s+/g, " ");

  it("declares the coarse-pointer variant", () => {
    expect(css).toContain("@custom-variant coarse (@media (pointer: coarse));");
  });

  it("drops the 50% alpha outline from the base * rule", () => {
    expect(css).not.toContain("outline-ring/50");
  });

  it("draws a global 2px focus ring with a 2px offset", () => {
    const m = /:focus-visible\s*\{([^}]*)\}/.exec(css);
    expect(m).not.toBeNull();
    const body = squash(m![1] as string);
    expect(body).toContain("outline: 2px solid var(--ring)");
    expect(body).toContain("outline-offset: 2px");
  });

  it("collapses motion under prefers-reduced-motion", () => {
    const m = /@media \(prefers-reduced-motion: reduce\)\s*\{([\s\S]*?\})\s*\}/.exec(css);
    expect(m).not.toBeNull();
    const body = squash(m![1] as string);
    expect(body).toContain("animation-duration: 0.01ms !important");
    expect(body).toContain("transition-duration: 0.01ms !important");
    expect(body).toContain("animation-iteration-count: 1 !important");
  });

  it("sets radius and motion tokens", () => {
    expect(tokenBlock(css, ":root").get("--radius")).toBe("0.75rem");
    expect(tokenBlock(css, ":root").get("--duration-fast")).toBe("120ms");
    expect(tokenBlock(css, ":root").get("--duration-base")).toBe("200ms");
    expect(css).toMatch(/--ease-out:\s*cubic-bezier\(\.2,\s*\.7,\s*\.2,\s*1\)/);
  });

  it("has no elevation in dark and maps raw elevation tokens to shadow utilities", () => {
    expect(tokenBlock(css, ".dark").get("--elevation-card")).toBe("none");
    expect(tokenBlock(css, ".dark").get("--elevation-overlay")).toBe("none");
    expect(css).toContain("--shadow-card: var(--elevation-card)");
    expect(css).toContain("--shadow-overlay: var(--elevation-overlay)");
    expect(tokenBlock(css, ":root").get("--elevation-overlay")).toBe(
      "0 4px 12px oklch(0.3 0.03 50 / 0.10), 0 12px 32px oklch(0.3 0.03 50 / 0.12)",
    );
  });
});
