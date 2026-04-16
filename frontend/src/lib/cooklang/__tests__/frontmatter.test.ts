import { describe, it, expect } from "vitest";
import { ensureFrontmatter } from "../frontmatter";

describe("ensureFrontmatter", () => {
  it("prepends frontmatter when none exists", () => {
    const result = ensureFrontmatter("Cook @rice{200%g}.", "Fried Rice", "Quick weeknight meal");
    expect(result).toBe("---\ntitle: Fried Rice\ndescription: Quick weeknight meal\n---\nCook @rice{200%g}.");
  });

  it("injects missing title into existing frontmatter", () => {
    const source = "---\ntags: dinner\n---\n\nCook @rice{200%g}.";
    const result = ensureFrontmatter(source, "Fried Rice", "");
    expect(result).toContain("title: Fried Rice");
    expect(result).toContain("tags: dinner");
  });

  it("injects missing description into existing frontmatter", () => {
    const source = "---\ntitle: Fried Rice\n---\n\nCook @rice{200%g}.";
    const result = ensureFrontmatter(source, "Fried Rice", "Quick meal");
    expect(result).toContain("description: Quick meal");
    expect(result).toContain("title: Fried Rice");
  });

  it("does nothing when both already present", () => {
    const source = "---\ntitle: Fried Rice\ndescription: Quick meal\n---\n\nCook @rice{200%g}.";
    const result = ensureFrontmatter(source, "Fried Rice", "Quick meal");
    expect(result).toBe(source);
  });

  it("returns source unchanged when no title or description provided", () => {
    const source = "Cook @rice{200%g}.";
    const result = ensureFrontmatter(source, "", "");
    expect(result).toBe(source);
  });
});
