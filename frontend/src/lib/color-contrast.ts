/** Gamma-encoded sRGB, each channel clamped to 0..1. */
export interface Rgb {
  r: number;
  g: number;
  b: number;
}

export const INK = "oklch(0.24 0.02 55)";
export const PAPER = "oklch(0.985 0.008 85)";

const HEX = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i;
const NUM = "[+-]?(?:\\d+\\.?\\d*|\\.\\d+)";
const OKLCH = new RegExp(
  `^oklch\\(\\s*(${NUM})(%?)\\s+(${NUM})\\s+(${NUM})(?:deg)?\\s*(?:/\\s*(?:${NUM})%?\\s*)?\\)$`,
  "i",
);

function clamp01(x: number): number {
  return Math.min(1, Math.max(0, x));
}

function encode(linear: number): number {
  const x = clamp01(linear);
  return x <= 0.0031308 ? 12.92 * x : 1.055 * Math.pow(x, 1 / 2.4) - 0.055;
}

function linearize(c: number): number {
  return c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
}

function parseHex(input: string): Rgb | null {
  const m = HEX.exec(input);
  if (!m) return null;
  let h = m[1] as string;
  if (h.length === 3) h = h.split("").map((c) => c + c).join("");
  return {
    r: parseInt(h.slice(0, 2), 16) / 255,
    g: parseInt(h.slice(2, 4), 16) / 255,
    b: parseInt(h.slice(4, 6), 16) / 255,
  };
}

function parseOklch(input: string): Rgb | null {
  const m = OKLCH.exec(input.trim());
  if (!m) return null;
  const L = m[2] === "%" ? parseFloat(m[1] as string) / 100 : parseFloat(m[1] as string);
  const C = parseFloat(m[3] as string);
  const H = (parseFloat(m[4] as string) * Math.PI) / 180;
  const a = C * Math.cos(H);
  const b = C * Math.sin(H);

  const l = Math.pow(L + 0.3963377774 * a + 0.2158037573 * b, 3);
  const mm = Math.pow(L - 0.1055613458 * a - 0.0638541728 * b, 3);
  const s = Math.pow(L - 0.0894841775 * a - 1.291485548 * b, 3);

  return {
    r: encode(4.0767416621 * l - 3.3077115913 * mm + 0.2309699292 * s),
    g: encode(-1.2684380046 * l + 2.6097574011 * mm - 0.3413193965 * s),
    b: encode(-0.0041960863 * l - 0.7034186147 * mm + 1.707614701 * s),
  };
}

/** Parses "#rgb", "#rrggbb", and "oklch(L C H [/ a])". Alpha is ignored. Anything else returns null. */
export function parseColor(input: string): Rgb | null {
  return parseHex(input) ?? parseOklch(input);
}

/** WCAG 2.x relative luminance. */
export function relativeLuminance(c: Rgb): number {
  return 0.2126 * linearize(c.r) + 0.7152 * linearize(c.g) + 0.0722 * linearize(c.b);
}

function luminanceOf(input: string): number {
  const c = parseColor(input);
  if (!c) throw new Error(`unparseable colour: ${input}`);
  return relativeLuminance(c);
}

export function contrastRatio(a: string, b: string): number {
  const la = luminanceOf(a);
  const lb = luminanceOf(b);
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}

/** Whichever of INK / PAPER has the higher contrast against bg. Unparseable bg falls back to "ink". */
export function readableForeground(bg: string): "ink" | "paper" {
  if (!parseColor(bg)) return "ink";
  return contrastRatio(INK, bg) >= contrastRatio(PAPER, bg) ? "ink" : "paper";
}
