/// <reference types="node" />
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

export function readIndexCss(): string {
  return readFileSync(resolve(__dirname, "../index.css"), "utf8");
}

/** Declarations of a top-level `:root` or `.dark` block, as "--name" -> raw value. */
export function tokenBlock(css: string, selector: ":root" | ".dark"): Map<string, string> {
  const escaped = selector.replace(/[.]/g, "\\.");
  const m = new RegExp(`(?:^|\\n)${escaped}\\s*\\{([^}]*)\\}`).exec(css);
  const out = new Map<string, string>();
  if (!m) return out;
  for (const decl of (m[1] as string).split(";")) {
    const i = decl.indexOf(":");
    if (i < 0) continue;
    const name = decl.slice(0, i).trim();
    if (name.startsWith("--")) out.set(name, decl.slice(i + 1).trim());
  }
  return out;
}

/** Resolves a token to a literal value, following var(--x) chains. Dark falls back to :root. */
export function resolveToken(css: string, theme: "light" | "dark", name: string): string {
  const root = tokenBlock(css, ":root");
  const dark = theme === "dark" ? tokenBlock(css, ".dark") : new Map<string, string>();
  let current = name;
  for (let depth = 0; depth < 10; depth++) {
    const value = dark.get(current) ?? root.get(current);
    if (value === undefined) throw new Error(`token ${current} is not defined (resolving ${name}, ${theme})`);
    const ref = /^var\(\s*(--[\w-]+)\s*\)$/.exec(value);
    if (!ref) return value;
    current = ref[1] as string;
  }
  throw new Error(`token ${name} has a var() chain deeper than 10`);
}
