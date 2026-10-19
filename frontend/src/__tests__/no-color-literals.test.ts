/// <reference types="node" />
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative, resolve, sep } from "node:path";
import { describe, expect, it } from "vitest";

const SRC = resolve(__dirname, "..");
const SCANNED_ROOTS = ["components/features", "pages"];

/** Path prefixes (relative to frontend/src) not yet migrated; files under them are skipped. */
const PENDING: string[] = [
  "components/features/households/",
  "components/features/packages/",
  "components/features/weather/",
  "components/features/workout/",
  "components/features/workouts/",
  "pages/DataRetentionPage.tsx",
  "pages/HouseholdMembersPage.tsx",
  "pages/HouseholdsPage.tsx",
  "pages/PackagesPage.tsx",
  "pages/SettingsPage.tsx",
  "pages/WeatherPage.tsx",
  "pages/WorkoutExercisesPage.tsx",
  "pages/WorkoutTaxonomyPage.tsx",
  "pages/WorkoutWeekPage.tsx",
];

/** Permitted literals: exact file + exact matched text; reason mirrors DESIGN.md. */
const ALLOWED: { file: string; match: string; reason: string }[] = [];

const PALETTE =
  "red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|slate|gray|zinc|neutral|stone";
const PROPS =
  "bg|text|border(?:-[trblxy])?|ring-offset|ring|fill|stroke|from|via|to|outline|decoration|divide|shadow|accent|caret|placeholder";
const PALETTE_CLASS = new RegExp(
  `(?<![\\w-])(?:[\\w-]+:)*(?:${PROPS})-(?:(?:${PALETTE})-\\d{2,3}|white|black)(?:/\\d+)?(?![\\w-])`,
  "g",
);
const FUNCTIONAL = /\b(?:rgba?|hsla?|oklch)\(/g;
const HEX = /#[0-9a-fA-F]{3,8}\b/g;
const STRING_LITERAL = /"(?:[^"\\\n]|\\.)*"|'(?:[^'\\\n]|\\.)*'|`(?:[^`\\]|\\.)*`/g;

/** Returns every colour literal found in a single line of source. */
export function findColorLiterals(line: string): string[] {
  const found: string[] = [];
  for (const s of line.match(STRING_LITERAL) ?? []) {
    found.push(...(s.match(HEX) ?? []));
  }
  found.push(...(line.match(FUNCTIONAL) ?? []));
  found.push(...(line.match(PALETTE_CLASS) ?? []));
  return found;
}

function walk(dir: string): string[] {
  if (!existsSync(dir)) return [];
  const out: string[] = [];
  for (const name of readdirSync(dir)) {
    const full = join(dir, name);
    if (statSync(full).isDirectory()) out.push(...walk(full));
    else if (/\.tsx?$/.test(name)) out.push(full);
  }
  return out;
}

function rel(full: string): string {
  return relative(SRC, full).split(sep).join("/");
}

function scannedFiles(): string[] {
  return SCANNED_ROOTS.flatMap((r) => walk(join(SRC, r)))
    .map(rel)
    .filter((p) => !p.includes("/__tests__/"));
}

const isPending = (p: string) => PENDING.some((prefix) => p.startsWith(prefix));

function matches(file: string): { line: number; match: string }[] {
  const lines = readFileSync(join(SRC, file), "utf8").split("\n");
  return lines.flatMap((text, i) => findColorLiterals(text).map((match) => ({ line: i + 1, match })));
}

describe("no colour literals in migrated frontend directories", () => {
  it("has no unallowed colour literal in any non-pending file", () => {
    const violations: string[] = [];
    for (const file of scannedFiles().filter((f) => !isPending(f))) {
      for (const { line, match } of matches(file)) {
        if (ALLOWED.some((a) => a.file === file && a.match === match)) continue;
        violations.push(`${file}:${line}: ${match}`);
      }
    }
    expect(violations, violations.join("\n")).toEqual([]);
  });

  it("has every PENDING entry present on disk", () => {
    const missing = PENDING.filter((p) => !existsSync(join(SRC, p)));
    expect(missing).toEqual([]);
  });

  it("has no stale ALLOWED entries", () => {
    const stale = ALLOWED.filter(
      (a) => !existsSync(join(SRC, a.file)) || !matches(a.file).some((m) => m.match === a.match),
    );
    expect(stale).toEqual([]);
  });

  it("matcher flags literals and ignores token-backed classes", () => {
    for (const bad of ['"bg-red-500"', '"dark:text-white"', '"bg-black/30"', '"#9ca3af"']) {
      expect(findColorLiterals(bad), bad).not.toEqual([]);
    }
    for (const good of ['"bg-primary"', '"text-muted-foreground"', '"bg-swatch-red"', '"border-l-swatch-blue"']) {
      expect(findColorLiterals(good), good).toEqual([]);
    }
    expect(findColorLiterals("// issue #123456 in comment")).toEqual([]);
  });
});
