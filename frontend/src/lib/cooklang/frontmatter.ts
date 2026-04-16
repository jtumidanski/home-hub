const FM_BLOCK = /^---\s*\n([\s\S]*?)\n---\s*(?:\n|$)/;
const HAS_KEY = (key: string) => new RegExp(`^\\s*${key}\\s*:`, "im");

export function ensureFrontmatter(source: string, title: string, description: string): string {
  const trimmed = source.trimStart();
  const match = FM_BLOCK.exec(trimmed);
  if (match) {
    const block = match[1] ?? "";
    const additions: string[] = [];
    if (!HAS_KEY("title").test(block) && title) additions.push(`title: ${title}`);
    if (!HAS_KEY("description").test(block) && description) additions.push(`description: ${description}`);
    if (additions.length === 0) return source;
    const offset = source.length - trimmed.length;
    const blockStart = offset + 4;
    return source.slice(0, blockStart) + additions.join("\n") + "\n" + source.slice(blockStart);
  }
  if (!title && !description) return source;
  const lines = ["---"];
  if (title) lines.push(`title: ${title}`);
  if (description) lines.push(`description: ${description}`);
  lines.push("---", "");
  return lines.join("\n") + source;
}
