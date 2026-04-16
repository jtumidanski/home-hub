---
description: Convert recipe text into a Cooklang-formatted markdown file, preserving metadata, and verify syntax
argument-hint: Path to a text file containing the recipe, OR the recipe text pasted inline
---

You are converting free-form recipe text into a [Cooklang](https://cooklang.org/docs/spec/) recipe file. The input is: **$ARGUMENTS**

## Cooklang syntax reference

- **Metadata**: Use YAML frontmatter (a `---` fenced block at the very top of the file) — the modern Cooklang form. The older `>> key: value` syntax is deprecated and will emit warnings. Common keys: `title`, `source`, `servings`, `yield`, `time`, `prep time`, `cook time`, `course`, `cuisine`, `tags`, `description`, `difficulty`, `author`. Render `tags` as a single comma-delimited string (e.g. `tags: chicken, salad, quick`), not a YAML list. Render `servings` as a bare number when possible (`servings: 9`) — strings like `"9 servings"` trigger an "Unsupported value" warning from the `cook` CLI.
- **Ingredients**:
  - Single word: `@salt`
  - Multi-word: `@olive oil{}`
  - With quantity: `@flour{500%g}`, `@eggs{2}`, `@onion{1/2}`
- **Cookware**:
  - Single word: `#pot`
  - Multi-word: `#frying pan{}`
- **Timers**: `~{10%minutes}` or named `~eggs{3%minutes}`
- **Comments**: `-- inline comment` or `[- block comment -]`
- **Notes**: Start a line with `> ` to attach a note (background, tips, anecdotes, nutrition info, etc.). Each `> ` line is one note; put a blank line between separate notes. Do NOT represent notes as a `== Notes ==` section with bulleted lists.
- **Steps**: Each step is a paragraph (blank line between steps). No numbered prefixes — paragraph breaks define steps.
- **Sections** (optional): `== Section Name ==`

## Process

### Step 1 — Load input

If `$ARGUMENTS` is a file path that exists, read it. Otherwise treat `$ARGUMENTS` as the raw recipe text. If the text is too short or clearly not a recipe, stop and ask the user for clarification.

### Step 2 — Extract metadata

Scan the input for every piece of metadata you can find and map it to Cooklang `>>` fields. Preserve as much as possible, even non-standard keys (use lowercase keys with spaces, e.g. `>> total time: 45 min`). Capture things like:

- title, author/source/url
- servings/yield, prep/cook/total time
- course, cuisine, tags, difficulty, dietary notes
- description/intro blurb
- any "notes" or "tips" the author includes — emit each one as a Cooklang note (`> one note per line`, blank line between notes), placed after the method. Do not use a `== Notes ==` section or bulleted lists.

Never invent metadata that isn't in the source.

### Step 3 — Parse ingredients and steps

- For each instruction paragraph, identify ingredients used and annotate them with `@name{qty%unit}` syntax, matching the quantities from the ingredients list.
- Identify cookware mentioned and annotate with `#name{}`.
- Identify explicit durations and annotate with `~{qty%unit}`.
- Keep the prose of the steps intact — Cooklang steps should still read like normal instructions, just with inline annotations.
- If the source has an ingredients list that is not all used in the steps (garnishes, optional extras), keep them by referencing them in the appropriate step or in a `== Notes ==` section.

### Step 4 — Determine output path

Create the file under `recipes/` at the repo root (create the directory if missing). Filename: slugified title with `.cook.md` extension (e.g. `recipes/chicken-tikka-masala.cook.md`). If a file already exists at that path, ask the user whether to overwrite.

### Step 5 — Write the file

Structure:

```
---
title: ...
servings: ...
tags: tag1, tag2, tag3
(other metadata)
---

(optional description paragraph or block comment)

== Ingredients == (optional, only if the source had a standalone list worth preserving verbatim)

== Method ==

Step one prose with @ingredient{qty%unit} and #cookware{} and ~{time%unit}.

Step two prose...

> optional note one (tip, anecdote, nutrition info, etc.)

> optional note two
```

### Step 6 — Verify Cooklang syntax

Run verification in this order, stopping at the first one that works:

1. If `cook` CLI is on PATH (`command -v cook`), run `cook recipe read <file>` — it parses and will error on invalid syntax.
2. Else if `npx` is available, try `npx --yes @cooklang/cooklang-ts parse <file>` (non-interactive, will fetch package).
3. Else do a structural lint yourself:
   - Every `@` ingredient with a brace has matching `{...}` and the body is either empty, a number/fraction, or `number%unit`.
   - Every `#` cookware follows the same rule.
   - Every `~` timer has `{qty%unit}`.
   - Metadata lines all start with `>> key: value` and are contiguous at the top.
   - No stray `{` or `}`.
   - Report the exact check you ran.

If verification fails, fix the file and re-verify. Do not report success until verification passes.

### Step 7 — Report

Tell the user:
- Output path
- Which verification method was used and that it passed
- Any metadata from the source you could not map cleanly (so they can decide if it matters)

## Rules

- Do not fabricate ingredients, quantities, times, or steps not present in the source.
- If a quantity is ambiguous (e.g. "a splash of oil"), leave it as `@oil{}` with no quantity rather than guessing.
- Preserve the author's voice in the step prose; only add Cooklang annotations, do not rewrite instructions.
- Do not add commentary or emoji to the output file.
